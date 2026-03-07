package app

import "fmt"

const (
	CodeComponentStartFailed = "COMPONENT_START_FAILED"
	CodeComponentStopFailed  = "COMPONENT_STOP_FAILED"
	CodeDependencyMissing    = "DEPENDENCY_MISSING"
	CodeInvalidState         = "INVALID_STATE"
)

// AppError 统一应用层生命周期错误。
type AppError struct {
	Code      string
	Op        string
	Component string
	Cause     error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s (%s)", e.Op, e.Component, e.Code)
	}
	return fmt.Sprintf("%s: %s (%s): %v", e.Op, e.Component, e.Code, e.Cause)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}
