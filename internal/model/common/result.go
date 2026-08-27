// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type Result[T any] struct {
	Code          int            `json:"code"`
	Message       string         `json:"msg"`
	ErrorCode     string         `json:"error_code,omitempty"`
	MessageKey    string         `json:"message_key,omitempty"`
	MessageParams map[string]any `json:"message_params,omitempty"`
	Data          T              `json:"data"`
}

const (
	DEFAULT_SUCCESS_CODE = 1
	DEFAULT_FAILED_CODE  = 0
)

func OK[T any](data T, messages ...string) Result[T] {
	message := SUCCESS_MESSAGE
	if len(messages) > 0 {
		message = messages[0]
	}

	return Result[T]{
		Code:    DEFAULT_SUCCESS_CODE,
		Message: message,
		Data:    data,
	}
}

func Fail[T any](message string) Result[T] {
	var zero T
	return Result[T]{
		Code:       DEFAULT_FAILED_CODE,
		Message:    message,
		ErrorCode:  "",
		MessageKey: "",
		Data:       zero,
	}
}

func FailWithErrorCode[T any](message, errorCode string) Result[T] {
	res := Fail[T](message)
	res.ErrorCode = errorCode
	return res
}

func FailWithLocalized[T any](message, errorCode, messageKey string, params map[string]any) Result[T] {
	res := FailWithErrorCode[T](message, errorCode)
	res.MessageKey = messageKey
	res.MessageParams = params
	return res
}

func OKWithCode[T any](data T, code int, messages ...string) Result[T] {
	message := SUCCESS_MESSAGE
	if len(messages) > 0 {
		message = messages[0]
	}

	return Result[T]{
		Code:    code,
		Message: message,
		Data:    data,
	}
}
