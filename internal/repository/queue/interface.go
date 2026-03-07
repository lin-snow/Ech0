package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/queue"
)

// QueueRepositoryInterface 队列仓储接口
type QueueRepositoryInterface interface {
	// SaveDeadLetter 保存死信任务
	SaveDeadLetter(ctx context.Context, deadLetter *model.DeadLetter) error

	// DeleteDeadLetter 删除死信任务
	DeleteDeadLetter(ctx context.Context, id int64) error

	// ListDeadLetters 列出所有可重试的死信任务
	ListDeadLetters(ctx context.Context, limit int) ([]model.DeadLetter, error)

	// UpdateDeadLetter 更新死信任务
	UpdateDeadLetter(ctx context.Context, deadLetter *model.DeadLetter) error
}
