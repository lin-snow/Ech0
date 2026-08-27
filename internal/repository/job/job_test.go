// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/job"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	jobRepository "github.com/lin-snow/ech0/internal/repository/job"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestRepo(t *testing.T) (*jobRepository.JobRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&jobModel.Job{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
	return jobRepository.NewJobRepository(func() *gorm.DB { return db }), db
}

func TestRepo_GetByType_NotFound(t *testing.T) {
	repo, _ := newTestRepo(t)
	if _, err := repo.GetByType(context.Background(), "reindex"); !errors.Is(err, job.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepo_UpsertOverwritesByType(t *testing.T) {
	repo, db := newTestRepo(t)
	ctx := context.Background()
	if err := repo.Upsert(ctx, &jobModel.Job{Type: "reindex", Status: jobModel.StatusPending}); err != nil {
		t.Fatalf("upsert pending failed: %v", err)
	}
	if err := repo.Upsert(ctx, &jobModel.Job{Type: "reindex", Status: jobModel.StatusSuccess, Payload: `{"indexed":3}`}); err != nil {
		t.Fatalf("upsert success failed: %v", err)
	}
	got, err := repo.GetByType(ctx, "reindex")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Status != jobModel.StatusSuccess || got.Payload != `{"indexed":3}` {
		t.Fatalf("expected single overwritten row, got %+v", got)
	}
	var count int64
	if err := db.Model(&jobModel.Job{}).Where("type = ?", "reindex").Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 row per type, got %d", count)
	}
}

func TestRepo_SweepRunning(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	_ = repo.Upsert(ctx, &jobModel.Job{Type: "reindex", Status: jobModel.StatusRunning})
	if err := repo.SweepRunning(ctx, "interrupted by restart"); err != nil {
		t.Fatalf("sweep failed: %v", err)
	}
	got, err := repo.GetByType(ctx, "reindex")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Status != jobModel.StatusFailed || got.Error != "interrupted by restart" || got.FinishedAt == nil {
		t.Fatalf("expected swept to failed with reason+finished, got %+v", got)
	}
}

func TestRepo_SweepRunning_LeavesTerminal(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	_ = repo.Upsert(ctx, &jobModel.Job{Type: "reindex", Status: jobModel.StatusSuccess})
	if err := repo.SweepRunning(ctx, "interrupted by restart"); err != nil {
		t.Fatalf("sweep failed: %v", err)
	}
	got, _ := repo.GetByType(ctx, "reindex")
	if got.Status != jobModel.StatusSuccess {
		t.Fatalf("sweep must not touch terminal rows, got %q", got.Status)
	}
}

func TestRepo_Delete(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()
	_ = repo.Upsert(ctx, &jobModel.Job{Type: "migration", Status: jobModel.StatusSuccess})
	if err := repo.Delete(ctx, "migration"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := repo.GetByType(ctx, "migration"); !errors.Is(err, job.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
