// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"
	"time"

	"github.com/lin-snow/ech0/internal/job"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JobRepository struct {
	db func() *gorm.DB
}

var _ job.JobRepository = (*JobRepository)(nil)

func NewJobRepository(dbProvider func() *gorm.DB) *JobRepository {
	return &JobRepository{db: dbProvider}
}

func (r *JobRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *JobRepository) Upsert(ctx context.Context, j *jobModel.Job) error {
	return r.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}},
		UpdateAll: true,
	}).Create(j).Error
}

func (r *JobRepository) GetByType(ctx context.Context, jobType string) (jobModel.Job, error) {
	var j jobModel.Job
	err := r.getDB(ctx).Where("type = ?", jobType).First(&j).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return jobModel.Job{}, job.ErrNotFound
	}
	return j, err
}

func (r *JobRepository) SweepRunning(ctx context.Context, reason string) error {
	now := time.Now().UTC().Unix()
	return r.getDB(ctx).Model(&jobModel.Job{}).
		Where("status IN ?", []jobModel.Status{jobModel.StatusPending, jobModel.StatusRunning}).
		Updates(map[string]any{
			"status":      jobModel.StatusFailed,
			"error":       reason,
			"finished_at": now,
		}).Error
}

func (r *JobRepository) Delete(ctx context.Context, jobType string) error {
	return r.getDB(ctx).Where("type = ?", jobType).Delete(&jobModel.Job{}).Error
}
