package main

type Config struct {
	// The web app address.
	Address string `split_words:"true" default:":8080"`

	// Namespace of the sealed secrets controller (e.g. `kube-system`).
	SealedSecretsControllerNamespace string `split_words:"true" required:"true"`
	// Name of sealed secrets controller (e.g. `sealed-secrets`).
	SealedSecretsControllerName string `split_words:"true" required:"true"`

	// Options: debug, info, warn, error, dpanic, panic, and fatal.
	LogLevel string `split_words:"true" default:"info"`

	// Set during build / compile time.
	Version string `split_words:"true" default:"unknown"`
}
