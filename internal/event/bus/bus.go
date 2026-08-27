// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"log/slog"
	"sync"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const (
	MetaKeySource    = "source"
	MetaKeyActorID   = "actor_id"
	MetaKeyTraceID   = "trace_id"
	MetaKeyRequestID = "request_id"
)

func New() *busen.Bus {
	ec := config.Config().Event
	hooks := busen.Hooks{
		OnHandlerError: func(info busen.HandlerError) {
			logUtil.GetLogger().Error("busen handler error",
				slog.String("event_type", safeTypeString(info.EventType)),
				slog.String("topic", info.Topic),
				slog.String("key", info.Key),
				slog.Bool("async", info.Async),
				logUtil.Err(info.Err))
		},
		OnHandlerPanic: func(info busen.HandlerPanic) {
			logUtil.GetLogger().Error("busen handler panic",
				slog.String("event_type", safeTypeString(info.EventType)),
				slog.String("topic", info.Topic),
				slog.String("key", info.Key),
				slog.Bool("async", info.Async),
				slog.Any("panic", info.Value))
		},
		OnPublishDone: func(info busen.PublishDone) {
			if info.Err == nil {
				return
			}
			logUtil.GetLogger().Warn("busen publish done with errors",
				slog.String("event_type", safeTypeString(info.EventType)),
				slog.String("topic", info.Topic),
				slog.String("key", info.Key),
				slog.Int("matched_subscribers", info.MatchedSubscribers),
				slog.Int("delivered_subscribers", info.DeliveredSubscribers),
				logUtil.Err(info.Err))
		},
		OnEventDropped: func(info busen.DroppedEvent) {
			logUtil.GetLogger().Warn("busen event dropped",
				slog.String("event_type", safeTypeString(info.EventType)),
				slog.String("topic", info.Topic),
				slog.String("key", info.Key),
				slog.Int("queue_len", info.QueueLen),
				slog.Int("queue_cap", info.QueueCap),
				logUtil.Err(info.Reason))
		},
		OnEventRejected: func(info busen.RejectedEvent) {
			logUtil.GetLogger().Warn("busen event rejected",
				slog.String("event_type", safeTypeString(info.EventType)),
				slog.String("topic", info.Topic),
				slog.String("key", info.Key),
				slog.Int("queue_len", info.QueueLen),
				slog.Int("queue_cap", info.QueueCap),
				logUtil.Err(info.Reason))
		},
	}

	b := busen.New(
		busen.WithDefaultBuffer(ec.DefaultBuffer),
		busen.WithDefaultOverflow(MapOverflow(ec.DefaultOverflow)),
		busen.WithHooks(hooks),
		busen.WithMetadataBuilder(func(input busen.PublishMetadataInput) map[string]string {
			return map[string]string{"source": "ech0"}
		}),
	)

	return b
}

func ProvideProvider() func() *busen.Bus {
	var once sync.Once
	var b *busen.Bus
	return func() *busen.Bus {
		once.Do(func() {
			b = New()
		})
		return b
	}
}

func MapOverflow(policy string) busen.OverflowPolicy {
	switch policy {
	case "fail_fast":
		return busen.OverflowFailFast
	case "drop_newest":
		return busen.OverflowDropNewest
	case "drop_oldest":
		return busen.OverflowDropOldest
	default:
		return busen.OverflowBlock
	}
}
