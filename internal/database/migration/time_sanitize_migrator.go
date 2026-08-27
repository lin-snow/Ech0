// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"
	"log/slog"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"gorm.io/gorm"
)

type TimeSanitizeReport struct {
	StartedAt      time.Time
	FinishedAt     time.Time
	TotalCandidate int64
	TotalUpdated   int64
	Details        []TimeMigrationStat
}

type storageTimeSanitizeMigrator struct{}

func NewStorageTimeSanitizeMigrator() Migrator {
	return &storageTimeSanitizeMigrator{}
}

func (m *storageTimeSanitizeMigrator) Name() string {
	return "storage_time_sanitize_migrator"
}

func (m *storageTimeSanitizeMigrator) Key() string {
	return commonModel.StorageTimeSanitizedKey
}

func (m *storageTimeSanitizeMigrator) CanRerun() bool {
	return false
}

func (m *storageTimeSanitizeMigrator) Migrate(db *gorm.DB) error {
	report, err := SanitizeLegacyStorageTimes(db)
	if err != nil {
		return err
	}

	logUtil.Info(
		"storage time sanitize completed",
		slog.String("module", "database"),
		slog.String("migrator", m.Name()),
		slog.Int64("total_candidate_rows", report.TotalCandidate),
		slog.Int64("total_updated_rows", report.TotalUpdated),
	)
	return nil
}

func SanitizeLegacyStorageTimes(db *gorm.DB) (*TimeSanitizeReport, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	report := &TimeSanitizeReport{
		StartedAt: time.Now().UTC(),
		Details:   make([]TimeMigrationStat, 0),
	}

	return report, db.Transaction(func(tx *gorm.DB) error {
		for _, plan := range StorageTimeColumnPlans() {
			stat := TimeMigrationStat{
				Table:  plan.Table,
				Column: plan.Column,
			}

			countSQL := fmt.Sprintf(
				"SELECT COUNT(1) FROM %s WHERE %s IS NOT NULL AND typeof(%s)='text' AND strftime('%%s', %s) IS NULL",
				plan.Table, plan.Column, plan.Column, plan.Column,
			)
			if err := tx.Raw(countSQL).Scan(&stat.CandidateRow).Error; err != nil {
				return fmt.Errorf("count sanitize candidates for %s.%s failed: %w", plan.Table, plan.Column, err)
			}

			normalizedExpr := buildSafeTimeNormalizeExpr(plan.Column)
			updateSQL := fmt.Sprintf(
				`UPDATE %s
SET %s = %s
WHERE %s IS NOT NULL
  AND typeof(%s)='text'
  AND strftime('%%s', %s) IS NULL
  AND %s <> %s
  AND strftime('%%s', %s) IS NOT NULL`,
				plan.Table,
				plan.Column,
				normalizedExpr,
				plan.Column,
				plan.Column,
				plan.Column,
				plan.Column,
				normalizedExpr,
				normalizedExpr,
			)
			result := tx.Exec(updateSQL)
			if result.Error != nil {
				return fmt.Errorf("sanitize %s.%s failed: %w", plan.Table, plan.Column, result.Error)
			}

			stat.UpdatedRow = result.RowsAffected
			report.TotalCandidate += stat.CandidateRow
			report.TotalUpdated += stat.UpdatedRow
			report.Details = append(report.Details, stat)
		}

		report.FinishedAt = time.Now().UTC()
		return nil
	})
}

func buildSafeTimeNormalizeExpr(column string) string {
	trimmed := fmt.Sprintf("trim(%s)", column)
	spaceFixed := fmt.Sprintf("replace(replace(replace(%s, 'T', ' '), 't', ' '), '/', '-')", trimmed)
	collapsedSpace := fmt.Sprintf("replace(replace(%s, '  ', ' '), '  ', ' ')", spaceFixed)
	return fmt.Sprintf(
		"CASE WHEN upper(substr(%s, -1, 1))='Z' THEN substr(%s, 1, length(%s)-1) ELSE %s END",
		collapsedSpace, collapsedSpace, collapsedSpace, collapsedSpace,
	)
}
