package app

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

// App 是应用生命周期编排器。
type App struct {
	mu sync.Mutex

	lifecycles []Lifecycle

	running bool
}

// NewApp 创建应用生命周期编排器。
func NewApp(lifecycles []Lifecycle) *App {
	return &App{
		lifecycles: lifecycles,
	}
}

// Start 按顺序启动应用组件链。
func (a *App) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "app.start",
			Component: "app",
		}
	}
	if len(a.lifecycles) == 0 {
		return &AppError{
			Code:      CodeDependencyMissing,
			Op:        "app.start",
			Component: "lifecycles",
		}
	}

	started := make([]Lifecycle, 0, len(a.lifecycles))
	for _, lifecycle := range a.lifecycles {
		if lifecycle == nil {
			return &AppError{
				Code:      CodeDependencyMissing,
				Op:        "app.start",
				Component: "nil_lifecycle",
			}
		}
		if err := lifecycle.Start(ctx); err != nil {
			a.stopLifecyclesReverse(ctx, started)
			return &AppError{
				Code:      CodeComponentStartFailed,
				Op:        "app.start",
				Component: lifecycleName(lifecycle),
				Cause:     err,
			}
		}
		started = append(started, lifecycle)
	}

	a.running = true
	return nil
}

// Stop 按反向顺序停止应用生命周期单元。
func (a *App) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return &AppError{
			Code:      CodeInvalidState,
			Op:        "app.stop",
			Component: "app",
		}
	}

	var errs []error
	for i := len(a.lifecycles) - 1; i >= 0; i-- {
		lifecycle := a.lifecycles[i]
		if lifecycle == nil {
			continue
		}
		if err := lifecycle.Stop(ctx); err != nil {
			errs = append(errs, &AppError{
				Code:      CodeComponentStopFailed,
				Op:        "app.stop",
				Component: lifecycleName(lifecycle),
				Cause:     err,
			})
		}
	}

	a.running = false
	return errors.Join(errs...)
}

// StopAll 停止所有已运行组件。
func (a *App) StopAll(ctx context.Context) error {
	var errs []error

	if a.IsRunning() {
		if err := a.Stop(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// IsRunning 返回应用组件链是否运行中。
func (a *App) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

func (a *App) stopLifecyclesReverse(ctx context.Context, lifecycles []Lifecycle) {
	for i := len(lifecycles) - 1; i >= 0; i-- {
		_ = lifecycles[i].Stop(ctx)
	}
}

func lifecycleName(lifecycle Lifecycle) string {
	if lifecycle == nil {
		return "nil_lifecycle"
	}
	if namer, ok := lifecycle.(Namer); ok {
		return namer.Name()
	}

	t := reflect.TypeOf(lifecycle)
	if t == nil {
		return "unknown_lifecycle"
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return "unknown_lifecycle"
}
