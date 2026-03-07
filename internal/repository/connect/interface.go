package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/connect"
)

type ConnectRepositoryInterface interface {
	// GetAllConnects 获取所有连接
	GetAllConnects(ctx context.Context) ([]model.Connected, error)

	// CreateConnect 创建一个新的连接
	CreateConnect(ctx context.Context, connect *model.Connected) error

	// DeleteConnect 删除连接
	DeleteConnect(ctx context.Context, id uint) error
}
