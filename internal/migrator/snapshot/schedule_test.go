// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"path/filepath"
	"testing"
	"time"

	jobModel "github.com/lin-snow/ech0/internal/model/job"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCronOccurrenceMatchesConfigurationOnce(t *testing.T) {
	now := time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC)
	config := settingModel.SnapshotSchedule{Enable: true, CronExpression: "15 * * * * *"}
	next, err := NextOccurrence(config, now)
	require.NoError(t, err)
	require.Equal(t, now.Add(15*time.Second).Unix(), next.SnapshotDue)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "schedule.db")))
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	defer func() { _ = raw.Close() }()
	require.NoError(t, db.AutoMigrate(&jobModel.Job{}))
	ctx := t.Context()
	// An obsolete and a disabled schedule must never start an export.
	require.NoError(t, RunOccurrence(ctx, db, settingModel.SnapshotSchedule{}, *next, now.Add(time.Minute)))
	require.NoError(t, RunOccurrence(ctx, db, settingModel.SnapshotSchedule{Enable: true, CronExpression: "30 * * * * *"}, *next, now.Add(time.Minute)))
	job, err := PendingExport(ctx, db)
	require.NoError(t, err)
	require.Nil(t, job)
	require.NoError(t, RunOccurrence(ctx, db, config, *next, now.Add(time.Minute)))
	job, err = PendingExport(ctx, db)
	require.NoError(t, err)
	require.NotNil(t, job)
	require.NoError(t, RunOccurrence(ctx, db, config, *next, now.Add(time.Minute)))
	duplicate, err := PendingExport(ctx, db)
	require.NoError(t, err)
	require.Equal(t, job, duplicate)
}
