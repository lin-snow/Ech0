// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

import "net/http"

type ErrorCode string

const (
	ErrCodeBadRequest ErrorCode = "bad_request"
	ErrCodeForbidden  ErrorCode = "forbidden"
	ErrCodeNotFound   ErrorCode = "not_found"
	ErrCodeRateLimit  ErrorCode = "rate_limit"
	ErrCodeInternal   ErrorCode = "internal"
)

type Error struct {
	Code       ErrorCode
	Message    string
	StatusCode int
}

func (e *Error) Error() string {
	return e.Message
}

func NewBadRequest(msg string) *Error {
	return &Error{Code: ErrCodeBadRequest, Message: msg, StatusCode: http.StatusBadRequest}
}

func NewForbidden(msg string) *Error {
	return &Error{Code: ErrCodeForbidden, Message: msg, StatusCode: http.StatusForbidden}
}

func NewNotFound(msg string) *Error {
	return &Error{Code: ErrCodeNotFound, Message: msg, StatusCode: http.StatusNotFound}
}

func NewRateLimit(msg string) *Error {
	return &Error{Code: ErrCodeRateLimit, Message: msg, StatusCode: http.StatusTooManyRequests}
}

func NewInternal(msg string) *Error {
	return &Error{Code: ErrCodeInternal, Message: msg, StatusCode: http.StatusInternalServerError}
}
