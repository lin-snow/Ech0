// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package app

import "fmt"

const (
	CodeComponentStartFailed = "COMPONENT_START_FAILED"
	CodeComponentStopFailed  = "COMPONENT_STOP_FAILED"
	CodeDependencyMissing    = "DEPENDENCY_MISSING"
	CodeHookFailed           = "HOOK_FAILED"
	CodeInvalidState         = "INVALID_STATE"
)

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
