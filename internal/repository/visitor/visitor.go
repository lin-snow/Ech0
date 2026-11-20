// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"time"

	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VisitorRepository struct {
	db func() *gorm.DB
}

func NewVisitorRepository(dbProvider func() *gorm.DB) *VisitorRepository {
	return &VisitorRepository{
		db: dbProvider,
	}
}

func (visitorRepository *VisitorRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return visitorRepository.db().WithContext(ctx)
}

func (visitorRepository *VisitorRepository) UpsertDailyStat(
	ctx context.Context,
	stat visitorModel.DailyStat,
) error {
	return visitorRepository.getDB(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "date"}},
			DoUpdates: clause.Assignments(map[string]any{
				"pv": stat.PV,
				"uv": stat.UV,
			}),
		},
	).Create(&stat).Error
}

func (visitorRepository *VisitorRepository) GetRecentDays(
	ctx context.Context,
	days int,
) ([]visitorModel.DailyStat, error) {
	if days <= 0 {
		return []visitorModel.DailyStat{}, nil
	}
	var stats []visitorModel.DailyStat
	err := visitorRepository.getDB(ctx).
		Order("date DESC").
		Limit(days).
		Find(&stats).Error
	if err != nil {
		return []visitorModel.DailyStat{}, err
	}
	return stats, nil
}

func (visitorRepository *VisitorRepository) DeleteOlderThan(
	ctx context.Context,
	cutoffDate string,
) error {
	return visitorRepository.getDB(ctx).
		Where("date < ?", cutoffDate).
		Delete(&visitorModel.DailyStat{}).Error
}

// RecordVisit serializes the deduplication decision and the two increments.
// A view window continues across midnight. Unique visitors reset by UTC date.
func (r *VisitorRepository) RecordVisit(ctx context.Context, hash string, at time.Time, window time.Duration) error {
	day := at.UTC().Format(time.DateOnly)
	stamp := at.UnixMilli()
	return r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		blank := visitorModel.VisitorSeen{Hash: hash}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&blank).Error; err != nil {
			return err
		}
		var seen visitorModel.VisitorSeen
		if err := tx.Where("hash = ?", hash).First(&seen).Error; err != nil {
			return err
		}
		pv, uv := 0, 0
		if stamp-seen.LastPV >= window.Milliseconds() {
			pv = 1
			seen.LastPV = stamp
		}
		if day > seen.Day {
			uv = 1
			seen.Day = day
		}
		if pv == 0 && uv == 0 {
			return nil
		}
		if err := tx.Save(&seen).Error; err != nil {
			return err
		}
		return tx.Exec("INSERT INTO daily_stats(date,pv,uv) VALUES(?,?,?) ON CONFLICT(date) DO UPDATE SET pv=pv+excluded.pv,uv=uv+excluded.uv", day, pv, uv).Error
	})
}

func (r *VisitorRepository) DeleteSeenBefore(ctx context.Context, day string) error {
	return r.db().WithContext(ctx).Where("day < ?", day).Delete(&visitorModel.VisitorSeen{}).Error
}
