// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"context"
	"log/slog"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

func (wd *Dispatcher) Registrations() []eventbus.Registration {
	return []eventbus.Registration{
		observe[event.UserCreated](wd.HandleObservation),
		observe[event.UserUpdated](wd.HandleObservation),
		observe[event.UserDeleted](wd.HandleObservation),
		observe[event.EchoCreated](wd.HandleObservation),
		observe[event.EchoUpdated](wd.HandleObservation),
		observe[event.EchoDeleted](wd.HandleObservation),
		observe[event.CommentCreated](wd.HandleObservation),
		observe[event.CommentStatusUpdated](wd.HandleObservation),
		observe[event.CommentDeleted](wd.HandleObservation),
		observe[event.ResourceUploaded](wd.HandleObservation),
		observe[event.SystemSnapshot](wd.HandleObservation),
		observe[event.SystemExport](wd.HandleObservation),
		observe[event.UpdateSnapshotSchedule](wd.HandleObservation),
	}
}

func observe[T event.Named](
	deliver func(context.Context, event.WebhookObservation) error,
) eventbus.Registration {
	return eventbus.OnWithMeta(func(ctx context.Context, v T, meta map[string]string) error {
		obs, err := event.NewWebhookObservation(v.EventName(), v, meta)
		if err != nil {
			logUtil.GetLogger().Warn("build webhook observation failed",
				slog.String("event", v.EventName()), logUtil.Err(err))
			return nil
		}
		if err := deliver(ctx, obs); err != nil {
			logUtil.GetLogger().Warn("dispatch webhook observation failed",
				slog.String("event", v.EventName()), logUtil.Err(err))
		}
		return nil
	})
}
