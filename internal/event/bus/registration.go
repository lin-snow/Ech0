// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"context"

	"github.com/lin-snow/ech0/pkg/busen"
)

type Registration func(*busen.Bus) (func(), error)

type Subscriber interface {
	Registrations() []Registration
}

func On[T any](handler func(context.Context, T) error, opts ...busen.SubscribeOption) Registration {
	return func(b *busen.Bus) (func(), error) {
		return b.Subscribe(func(ctx context.Context, e busen.Event[T]) error {
			return handler(ctx, e.Value)
		}, opts...)
	}
}

func OnWithMeta[T any](
	handler func(context.Context, T, map[string]string) error,
	opts ...busen.SubscribeOption,
) Registration {
	return func(b *busen.Bus) (func(), error) {
		return b.Subscribe(func(ctx context.Context, e busen.Event[T]) error {
			return handler(ctx, e.Value, e.Meta)
		}, opts...)
	}
}
