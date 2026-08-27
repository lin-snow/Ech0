// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import "errors"

var (
	ErrClosed = errors.New("busen: bus closed")

	ErrHandlerNil = errors.New("busen: handler is nil")

	ErrBufferFull = errors.New("busen: subscriber buffer full")

	ErrDropped = errors.New("busen: event dropped")

	ErrInvalidPattern = errors.New("busen: invalid topic pattern")

	ErrInvalidOption = errors.New("busen: invalid option")

	ErrHandlerPanic = errors.New("busen: handler panic")

	ErrCloseIncomplete = errors.New("busen: close incomplete")
)
