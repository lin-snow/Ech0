// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"context"
	"log/slog"

	"github.com/lin-snow/ech0/internal/event"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

func Emit[T any](ctx context.Context, b *busen.Bus, evt T) error {
	var opts []busen.PublishOption
	if k, ok := any(evt).(event.Keyed); ok {
		if key := k.OrderingKey(); key != "" {
			opts = append(opts, busen.WithKey(key))
		}
	}
	return b.Publish(ctx, evt, opts...)
}

func Notify[T any](ctx context.Context, b *busen.Bus, evt T) {
	if err := Emit(ctx, b, evt); err != nil {
		name := ""
		if n, ok := any(evt).(event.Named); ok {
			name = n.EventName()
		}
		logUtil.GetLogger().Warn("event publish failed", slog.String("event", name), logUtil.Err(err))
	}
}

// Record participates in the mutation transaction on Vercel. Local subscribers
// receive the event only once the outermost transaction has committed.
func Record[T any](ctx context.Context, b *busen.Bus, evt T) error {
	if transaction.Durable(ctx) {
		return Emit(ctx, b, evt)
	}
	transaction.AfterCommit(ctx, func() { Notify(context.Background(), b, evt) })
	return nil
}
