package appx

import (
	"context"
	"os"
)

var (
	globalRegistry = NewRegistry()
)

func Use(middlewares ...func(Standard) Standard) {
	globalRegistry.Use(middlewares...)
}

func Register(app *App) error {
	return globalRegistry.Register(app)
}

func MustRegister(app *App) {
	globalRegistry.MustRegister(app)
}

func SetOptions(opts *Options) {
	globalRegistry.SetOptions(opts)
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

func Graph() map[string][]string {
	return globalRegistry.Graph()
}
