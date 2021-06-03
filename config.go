package appx

import (
	"context"
	"time"
)

var (
	// config is the Config singleton.
	config = Config{}
)

type Unmarshaller func(context.Context, string, interface{}) error

// Config is the appx configuration.
type Config struct {
	// The timeout of application startup. Defaults to 15s.
	StartTimeout time.Duration

	// The timeout of application shutdown. Defaults to 15s.
	StopTimeout time.Duration

	// The handler for errors during the Stop and Uninstall phases.
	ErrorHandler func(error)

	// The unmarshaller that unmarshals an application's configuration
	// into its instance.
	Unmarshaller Unmarshaller
}

func (c Config) startTimeout() time.Duration {
	if c.StartTimeout == 0 {
		return 15 * time.Second
	}
	return c.StartTimeout
}

func (c Config) stopTimeout() time.Duration {
	if c.StopTimeout == 0 {
		return 15 * time.Second
	}
	return c.StopTimeout
}

func (c Config) errorHandler() func(error) {
	if c.ErrorHandler == nil {
		return func(error) {}
	}
	return c.ErrorHandler
}

func (c Config) unmarshaller() Unmarshaller {
	if c.Unmarshaller == nil {
		return func(context.Context, string, interface{}) error { return nil }
	}
	return c.Unmarshaller
}

// SetConfig sets the appx configuration.
func SetConfig(c Config) {
	config = c
}
