// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package task

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const logModule = "task"

type Manager struct {
	scheduler gocron.Scheduler
	tasks     []Task

	mu      sync.Mutex
	started bool
}

func NewManager(tasks ...Task) (*Manager, error) {
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return nil, err
	}
	return &Manager{scheduler: scheduler, tasks: tasks}, nil
}

func (m *Manager) Name() string { return "task" }

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}
	if m.scheduler == nil {
		return errors.New("scheduler is nil")
	}

	for _, t := range m.tasks {
		if err := t.Schedule(ctx, m.scheduler); err != nil {
			logUtil.GetLogger().Error("failed to schedule task",
				slog.String("module", logModule), slog.String("task", t.Name()), logUtil.Err(err))
			return err
		}
	}

	m.scheduler.Start()
	m.started = true
	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started || m.scheduler == nil {
		return nil
	}

	for _, t := range m.tasks {
		if h, ok := t.(StopHook); ok {
			h.OnStop(ctx)
		}
	}

	if err := m.scheduler.Shutdown(); err != nil {
		logUtil.GetLogger().Error("failed to shutdown scheduler",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	m.started = false
	return nil
}
