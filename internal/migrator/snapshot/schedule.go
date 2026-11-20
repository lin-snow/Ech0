// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/lin-snow/ech0/internal/job"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	formatUtil "github.com/lin-snow/ech0/internal/util/format"
	"gorm.io/gorm"
)

// The hourly reconciliation also covers occurrences farther away than Queue TTL.
func NextOccurrence(s settingModel.SnapshotSchedule, after time.Time) (*Wake, error) {
	if !s.Enable {
		return nil, nil
	}
	cron, err := formatUtil.ParseCrontabExpression(s.CronExpression)
	if err != nil {
		return nil, err
	}
	next := cron.Next(after.UTC())
	if next.IsZero() || next.Sub(after) > 48*time.Hour {
		return nil, nil
	}
	digest := sha256.Sum256([]byte(s.CronExpression))
	return &Wake{SnapshotDue: next.Unix(), Schedule: hex.EncodeToString(digest[:12])}, nil
}
func (w Wake) ScheduleKey() string {
	return "schedule-" + w.Schedule + "-" + strconv.FormatInt(w.SnapshotDue, 10)
}

func RunOccurrence(ctx context.Context, db *gorm.DB, s settingModel.SnapshotSchedule, w Wake, now time.Time) error {
	if w.SnapshotDue > now.Unix() {
		return errors.New("snapshot occurrence not due")
	}
	expected, err := NextOccurrence(s, time.Unix(w.SnapshotDue-1, 0))
	if err != nil || expected == nil || expected.Schedule != w.Schedule || expected.SnapshotDue != w.SnapshotDue {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := setupExports(tx); err != nil {
			return err
		}
		insert := tx.Exec("INSERT INTO ech0_export_occurrences(id,due) VALUES(?,?) ON CONFLICT(id) DO NOTHING", w.ScheduleKey(), w.SnapshotDue)
		if insert.Error != nil || insert.RowsAffected == 0 {
			return insert.Error
		}
		if err := tx.Exec("DELETE FROM ech0_export_occurrences WHERE due<?", now.Add(-7*24*time.Hour).Unix()).Error; err != nil {
			return err
		}
		_, _, err := StartExport(ctx, tx)
		if errors.Is(err, job.ErrAlreadyRunning) {
			return nil
		}
		return err
	})
}
