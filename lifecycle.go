package appx

import (
	"context"
)

// Lifecycle allows application initializers to register callbacks that are
// executed on application start and stop.
//
// The concept of lifecycle is borrowed from https://github.com/uber-go/fx.
type Lifecycle interface {
	Append(Hook)
}

// A Hook is a pair of start and stop callbacks, either of which can be nil.
// If a Hook's OnStart callback isn't executed (because a previous OnStart
// failure short-circuited application startup), its OnStop callback won't be
// executed.
type Hook struct {
	OnStart func(ctx context.Context) error
	OnStop  func(ctx context.Context) error
}

type lifecycleImpl struct {
	hooks      []Hook
	numStarted int
}

func (l *lifecycleImpl) Append(hook Hook) {
	l.hooks = append(l.hooks, hook)
}

// Start runs all OnStart hooks, returning immediately if it encounters an
// error.
func (l *lifecycleImpl) Start(ctx context.Context) error {
	for _, hook := range l.hooks {
		if hook.OnStart != nil {
			if err := hook.OnStart(ctx); err != nil {
				return err
			}
		}
		l.numStarted++
	}
	return nil
}

// Stop runs any OnStop hooks whose OnStart counterpart succeeded. OnStop
// hooks run in reverse order.
func (l *lifecycleImpl) Stop(ctx context.Context) (errs []error) {
	// Run backward from last successful OnStart.
	for ; l.numStarted > 0; l.numStarted-- {
		hook := l.hooks[l.numStarted-1]
		if hook.OnStop == nil {
			continue
		}
		if err := hook.OnStop(ctx); err != nil {
			// For best-effort cleanup, keep going after errors.
			errs = append(errs, err)
		}
	}
	return
}
