// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"reflect"
	"slices"
	"sync"
)

type App struct {
	mu sync.Mutex

	opts options

	ctx    context.Context
	cancel context.CancelFunc

	running    bool
	stopping   bool
	stopErr    error
	stoppedCh  chan struct{}
	startedSet []Component
}

func New(opts ...Option) *App {
	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}
	ctx, cancel := context.WithCancel(o.ctx)
	return &App{
		opts:      o,
		ctx:       ctx,
		cancel:    cancel,
		stoppedCh: make(chan struct{}),
	}
}

func (a *App) Run() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "app.run",
			Component: "app",
		}
	}
	if len(a.opts.components) == 0 {
		a.mu.Unlock()
		return &AppError{
			Code:      CodeDependencyMissing,
			Op:        "app.run",
			Component: "components",
		}
	}
	a.running = true
	a.stopping = false
	a.stopErr = nil
	a.startedSet = nil
	a.stoppedCh = make(chan struct{})
	a.mu.Unlock()

	if err := a.runHooks(a.ctx, "app.before_start", a.opts.beforeStart); err != nil {
		a.mu.Lock()
		a.running = false
		close(a.stoppedCh)
		a.mu.Unlock()
		return err
	}

	started := make([]Component, 0, len(a.opts.components))
	for _, component := range a.opts.components {
		if component == nil {
			a.rollbackStart(started)
			a.mu.Lock()
			a.running = false
			close(a.stoppedCh)
			a.mu.Unlock()
			return &AppError{
				Code:      CodeDependencyMissing,
				Op:        "app.run",
				Component: "nil_component",
			}
		}
		if err := component.Start(a.ctx); err != nil {
			a.rollbackStart(started)
			a.mu.Lock()
			a.running = false
			close(a.stoppedCh)
			a.mu.Unlock()
			return &AppError{
				Code:      CodeComponentStartFailed,
				Op:        "app.run",
				Component: componentName(component),
				Cause:     err,
			}
		}
		started = append(started, component)
	}
	a.mu.Lock()
	a.startedSet = append([]Component(nil), started...)
	a.mu.Unlock()

	if err := a.runHooks(a.ctx, "app.after_start", a.opts.afterStart); err != nil {
		_ = a.Stop()
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, a.opts.sigs...)
	defer signal.Stop(c)

	select {
	case <-a.ctx.Done():
	case <-c:
		if err := a.Stop(); err != nil {
			return err
		}
	}

	<-a.stoppedCh
	return a.stopErr
}

func (a *App) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	if a.stopping {
		ch := a.stoppedCh
		a.mu.Unlock()
		<-ch
		return a.stopErr
	}
	a.stopping = true
	stoppedCh := a.stoppedCh
	components := append([]Component(nil), a.startedSet...)
	beforeStop := append([]Hook(nil), a.opts.beforeStop...)
	afterStop := append([]Hook(nil), a.opts.afterStop...)
	stopTimeout := a.opts.stopTimeout
	baseCtx := context.WithoutCancel(a.ctx)
	cancel := a.cancel
	a.mu.Unlock()

	stopCtx := baseCtx
	var stopCancel context.CancelFunc
	if stopTimeout > 0 {
		stopCtx, stopCancel = context.WithTimeout(stopCtx, stopTimeout)
		defer stopCancel()
	}

	var errs []error
	if err := a.runHooks(stopCtx, "app.before_stop", beforeStop); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, a.stopComponentsReverse(stopCtx, components)...)
	if err := a.runHooks(stopCtx, "app.after_stop", afterStop); err != nil {
		errs = append(errs, err)
	}

	if cancel != nil {
		cancel()
	}
	stopErr := errors.Join(errs...)

	a.mu.Lock()
	a.running = false
	a.stopping = false
	a.startedSet = nil
	a.stopErr = stopErr
	close(stoppedCh)
	a.mu.Unlock()
	return stopErr
}

func (a *App) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

func (a *App) rollbackStart(started []Component) {
	stopCtx := context.WithoutCancel(a.ctx)
	if a.opts.stopTimeout > 0 {
		var cancel context.CancelFunc
		stopCtx, cancel = context.WithTimeout(stopCtx, a.opts.stopTimeout)
		defer cancel()
	}
	_ = errors.Join(a.stopComponentsReverse(stopCtx, started)...)
}

func (a *App) runHooks(ctx context.Context, op string, hooks []Hook) error {
	for _, hook := range hooks {
		if hook == nil {
			continue
		}
		if err := hook(ctx); err != nil {
			return &AppError{
				Code:      CodeHookFailed,
				Op:        op,
				Component: "hook",
				Cause:     err,
			}
		}
	}
	return nil
}

func (a *App) stopComponentsReverse(ctx context.Context, components []Component) []error {
	errs := make([]error, 0)
	for _, component := range slices.Backward(components) {

		if component == nil {
			continue
		}
		if err := component.Stop(ctx); err != nil {
			errs = append(errs, &AppError{
				Code:      CodeComponentStopFailed,
				Op:        "app.stop",
				Component: componentName(component),
				Cause:     err,
			})
		}
	}
	return errs
}

func (a *App) StopAll(context.Context) error {
	return a.Stop()
}

func componentName(component Component) string {
	if component == nil {
		return "nil_component"
	}
	if namer, ok := component.(Namer); ok {
		return namer.Name()
	}

	t := reflect.TypeOf(component)
	if t == nil {
		return "unknown_component"
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return "unknown_component"
}
