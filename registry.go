package appx

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Registry struct {
	// The common middleware stack that will be applied to all registered applications.
	middlewares []func(Standard) Standard

	// registered holds all the registered applications.
	registered map[string]*App

	// installed holds all the installed applications, in dependency order.
	installed []*App

	// lifecycle holds the Start and Stop callbacks of the runnable applications.
	lifecycle *lifecycleImpl

	options *Options
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{
		registered: make(map[string]*App),
		lifecycle:  new(lifecycleImpl),
		options:    new(Options).init(),
	}
}

// Register registers the application app into the registry.
func (r *Registry) Register(app *App) error {
	if app == nil {
		return fmt.Errorf("nil app %v", app)
	}

	if app.Name == "" {
		return fmt.Errorf("the name of app %v is empty", app)
	}

	if _, ok := r.registered[app.Name]; ok {
		return fmt.Errorf("app %q is already registered", app.Name)
	}

	r.registered[app.Name] = app
	app.getAppFunc = r.getApp // Find an application in the registry.
	app.getConfigFunc = func(name string) interface{} { return r.options.AppConfigs[name] }
	return nil
}

// MustRegister is like Register but panics if there is an error.
func (r *Registry) MustRegister(app *App) {
	if err := r.Register(app); err != nil {
		panic(err)
	}
}

// SetOptions sets the options for the registry.
func (r *Registry) SetOptions(opts *Options) {
	r.options = opts.init()
}

// Use appends one or more middlewares to the common middleware stack, which
// will be applied to all registered applications.
func (r *Registry) Use(middlewares ...func(Standard) Standard) {
	if len(r.installed) > 0 {
		panic("appx: all middlewares must be defined prior to installation")
	}
	r.middlewares = append(r.middlewares, middlewares...)
}

// Install installs the applications specified by names, with the given ctx.
// If no name is specified, all registered applications will be installed.
//
// Note that applications will be installed in dependency order.
func (r *Registry) Install(ctx context.Context, names ...string) error {
	before := func(app *App) {
		// Insert common middlewares, if any, before app's middlewares.
		if len(r.middlewares) > 0 {
			mws := make([]func(Standard) Standard, len(r.middlewares))
			copy(mws, r.middlewares)
			mws = append(mws, app.middlewares...)

			app.middlewares = mws
		}
	}
	after := func(app *App) {
		r.installed = append(r.installed, app)
	}

	if len(names) == 0 {
		for _, app := range r.registered {
			if err := app.Install(ctx, r.lifecycle, before, after); err != nil {
				// Install failed, roll back.
				Uninstall()
				return fmt.Errorf("install failed for app %q, err: %s", app.Name, err)
			}
		}
	}

	for _, name := range names {
		app, err := r.getApp(name)
		if err != nil {
			// Install failed, roll back.
			Uninstall()
			return fmt.Errorf("install failed for app %q, err: %s", app.Name, err)
		}
		if err := app.Install(ctx, r.lifecycle, before, after); err != nil {
			// Install failed, roll back.
			Uninstall()
			return fmt.Errorf("install failed for app %q, err: %s", app.Name, err)
		}
	}

	return nil
}

// Uninstall uninstalls the applications that has already been installed, in
// the reverse order of installation.
func (r *Registry) Uninstall() {
	for i := len(r.installed); i > 0; i-- {
		if err := r.installed[i-1].Uninstall(); err != nil {
			r.options.ErrorHandler(err)
		}
	}
}

func (r *Registry) getApp(name string) (*App, error) {
	app, ok := r.registered[name]
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
func (r *Registry) Run() (os.Signal, error) {
	startCtx, cancel := context.WithTimeout(context.Background(), r.options.StartTimeout)
	defer cancel()
	if err := r.Start(startCtx); err != nil {
		return nil, err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c

	stopCtx, cancel := context.WithTimeout(context.Background(), r.options.StopTimeout)
	defer cancel()
	r.Stop(stopCtx)

	return sig, nil
}

// Start kicks off all long-running applications, like network servers or
// message queue consumers. It will returns immediately if it encounters an error.
func (r *Registry) Start(ctx context.Context) error {
	return withTimeout(ctx, r.start)
}

// Stop gracefully stops all long-running applications. For best-effort cleanup,
// It will keep going after encountering errors, and all errors will be passed
// to the handler specified by ErrorHandler.
func (r *Registry) Stop(ctx context.Context) {
	_ = withTimeout(ctx, r.stop)
}

func (r *Registry) start(ctx context.Context) error {
	if err := r.lifecycle.Start(ctx); err != nil {
		// Start failed, roll back.
		_ = r.stop(ctx)
		return err
	}
	return nil
}

func (r *Registry) stop(ctx context.Context) error {
	for _, err := range r.lifecycle.Stop(ctx) {
		r.options.ErrorHandler(err)
	}
	return nil
}

// Graph generates the dependency graph for all installed applications in map form.
//
// The format of the returned map is as below:
//
//	appName -> [dependencyAppName1, dependencyAppName2, ...]
func (r *Registry) Graph() map[string][]string {
	graph := make(map[string][]string)
	for _, app := range r.installed {
		graph[app.Name] = app.Requirements()
	}
	return graph
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

// Options is a set of optional configurations for a registry.
type Options struct {
	// The timeout of application startup. Defaults to 15s.
	StartTimeout time.Duration

	// The timeout of application shutdown. Defaults to 15s.
	StopTimeout time.Duration

	// The handler for errors during the Stop and Uninstall phases.
	ErrorHandler func(error)

	// The configurations for all registered applications.
	AppConfigs map[string]interface{}
}

func (o *Options) init() *Options {
	if o.StartTimeout == 0 {
		o.StartTimeout = 15 * time.Second
	}
	if o.StopTimeout == 0 {
		o.StopTimeout = 15 * time.Second
	}
	if o.ErrorHandler == nil {
		o.ErrorHandler = func(error) {}
	}
	return o
}
