package appx

import (
	"context"
	"os"
)

var (
	globalRegistry = NewRegistry()
)

func Register(app *App) error {
	return globalRegistry.Register(app)
}

func MustRegister(app *App) {
	globalRegistry.MustRegister(app)
}

func Install(ctx context.Context, names ...string) error {
	return globalRegistry.Install(ctx, names...)
}

func Uninstall() {
	globalRegistry.Uninstall()
}

func Run() (os.Signal, error) {
	return globalRegistry.Run()
}

func Start(ctx context.Context) error {
	return globalRegistry.Start(ctx)
}

func Stop(ctx context.Context) {
	globalRegistry.Stop(ctx)
}

// SetConfig sets the configuration of globalRegistry.
func SetConfig(c Config) {
	if c.StartTimeout > 0 {
		globalRegistry.options.StartTimeout = c.StartTimeout
	}
	if c.StopTimeout > 0 {
		globalRegistry.options.StopTimeout = c.StopTimeout
	}
	if c.ErrorHandler != nil {
		globalRegistry.options.ErrorHandler = c.ErrorHandler
	}
	if c.AppConfigs != nil {
		globalRegistry.options.AppConfigs = c.AppConfigs
	}
}
