// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"fmt"
	"reflect"
	"slices"
)

type Dispatch struct {
	EventType reflect.Type
	Topic     string
	Key       string
	Headers   map[string]string
	Meta      map[string]string
	Value     any
	Async     bool
}

type Next func(context.Context, Dispatch) error

type Middleware func(Next) Next

func (b *Bus) Use(middlewares ...Middleware) error {
	if b == nil {
		return fmt.Errorf("%w: nil bus", ErrInvalidOption)
	}
	if b.gate.Closed() {
		return ErrClosed
	}
	if len(middlewares) == 0 {
		return nil
	}

	b.middlewareMu.Lock()
	defer b.middlewareMu.Unlock()

	combined := make([]Middleware, 0, len(b.middlewares)+len(middlewares))
	combined = append(combined, b.middlewares...)
	for _, middleware := range middlewares {
		if middleware == nil {
			return fmt.Errorf("%w: middleware is nil", ErrInvalidOption)
		}
		combined = append(combined, middleware)
	}
	b.middlewares = combined
	b.middleware = buildMiddlewareChain(combined)
	b.middlewareVersion.Add(1)
	return nil
}

func buildMiddlewareChain(middlewares []Middleware) func(Next) Next {
	if len(middlewares) == 0 {
		return nil
	}

	cached := append([]Middleware(nil), middlewares...)
	return func(next Next) Next {
		wrapped := next
		for _, c := range slices.Backward(cached) {
			wrapped = c(wrapped)
		}
		return wrapped
	}
}
