package appx

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	// registry holds all the registered applications.
	registry = make(map[string]*App)

	// installed holds all the installed applications, in dependency order.
	installed []*App

	// lifecycle holds the Start and Stop callbacks of the runnable applications.
	lifecycle = new(lifecycleImpl)
)

// Register registers the application app into the registry.
func Register(app *App) error {
	if app == nil {
		return fmt.Errorf("nil app %v", app)
	}

	if app.Name == "" {
		return fmt.Errorf("the name of app %v is empty", app)
	}

	if _, ok := registry[app.Name]; ok {
		return fmt.Errorf("app %q is already registered", app.Name)
	}

	registry[app.Name] = app
	app.getAppFunc = getApp // Find an application in the registry.
	return nil
}

// MustRegister is like Register but panics if there is an error.
func MustRegister(app *App) {
	if err := Register(app); err != nil {
		panic(err)
	}
}

// Install installs the applications specified by names, with the given ctx.
// If no name is specified, all registered applications will be installed.
//
// Note that applications will be installed in dependency order.
func Install(ctx context.Context, names ...string) error {
	after := func(app *App) {
		installed = append(installed, app)
	}

	if len(names) == 0 {
		for _, app := range registry {
			if err := app.Install(ctx, lifecycle, after); err != nil {
				// Install failed, roll back.
				Uninstall()
				return err
			}
		}
	}

	for _, name := range names {
		app, err := getApp(name)
		if err != nil {
			// Install failed, roll back.
			Uninstall()
			return err
		}
		if err := app.Install(ctx, lifecycle, after); err != nil {
			// Install failed, roll back.
			Uninstall()
			return err
		}
	}

	return nil
}

// Uninstall uninstalls the applications that has already been installed, in
// the reverse order of installation.
func Uninstall() {
	for i := len(installed); i > 0; i-- {
		if err := installed[i-1].Uninstall(); err != nil {
			config.errorHandler()(err)
		}
	}
}

func getApp(name string) (*App, error) {
	app, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("app %q is not registered", name)
	}

	return app, nil
}

// Run starts all long-running applications, blocks on the signal channel,
// and then gracefully stops the applications. It is designed as a shortcut
// of calling Start and Stop for typical usage scenarios.
//
// The default timeout for application startup and shutdown is 15s, which can
// be changed by using SetConfig.
func Run() error {
	startCtx, cancel := context.WithTimeout(context.Background(), config.startTimeout())
	defer cancel()
	if err := Start(startCtx); err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c

	stopCtx, cancel := context.WithTimeout(context.Background(), config.stopTimeout())
	defer cancel()
	Stop(stopCtx)

	return nil
}

// Start kicks off all long-running applications, like network servers or
// message queue consumers. It will returns immediately if it encounters an error.
func Start(ctx context.Context) error {
	return withTimeout(ctx, start)
}

// Stop gracefully stops all long-running applications. For best-effort cleanup,
// It will keep going after encountering errors, and all errors will be passed
// to the handler specified by ErrorHandler.
func Stop(ctx context.Context) {
	withTimeout(ctx, stop)
}

func start(ctx context.Context) error {
	if err := lifecycle.Start(ctx); err != nil {
		// Start failed, roll back.
		stop(ctx)
		return err
	}
	return nil
}

func stop(ctx context.Context) error {
	for _, err := range lifecycle.Stop(ctx) {
		config.errorHandler()(err)
	}
	return nil
}

func withTimeout(ctx context.Context, f func(context.Context) error) error {
	c := make(chan error, 1)
	go func() { c <- f(ctx) }()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}
