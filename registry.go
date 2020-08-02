package appx

import (
	"context"
	"fmt"
)

var (
	// registry holds all the registered applications.
	registry = make(map[string]*App)

	// lifecycle holds the Start and Stop callbacks of the runnable applications.
	lifecycle = new(lifecycleImpl)

	// errorHandler is the handler for errors from the Stop and Uninstall phases.
	errorHandler = func(error) {}
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
func Install(ctx context.Context, names ...string) error {
	if len(names) == 0 {
		for _, app := range registry {
			if err := app.Install(ctx, lifecycle); err != nil {
				return err
			}
		}
	}

	for _, name := range names {
		app, err := getApp(name)
		if err != nil {
			return err
		}
		if err := app.Install(ctx, lifecycle); err != nil {
			return err
		}
	}

	return nil
}

// Uninstall uninstalls the applications specified by names.
// If no name is specified, all registered applications will be uninstalled.
func Uninstall(names ...string) error {
	if len(names) == 0 {
		for _, app := range registry {
			if err := app.Uninstall(); err != nil {
				return err
			}
		}
	}

	for _, name := range names {
		app, err := getApp(name)
		if err != nil {
			return err
		}
		if err := app.Uninstall(); err != nil {
			return err
		}
	}

	return nil
}

func getApp(name string) (*App, error) {
	app, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("app %q is not registered", name)
	}

	return app, nil
}

// ErrorHandler sets the handler for errors from Stop and Uninstall phases.
func ErrorHandler(f func(error)) {
	if f != nil {
		errorHandler = f
	}
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
		errorHandler(err)
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
