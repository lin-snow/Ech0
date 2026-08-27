// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"gorm.io/gorm"
)

const DefaultLegacySourceTimezone = "Asia/Shanghai"

type TimeNormalizationReport struct {
	TimezoneSource string
	StartedAt      time.Time
	FinishedAt     time.Time
	TotalCandidate int64
	TotalUpdated   int64
	Details        []TimeMigrationStat
}

type legacyTimeNormalizerMigrator struct {
	sourceTimezone string
}

func NewLegacyTimeNormalizerMigrator(sourceTimezone string) Migrator {
	return &legacyTimeNormalizerMigrator{sourceTimezone: strings.TrimSpace(sourceTimezone)}
}

func (m *legacyTimeNormalizerMigrator) Name() string {
	return "legacy_time_normalizer"
}

func (m *legacyTimeNormalizerMigrator) Key() string {
	return commonModel.StorageTimeUTCNormalizedKey
}

func (m *legacyTimeNormalizerMigrator) CanRerun() bool {
	return false
}

func (m *legacyTimeNormalizerMigrator) Migrate(db *gorm.DB) error {
	report, err := NormalizeLegacyStorageTimesToUTC(db, m.sourceTimezone)
	if err != nil {
		return err
	}

	logUtil.Info(
		"storage time normalization completed",
		slog.String("module", "database"),
		slog.String("migrator", m.Name()),
		slog.String("source_timezone", report.TimezoneSource),
		slog.Int64("total_candidate_rows", report.TotalCandidate),
		slog.Int64("total_updated_rows", report.TotalUpdated),
	)
	return nil
}

func NormalizeLegacyStorageTimesToUTC(db *gorm.DB, sourceTimezone string) (*TimeNormalizationReport, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	tzName := strings.TrimSpace(sourceTimezone)
	if tzName == "" {
		return nil, fmt.Errorf("source timezone is required")
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("load timezone %s failed: %w", tzName, err)
	}
	_, offsetSeconds := time.Date(2000, 1, 1, 0, 0, 0, 0, loc).Zone()
	if offsetSeconds == 0 {
		return nil, fmt.Errorf("timezone %s offset is zero, migration would be a no-op", tzName)
	}

	report := &TimeNormalizationReport{
		TimezoneSource: tzName,
		StartedAt:      time.Now().UTC(),
		Details:        make([]TimeMigrationStat, 0),
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, plan := range StorageTimeColumnPlans() {
			stat := TimeMigrationStat{
				Table:  plan.Table,
				Column: plan.Column,
			}

			countSQL := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s IS NOT NULL", plan.Table, plan.Column)
			if err := tx.Raw(countSQL).Scan(&stat.CandidateRow).Error; err != nil {
				return fmt.Errorf("count %s.%s failed: %w", plan.Table, plan.Column, err)
			}

			updateSQL := fmt.Sprintf(
				"UPDATE %s SET %s = strftime('%%Y-%%m-%%dT%%H:%%M:%%fZ', julianday(%s) - (?/86400.0)) WHERE %s IS NOT NULL",
				plan.Table,
				plan.Column,
				plan.Column,
				plan.Column,
			)
			result := tx.Exec(updateSQL, offsetSeconds)
			if result.Error != nil {
				return fmt.Errorf("update %s.%s failed: %w", plan.Table, plan.Column, result.Error)
			}
			stat.UpdatedRow = result.RowsAffected
			report.TotalCandidate += stat.CandidateRow
			report.TotalUpdated += stat.UpdatedRow
			report.Details = append(report.Details, stat)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	report.FinishedAt = time.Now().UTC()
	return report, nil
}
