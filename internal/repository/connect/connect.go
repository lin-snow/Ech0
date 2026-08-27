// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/connect"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type ConnectRepository struct {
	db func() *gorm.DB
}

var _ connectService.Repository = (*ConnectRepository)(nil)

func NewConnectRepository(dbProvider func() *gorm.DB) *ConnectRepository {
	return &ConnectRepository{
		db: dbProvider,
	}
}

func (connectRepository *ConnectRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return connectRepository.db()
}

func (connectRepository *ConnectRepository) GetAllConnects(ctx context.Context) ([]model.Connected, error) {
	var connects []model.Connected
	if err := connectRepository.getDB(ctx).Find(&connects).Error; err != nil {
		return nil, err
	}
	if len(connects) == 0 {
		return []model.Connected{}, nil
	}
	return connects, nil
}

func (connectRepository *ConnectRepository) CreateConnect(
	ctx context.Context,
	connect *model.Connected,
) error {
	if err := connectRepository.getDB(ctx).Create(connect).Error; err != nil {
		return err
	}
	return nil
}

func (connectRepository *ConnectRepository) DeleteConnect(ctx context.Context, id string) error {
	if err := connectRepository.getDB(ctx).Where("id = ?", id).Delete(&model.Connected{}).Error; err != nil {
		return err
	}

	return nil
}
