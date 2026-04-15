package repository

import (
	"context"

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
	return visitorRepository.db()
}

func (visitorRepository *VisitorRepository) UpsertDailyStat(
	ctx context.Context,
	stat visitorModel.DailyStat,
) error {
	return visitorRepository.getDB(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "date"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
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
