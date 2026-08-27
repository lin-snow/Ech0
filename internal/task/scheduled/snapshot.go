// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"log/slog"
	"strings"
	"sync"

	"github.com/go-co-op/gocron/v2"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/kvstore"
	coreMigrator "github.com/lin-snow/ech0/internal/migrator"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/pkg/busen"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const snapshotScheduleTag = "SnapshotSchedule"

type Snapshot struct {
	durableKV kvstore.Store
	exporter  *coreMigrator.ExportEngine
	bus       *busen.Bus

	mu        sync.Mutex
	scheduler gocron.Scheduler
	unsub     func()
}

func NewSnapshot(
	durableKV kvstore.Store,
	exporter *coreMigrator.ExportEngine,
	busProvider func() *busen.Bus,
) *Snapshot {
	return &Snapshot{durableKV: durableKV, exporter: exporter, bus: busProvider()}
}

func (s *Snapshot) Name() string { return "snapshot" }

func (s *Snapshot) Schedule(ctx context.Context, scheduler gocron.Scheduler) error {
	s.mu.Lock()
	s.scheduler = scheduler
	s.mu.Unlock()

	if err := s.subscribe(); err != nil {
		return err
	}
	return s.reload(ctx)
}

func (s *Snapshot) OnStop(_ context.Context) {
	s.mu.Lock()
	unsub := s.unsub
	s.unsub = nil
	s.mu.Unlock()
	if unsub != nil {
		unsub()
	}
}

func (s *Snapshot) subscribe() error {
	unsub, err := eventbus.On(s.handleScheduleChanged, eventbus.AsyncSequential()...)(s.bus)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.unsub = unsub
	s.mu.Unlock()
	return nil
}

func (s *Snapshot) handleScheduleChanged(ctx context.Context, _ event.UpdateSnapshotSchedule) error {
	return s.reload(ctx)
}

func (s *Snapshot) reload(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.scheduler == nil {
		return nil
	}

	schedule, err := coreSetting.Get(ctx, s.durableKV, coreSetting.Snapshot)
	if err != nil {
		logUtil.GetLogger().Error("Failed to read snapshot schedule, keeping current jobs",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}

	s.scheduler.RemoveByTags(snapshotScheduleTag)
	if !schedule.Enable {
		logUtil.GetLogger().Info("Snapshot schedule disabled, jobs removed", slog.String("module", logModule))
		return nil
	}

	if err := s.scheduleJob(schedule.CronExpression); err != nil {
		logUtil.GetLogger().Error("Failed to apply snapshot schedule",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	logUtil.GetLogger().Info("Snapshot schedule applied",
		slog.String("module", logModule), slog.String("cron", schedule.CronExpression))
	return nil
}

func (s *Snapshot) scheduleJob(cronExpression string) error {
	withSeconds := len(strings.Fields(cronExpression)) == 6

	_, err := s.scheduler.NewJob(
		gocron.CronJob(cronExpression, withSeconds),
		gocron.NewTask(func() {
			ctx := context.Background()

			if _, err := s.exporter.Export(ctx, func(string, any) {}); err != nil {
				logUtil.GetLogger().Error("Failed to execute scheduled snapshot",
					slog.String("module", logModule),
					logUtil.Err(err))
				return
			}

			eventbus.Notify(ctx, s.bus, event.SystemSnapshot{Info: "System scheduled snapshot completed"})
		}),
		gocron.WithTags(snapshotScheduleTag),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule snapshot task",
			slog.String("module", logModule), logUtil.Err(err))
	}
	return err
}
