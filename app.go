package appx

import (
	"context"
	"fmt"
)

const (
	stateInstalling = iota + 1
	stateInstalled
	stateUninstalled
)

type Initializer interface {
	// Init initializes an application with the given context ctx.
	// It will return an error if it fails.
	Init(ctx Context) error
}

type Cleaner interface {
	// Clean does the cleanup work for an application. It will return
	// an error if it fails.
	Clean() error
}

type StartStopper interface {
	// Start kicks off a long-running application, like network servers or
	// message queue consumers. It will return an error if it fails.
	Start(ctx context.Context) error

	// Stop gracefully stops a long-running application. It will return an
	// error if it fails.
	Stop(ctx context.Context) error
}

type Instancer interface {
	Instance() interface{}
}

type Validator interface {
	Validate() error
}

// Context is a set of context parameters used to initialize the
// associated application.
type Context struct {
	context.Context

	App      *App // The application associated with this context.
	required map[string]*App
}

// Load loads the application instance specified by name. It will return
// an error if the given name does not refer to any required application.
func (ctx Context) Load(name string) (interface{}, error) {
	if app, ok := ctx.required[name]; ok {
		if instancer, ok := app.instance.(Instancer); ok {
			return instancer.Instance(), nil
		}
		return app.instance, nil
	}
	return nil, fmt.Errorf("app %q is not required", name)
}

// MustLoad is like Load but panics if there is an error.
func (ctx Context) MustLoad(name string) interface{} {
	instance, err := ctx.Load(name)
	if err != nil {
		panic(err)
	}
	return instance
}

// Config returns the configuration of the application associated with this
// context. It will return nil if there is no configuration.
//
// Note that in the current implementation, in order to get the configuration
// successfully, Context.Config() must be called after AppConfigs has already
// been set by calling Registry.SetOptions().
func (ctx Context) Config() interface{} {
	return ctx.App.getConfigFunc(ctx.App.Name)
}

// App is a modular application.
type App struct {
	Name     string      // The application name.
	instance interface{} // The user-defined application instance.

	requiredNames map[string]bool
	requiredApps  map[string]*App

	// The function used to find an application by its name.
	getAppFunc func(name string) (*App, error)

	// The function used to get the application's configuration by its name.
	getConfigFunc func(name string) interface{}

	state int // The installation state.
}

// New creates an application with the given name and the user-defined instance.
func New(name string, instance interface{}) *App {
	a := &App{
		Name:          name,
		requiredNames: make(map[string]bool),
		requiredApps:  make(map[string]*App),
		getAppFunc: func(name string) (*App, error) {
			return nil, fmt.Errorf("app %q is not registered", name)
		},
		instance: instance,
	}
	return a
}

// Require sets the names of the applications that the current application requires.
func (a *App) Require(names ...string) *App {
	for _, name := range names {
		a.requiredNames[name] = true
	}
	return a
}

// Install does the initialization work for the current application.
func (a *App) Install(ctx context.Context, lc Lifecycle, after func(*App)) (err error) {
	switch a.state {
	case stateInstalled:
		return nil // Do nothing since the application has already been installed.
	case stateInstalling:
		return fmt.Errorf("circular dependency is detected for app %q", a.Name)
	}

	// Mark the state as `installing`.
	a.state = stateInstalling

	// Install all the required applications.
	if err := a.prepareRequiredApps(); err != nil {
		return err
	}
	for _, app := range a.requiredApps {
		if err = app.Install(ctx, lc, after); err != nil {
			return err
		}
	}

	//////////////////////////////////
	// Finally install the app itself.

	// 1. If a.instance implements Initializer, initialize the app instance.
	appCtx := Context{
		Context:  ctx,
		App:      a,
		required: a.requiredApps,
	}
	if initializer, ok := a.instance.(Initializer); ok {
		if err := initializer.Init(appCtx); err != nil {
			return err
		}
	}

	// 2. If a.instance implements StartStopper, set the appropriate
	// lifecycle hooks.
	if startStopper, ok := a.instance.(StartStopper); ok {
		lc.Append(Hook{
			OnStart: startStopper.Start,
			OnStop:  startStopper.Stop,
		})
	}

	// 3. If a.instance implements Validator, trigger the validation.
	if validator, ok := a.instance.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	if after != nil {
		// Call the hook function after installed, if any.
		after(a)
	}

	a.state = stateInstalled
	return nil
}

// Uninstall does the cleanup work for the current application.
func (a *App) Uninstall() (err error) {
	if a.state == stateUninstalled {
		return nil
	}

	// If a.instance implements Cleaner, cleanup the app instance.
	if cleaner, ok := a.instance.(Cleaner); ok {
		if err = cleaner.Clean(); err != nil {
			return err
		}
	}

	a.state = stateUninstalled
	return nil
}

// prepareRequiredApps sets the field a.requiredApps of app if it's not set.
func (a *App) prepareRequiredApps() error {
	if len(a.requiredNames) == len(a.requiredApps) {
		return nil
	}

	for name := range a.requiredNames {
		app, err := a.getAppFunc(name)
		if err != nil {
			return err
		}
		a.requiredApps[name] = app
	}

	return nil
}
