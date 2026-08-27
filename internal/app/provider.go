// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package app

import (
	"context"

	"github.com/google/wire"
	bus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/job"
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/task"
)

func ProvideOptions(
	registrar *bus.EventRegistrar,
	jobManager *job.Manager,
	taskManager *task.Manager,
	httpServer *server.Server,
	durableKV kvstore.Store,
) []Option {
	return []Option{
		Components(jobManager, taskManager, httpServer),
		BeforeStart(func(ctx context.Context) error {
			if err := setting.Seed(ctx, durableKV); err != nil {
				return err
			}
			return registrar.Register()
		}),
		AfterStop(func(context.Context) error {
			return registrar.Stop()
		}),
	}
}

func NewApp(opts []Option) *App {
	return New(opts...)
}

var ProviderSet = wire.NewSet(ProvideOptions, NewApp)
