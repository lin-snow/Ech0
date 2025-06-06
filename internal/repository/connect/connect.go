package repository

import (
	"github.com/lin-snow/ech0/internal/database"
	model "github.com/lin-snow/ech0/internal/model/connect"
	"gorm.io/gorm"
)

type ConnectRepository struct {
	db *gorm.DB
}

func NewConnectRepository(db *gorm.DB) ConnectRepositoryInterface {
	return &ConnectRepository{
		db: db,
	}
}

func (c ConnectRepository) GetAllConnects() ([]model.Connected, error) {
	var connects []model.Connected
	// 查询数据库
	if err := database.DB.Find(&connects).Error; err != nil {
		return nil, err
	}
	// 如果没有找到，返回空切片
	if len(connects) == 0 {
		return []model.Connected{}, nil
	}
	// 返回查询到的 connects
	return connects, nil
}

func (c ConnectRepository) CreateConnect(connect *model.Connected) error {
	if err := database.DB.Create(connect).Error; err != nil {
		return err
	}
	return nil
}

func (c ConnectRepository) DeleteConnect(id uint) error {
	// 根据 ID 删除 Connect
	if err := database.DB.Delete(&model.Connected{}, id).Error; err != nil {
		return err
	}

	return nil
}
