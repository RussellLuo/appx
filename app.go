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

// InitFunc initializes an application with the given context ctx, lifecycle lc
// and the required applications apps. When successful, It will return a value
// and a cleanup function that associated with the initialized application.
// Otherwise, it will return an error.
type InitFunc func(ctx context.Context, lc Lifecycle, apps map[string]*App) (Value, CleanFunc, error)

type OldInitFunc func(ctx context.Context, apps map[string]*App) (Value, CleanFunc, error)

// CleanFunc does the cleanup work for an application. It will return an error if fails.
type CleanFunc func() error

// Value is the value of an application, which is use-case specific and should
// be customized by users.
type Value interface{}

// App is a modular application.
type App struct {
	Name  string
	Value Value

	requiredNames map[string]bool
	requiredApps  map[string]*App
	getAppFunc    func(name string) (*App, error) // The function used to find an application by its name.

	initFunc  InitFunc
	cleanFunc CleanFunc

	state int // The installation state.
}

// New creates an application with the given name.
func New(name string) *App {
	return &App{
		Name:          name,
		requiredNames: make(map[string]bool),
		requiredApps:  make(map[string]*App),
		getAppFunc: func(name string) (*App, error) {
			return nil, fmt.Errorf("app %q is not registered", name)
		},
	}
}

// Require sets the names of the applications that the current application requires.
func (a *App) Require(names ...string) *App {
	for _, name := range names {
		a.requiredNames[name] = true
	}
	return a
}

// Init sets the function used to initialize the current application.
// Init is deprecated in favor of Init2.
func (a *App) Init(initFunc OldInitFunc) *App {
	a.initFunc = func(ctx context.Context, lc Lifecycle, apps map[string]*App) (Value, CleanFunc, error) {
		return initFunc(ctx, apps)
	}
	return a
}

// Init2 sets the function used to initialize the current application.
func (a *App) Init2(initFunc InitFunc) *App {
	a.initFunc = initFunc
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

	// Finally install the app itself.
	if a.initFunc != nil {
		a.Value, a.cleanFunc, err = a.initFunc(ctx, lc, a.requiredApps)
		if err != nil {
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

	if a.cleanFunc != nil {
		if err = a.cleanFunc(); err != nil {
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
