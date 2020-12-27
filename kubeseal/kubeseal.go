package kubeseal

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

type KubesealClient struct {
	ControllerNamespace string
	ControllerName      string
}

func (c *KubesealClient) Seal(secretJSON string) ([]byte, error) {
	args := []string{
		"--format=json",
		fmt.Sprintf("--controller-namespace=%s", c.ControllerNamespace),
		fmt.Sprintf("--controller-name=%s", c.ControllerName),
	}
	cmd := exec.Command("kubeseal", args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialise stdin pipe for kubeseal")
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, secretJSON)
	}()

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, errors.Wrapf(err, "failed to run kubeseal, stderr: %s", stderr.String())
		}
		return nil, errors.Wrap(err, "failed to run kubeseal")
	}

	return stdout.Bytes(), nil
}

func (c *KubesealClient) SealRaw(name, raw string) ([]byte, error) {
	args := []string{
		"--raw",
		"--from-file=/dev/stdin",
		"--format=json",
		fmt.Sprintf("--controller-namespace=%s", c.ControllerNamespace),
		fmt.Sprintf("--controller-name=%s", c.ControllerName),
		fmt.Sprintf("--name=%s", name),
	}
	cmd := exec.Command("kubeseal", args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialise stdin pipe for kubeseal")
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, raw)
	}()

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return nil, errors.Wrapf(err, "failed to run kubeseal, stderr: %s", stderr.String())
		}
		return nil, errors.Wrap(err, "failed to run kubeseal")
	}

	return stdout.Bytes(), nil
}
