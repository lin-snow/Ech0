// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

// Cleanup keeps the file tables honest: unfinished direct uploads die, and
// bytes whose row is already gone are retried until the store lets go.
type Cleanup struct {
	fileService fileService.Service
}

func NewCleanup(fileSvc fileService.Service) *Cleanup {
	return &Cleanup{fileService: fileSvc}
}

func (c *Cleanup) Name() string { return "file-lifecycle" }

func (c *Cleanup) Schedule(_ context.Context, s gocron.Scheduler) error {
	if _, err := s.NewJob(
		gocron.DurationJob(10*time.Minute),
		gocron.NewTask(c.sweepPending),
	); err != nil {
		logUtil.GetLogger().Error("Failed to schedule pending upload sweep",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	if _, err := s.NewJob(
		gocron.DurationJob(time.Minute),
		gocron.NewTask(c.drainGC),
	); err != nil {
		logUtil.GetLogger().Error("Failed to schedule file GC drain",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	return nil
}

func (c *Cleanup) sweepPending() {
	swept, err := c.fileService.SweepPending(context.Background(), 500)
	if err != nil {
		logUtil.GetLogger().Error("Failed to sweep pending uploads",
			slog.String("module", "storage.gc"), logUtil.Err(err))
		return
	}
	if swept > 0 {
		logUtil.GetLogger().Info("Swept pending uploads",
			slog.String("module", "storage.gc"), slog.Int("count", swept))
	}
}

func (c *Cleanup) drainGC() {
	drained, err := c.fileService.DrainGC(context.Background(), 100)
	if err != nil {
		logUtil.GetLogger().Error("Failed to drain file GC",
			slog.String("module", "storage.gc"), logUtil.Err(err))
		return
	}
	if drained > 0 {
		logUtil.GetLogger().Info("Drained file GC",
			slog.String("module", "storage.gc"), slog.Int("count", drained))
	}
}
