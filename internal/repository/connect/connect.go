package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/connect"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type ConnectRepository struct {
	db func() *gorm.DB
}

func NewConnectRepository(dbProvider func() *gorm.DB) ConnectRepositoryInterface {
	return &ConnectRepository{
		db: dbProvider,
	}
}

// getDB 从上下文中获取事务
func (connectRepository *ConnectRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return connectRepository.db()
}

// GetAllConnects 获取所有连接
func (connectRepository *ConnectRepository) GetAllConnects(ctx context.Context) ([]model.Connected, error) {
	var connects []model.Connected
	// 查询数据库
	if err := connectRepository.getDB(ctx).Find(&connects).Error; err != nil {
		return nil, err
	}
	// 如果没有找到，返回空切片
	if len(connects) == 0 {
		return []model.Connected{}, nil
	}
	// 返回查询到的 connects
	return connects, nil
}

// CreateConnect 创建一个新的连接
func (connectRepository *ConnectRepository) CreateConnect(
	ctx context.Context,
	connect *model.Connected,
) error {
	if err := connectRepository.getDB(ctx).Create(connect).Error; err != nil {
		return err
	}
	return nil
}

// DeleteConnect 删除连接
func (connectRepository *ConnectRepository) DeleteConnect(ctx context.Context, id uint) error {
	// 根据 ID 删除 Connect
	if err := connectRepository.getDB(ctx).Delete(&model.Connected{}, id).Error; err != nil {
		return err
	}

	return nil
}
