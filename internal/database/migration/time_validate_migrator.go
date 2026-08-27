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

const defaultInvalidSampleLimit = 5

type TimeValidateReport struct {
	StartedAt         time.Time
	FinishedAt        time.Time
	TotalInvalid      int64
	InvalidByColumn   []TimeMigrationStat
	InvalidSampleRows []TimeMigrationInvalidSample
}

type storageTimeValidateMigrator struct {
	sampleLimit int
}

func NewStorageTimeValidateMigrator() Migrator {
	return &storageTimeValidateMigrator{sampleLimit: defaultInvalidSampleLimit}
}

func (m *storageTimeValidateMigrator) Name() string {
	return "storage_time_validate_migrator"
}

func (m *storageTimeValidateMigrator) Key() string {
	return commonModel.StorageTimeValidatedKey
}

func (m *storageTimeValidateMigrator) CanRerun() bool {
	return false
}

func (m *storageTimeValidateMigrator) Migrate(db *gorm.DB) error {
	report, err := ValidateStorageTimesParseable(db, m.sampleLimit)
	if err != nil {
		return err
	}
	logUtil.Info(
		"storage time validate completed",
		slog.String("module", "database"),
		slog.String("migrator", m.Name()),
		slog.Int64("total_invalid_rows", report.TotalInvalid),
	)
	return nil
}

func ValidateStorageTimesParseable(db *gorm.DB, sampleLimit int) (*TimeValidateReport, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if sampleLimit <= 0 {
		sampleLimit = defaultInvalidSampleLimit
	}

	report := &TimeValidateReport{
		StartedAt:         time.Now().UTC(),
		InvalidByColumn:   make([]TimeMigrationStat, 0),
		InvalidSampleRows: make([]TimeMigrationInvalidSample, 0),
	}

	for _, plan := range StorageTimeColumnPlans() {
		stat := TimeMigrationStat{
			Table:  plan.Table,
			Column: plan.Column,
		}
		countSQL := fmt.Sprintf(
			"SELECT COUNT(1) FROM %s WHERE %s IS NOT NULL AND typeof(%s)='text' AND strftime('%%s', %s) IS NULL",
			plan.Table, plan.Column, plan.Column, plan.Column,
		)
		if err := db.Raw(countSQL).Scan(&stat.CandidateRow).Error; err != nil {
			return nil, fmt.Errorf("validate count %s.%s failed: %w", plan.Table, plan.Column, err)
		}
		if stat.CandidateRow == 0 {
			continue
		}
		report.TotalInvalid += stat.CandidateRow
		report.InvalidByColumn = append(report.InvalidByColumn, stat)

		sampleSQL := fmt.Sprintf(
			"SELECT rowid, %s AS value FROM %s WHERE %s IS NOT NULL AND typeof(%s)='text' AND strftime('%%s', %s) IS NULL LIMIT ?",
			plan.Column, plan.Table, plan.Column, plan.Column, plan.Column,
		)
		var rows []struct {
			RowID int64  `gorm:"column:rowid"`
			Value string `gorm:"column:value"`
		}
		if err := db.Raw(sampleSQL, sampleLimit).Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("validate sample %s.%s failed: %w", plan.Table, plan.Column, err)
		}
		for _, row := range rows {
			report.InvalidSampleRows = append(report.InvalidSampleRows, TimeMigrationInvalidSample{
				Table:  plan.Table,
				Column: plan.Column,
				RowID:  row.RowID,
				Value:  row.Value,
			})
		}
	}

	report.FinishedAt = time.Now().UTC()
	if report.TotalInvalid > 0 {
		return report, fmt.Errorf(
			"found %d invalid time values before unix migration: %s",
			report.TotalInvalid,
			formatInvalidSamples(report.InvalidSampleRows, sampleLimit),
		)
	}
	return report, nil
}

func formatInvalidSamples(samples []TimeMigrationInvalidSample, sampleLimit int) string {
	if len(samples) == 0 {
		return "no sample available"
	}
	if len(samples) > sampleLimit {
		samples = samples[:sampleLimit]
	}

	items := make([]string, 0, len(samples))
	for _, sample := range samples {
		items = append(items, fmt.Sprintf("%s.%s(rowid=%d,value=%q)", sample.Table, sample.Column, sample.RowID, sample.Value))
	}
	return strings.Join(items, "; ")
}
