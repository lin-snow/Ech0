// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
	"github.com/lin-snow/ech0/internal/visitor"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const visitorSnapshotTag = "VisitorSnapshotSchedule"

type VisitorSnapshot struct {
	tracker *visitor.Tracker
	repo    *visitorRepository.VisitorRepository
}

func NewVisitorSnapshot(tracker *visitor.Tracker, repo *visitorRepository.VisitorRepository) *VisitorSnapshot {
	return &VisitorSnapshot{tracker: tracker, repo: repo}
}

func (v *VisitorSnapshot) Name() string { return "visitor-snapshot" }

func (v *VisitorSnapshot) Schedule(ctx context.Context, s gocron.Scheduler) error {
	if err := v.restore(ctx); err != nil {
		return err
	}

	_, err := s.NewJob(
		gocron.DurationJob(60*time.Minute),
		gocron.NewTask(func() {
			v.flush(context.Background())
			cutoff := cutoffDate(time.Now().UTC())
			if err := v.repo.DeleteOlderThan(context.Background(), cutoff); err != nil {
				logUtil.GetLogger().Error("Failed to cleanup visitor stats",
					slog.String("module", logModule), logUtil.Err(err))
			}
		}),
		gocron.WithTags(visitorSnapshotTag),
	)
	if err != nil {
		logUtil.GetLogger().Error("Failed to schedule visitor snapshot task",
			slog.String("module", logModule), logUtil.Err(err))
	}
	return err
}

func (v *VisitorSnapshot) OnStop(context.Context) {
	v.flush(context.Background())
}

func (v *VisitorSnapshot) flush(ctx context.Context) {
	if v.tracker == nil || v.repo == nil {
		return
	}
	today := v.tracker.TodayStat()
	if err := v.repo.UpsertDailyStat(ctx, buildDailyStat(today)); err != nil {
		logUtil.GetLogger().Error("Failed to upsert visitor daily stat",
			slog.String("module", logModule), logUtil.Err(err))
	}
}

func (v *VisitorSnapshot) restore(ctx context.Context) error {
	if v.tracker == nil || v.repo == nil {
		return nil
	}
	stats, err := v.repo.GetRecentDays(ctx, 7)
	if err != nil {
		logUtil.GetLogger().Error("Failed to load visitor stats",
			slog.String("module", logModule), logUtil.Err(err))
		return err
	}
	if len(stats) == 0 {
		return nil
	}
	v.tracker.LoadHistory(convertHistory(stats))
	return nil
}

func buildDailyStat(stat visitor.DayStat) visitorModel.DailyStat {
	return visitorModel.DailyStat{Date: stat.Date, PV: stat.PV, UV: stat.UV}
}

func convertHistory(stats []visitorModel.DailyStat) []visitor.DayStat {
	history := make([]visitor.DayStat, 0, len(stats))
	for _, s := range stats {
		history = append(history, visitor.DayStat{Date: s.Date, PV: s.PV, UV: s.UV})
	}
	return history
}

func cutoffDate(now time.Time) string {
	return now.UTC().AddDate(0, 0, -6).Format("2006-01-02")
}
