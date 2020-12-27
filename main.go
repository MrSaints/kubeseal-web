package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/kelseyhightower/envconfig"
	"github.com/markbates/pkger"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/mrsaints/kubeseal-web/kubeseal"
)

type SealRequest struct {
	SecretJSON string `json:"secretJson" binding:"required"`
}

type SealResponse struct {
	SealedSecretJSON string `json:"sealedSecretJson"`
}

func main() {
	var c Config
	err := envconfig.Process("ksweb", &c)
	if err != nil {
		panic(errors.Wrap(err, "failed to process configuration"))
	}

	if c.Version == "" {
		c.Version = "unknown"
	}

	logConfig := zap.NewProductionConfig()
	err = logConfig.Level.UnmarshalText([]byte(c.LogLevel))
	if err != nil {
		panic(errors.Wrapf(err, "failed to determine log-level: %s", c.LogLevel))
	}
	commonLogFields := zap.Fields(
		zap.String("version", c.Version),
	)
	logger, err := logConfig.Build(commonLogFields)
	if err != nil {
		panic(errors.Wrap(err, "failed to set-up logging"))
	}

	// Test `kubeseal` works before starting the service
	ks := &kubeseal.KubesealClient{
		ControllerNamespace: c.SealedSecretsControllerNamespace,
		ControllerName:      c.SealedSecretsControllerName,
	}
	_, err = ks.SealRaw("healthcheck", time.Now().UTC().String())
	if err != nil {
		logger.Panic("`kubeseal` is not configured correctly", zap.Error(err))
	}

	r := gin.New()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.StaticFS("/", pkger.Dir("/static"))

	r.POST("/seal", func(c *gin.Context) {
		var req SealRequest
		err := c.ShouldBindJSON(&req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "empty request body"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sealedSecret, err := ks.Seal(req.SecretJSON)
		if err != nil {
			if strings.Contains(err.Error(), "json parse error") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse Kubernetes secret JSON"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res := &SealResponse{
			SealedSecretJSON: string(sealedSecret),
		}

		c.JSON(http.StatusOK, res)
	})

	srv := &http.Server{
		Addr:    c.Address,
		Handler: r,
	}

	go func() {
		logger.Info("Starting service", zap.Any("config", c))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Panic("Failed to start service", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Stopping service")
	_ = logger.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Panic("Failed to stop service", zap.Error(err))
	}
}
