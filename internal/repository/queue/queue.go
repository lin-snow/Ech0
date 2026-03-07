package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/queue"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type QueueRepository struct {
	db func() *gorm.DB
}

func NewQueueRepository(db func() *gorm.DB) QueueRepositoryInterface {
	return &QueueRepository{db: db}
}

func (queueRepository *QueueRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return queueRepository.db()
}

// SaveDeadLetter 保存死信任务
func (queueRepository *QueueRepository) SaveDeadLetter(
	ctx context.Context,
	deadLetter *model.DeadLetter,
) error {
	return queueRepository.getDB(ctx).Create(deadLetter).Error
}

// DeleteDeadLetter 删除死信任务
func (queueRepository *QueueRepository) DeleteDeadLetter(ctx context.Context, id int64) error {
	return queueRepository.getDB(ctx).Delete(&model.DeadLetter{}, id).Error
}

// ListDeadLetters 列出所有死信任务
func (queueRepository *QueueRepository) ListDeadLetters(ctx context.Context, limit int) ([]model.DeadLetter, error) {
	var deadLetters []model.DeadLetter
	err := queueRepository.getDB(ctx).Limit(limit).Find(&deadLetters).Error
	if err != nil {
		return []model.DeadLetter{}, err
	}
	return deadLetters, nil
}

// UpdateDeadLetter 更新死信任务
func (queueRepository *QueueRepository) UpdateDeadLetter(
	ctx context.Context,
	deadLetter *model.DeadLetter,
) error {
	return queueRepository.getDB(ctx).Save(deadLetter).Error
}
