// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository_test

import (
	"context"
	"errors"
	"testing"

	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func closedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:closedvisitor?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return db
}

func newVisitorRepo(t *testing.T) (*visitorRepository.VisitorRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return visitorRepository.NewVisitorRepository(func() *gorm.DB { return db }), db
}

func TestVisitorRepository_UpsertDailyStat_Insert(t *testing.T) {
	repo, db := newVisitorRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.UpsertDailyStat(ctx, visitorModel.DailyStat{Date: "2026-06-30", PV: 12, UV: 5}))

	var got visitorModel.DailyStat
	require.NoError(t, db.First(&got, "date = ?", "2026-06-30").Error)
	assert.Equal(t, int64(12), got.PV)
	assert.Equal(t, int64(5), got.UV)
}

func TestVisitorRepository_UpsertDailyStat_OverwritesOnConflict(t *testing.T) {
	repo, db := newVisitorRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.UpsertDailyStat(ctx, visitorModel.DailyStat{Date: "2026-06-30", PV: 10, UV: 4}))
	require.NoError(t, repo.UpsertDailyStat(ctx, visitorModel.DailyStat{Date: "2026-06-30", PV: 25, UV: 9}))

	var got visitorModel.DailyStat
	require.NoError(t, db.First(&got, "date = ?", "2026-06-30").Error)
	assert.Equal(t, int64(25), got.PV, "冲突时 pv 应被覆盖为新值")
	assert.Equal(t, int64(9), got.UV, "冲突时 uv 应被覆盖为新值")

	var count int64
	require.NoError(t, db.Model(&visitorModel.DailyStat{}).Where("date = ?", "2026-06-30").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestVisitorRepository_GetRecentDays(t *testing.T) {
	repo, _ := newVisitorRepo(t)
	ctx := context.Background()

	seed := []visitorModel.DailyStat{
		{Date: "2026-06-27", PV: 1, UV: 1},
		{Date: "2026-06-28", PV: 2, UV: 2},
		{Date: "2026-06-29", PV: 3, UV: 3},
		{Date: "2026-06-30", PV: 4, UV: 4},
	}
	for _, s := range seed {
		require.NoError(t, repo.UpsertDailyStat(ctx, s))
	}

	cases := []struct {
		name      string
		days      int
		wantDates []string
	}{
		{"zero returns empty", 0, nil},
		{"negative returns empty", -3, nil},
		{"limit two most recent desc", 2, []string{"2026-06-30", "2026-06-29"}},
		{"limit exceeding count returns all desc", 10, []string{"2026-06-30", "2026-06-29", "2026-06-28", "2026-06-27"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetRecentDays(ctx, tc.days)
			require.NoError(t, err)
			require.NotNil(t, got, "即使为空也应返回非 nil 切片")
			gotDates := make([]string, 0, len(got))
			for _, s := range got {
				gotDates = append(gotDates, s.Date)
			}
			assert.Equal(t, tc.wantDates, nilIfEmpty(gotDates))
		})
	}
}

func TestVisitorRepository_DeleteOlderThan(t *testing.T) {
	repo, db := newVisitorRepo(t)
	ctx := context.Background()

	seed := []string{"2026-06-01", "2026-06-15", "2026-06-29", "2026-06-30"}
	for _, d := range seed {
		require.NoError(t, repo.UpsertDailyStat(ctx, visitorModel.DailyStat{Date: d, PV: 1, UV: 1}))
	}

	require.NoError(t, repo.DeleteOlderThan(ctx, "2026-06-29"))

	var remaining []visitorModel.DailyStat
	require.NoError(t, db.Order("date ASC").Find(&remaining).Error)
	gotDates := make([]string, 0, len(remaining))
	for _, s := range remaining {
		gotDates = append(gotDates, s.Date)
	}
	assert.Equal(t, []string{"2026-06-29", "2026-06-30"}, gotDates)

	t.Run("no match is a no-op", func(t *testing.T) {
		require.NoError(t, repo.DeleteOlderThan(ctx, "2000-01-01"))
		var count int64
		require.NoError(t, db.Model(&visitorModel.DailyStat{}).Count(&count).Error)
		assert.Equal(t, int64(2), count)
	})
}

func TestVisitorRepository_DBErrorsPropagate(t *testing.T) {
	repo := visitorRepository.NewVisitorRepository(func() *gorm.DB { return closedDB(t) })
	ctx := context.Background()

	t.Run("upsert", func(t *testing.T) {
		assert.Error(t, repo.UpsertDailyStat(ctx, visitorModel.DailyStat{Date: "2026-06-30", PV: 1, UV: 1}))
	})
	t.Run("get recent days", func(t *testing.T) {
		got, err := repo.GetRecentDays(ctx, 7)
		assert.Error(t, err)
		assert.Empty(t, got, "出错时返回空切片")
	})
	t.Run("delete older than", func(t *testing.T) {
		assert.Error(t, repo.DeleteOlderThan(ctx, "2026-06-30"))
	})
}

func TestVisitorRepository_TxContext(t *testing.T) {
	repo, db := newVisitorRepo(t)
	transactor := transaction.NewGormTransactor(func() *gorm.DB { return db })

	t.Run("commit persists", func(t *testing.T) {
		err := transactor.Run(context.Background(), func(ctx context.Context) error {
			return repo.UpsertDailyStat(ctx, visitorModel.DailyStat{Date: "2026-07-01", PV: 7, UV: 3})
		})
		require.NoError(t, err)
		got, err := repo.GetRecentDays(context.Background(), 1)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "2026-07-01", got[0].Date)
		assert.Equal(t, int64(7), got[0].PV)
	})

	t.Run("rollback discards", func(t *testing.T) {
		sentinel := errors.New("boom")
		err := transactor.Run(context.Background(), func(ctx context.Context) error {
			if err := repo.UpsertDailyStat(ctx, visitorModel.DailyStat{Date: "2026-07-02", PV: 1, UV: 1}); err != nil {
				return err
			}
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)

		var count int64
		require.NoError(t, db.Model(&visitorModel.DailyStat{}).Where("date = ?", "2026-07-02").Count(&count).Error)
		assert.Equal(t, int64(0), count, "回滚后该行不应落库")
	})
}

func nilIfEmpty(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	return s
}
