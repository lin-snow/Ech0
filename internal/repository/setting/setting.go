// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/setting"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type SettingRepository struct {
	db func() *gorm.DB
}

var _ settingService.SettingRepository = (*SettingRepository)(nil)

func NewSettingRepository(dbProvider func() *gorm.DB) *SettingRepository {
	return &SettingRepository{
		db: dbProvider,
	}
}

func (settingRepository *SettingRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return settingRepository.db()
}

func (settingRepository *SettingRepository) ListAccessTokens(
	ctx context.Context,
	userID string,
) ([]model.AccessTokenSetting, error) {
	var tokens []model.AccessTokenSetting
	if err := settingRepository.getDB(ctx).Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

func (settingRepository *SettingRepository) CreateAccessToken(
	ctx context.Context,
	token *model.AccessTokenSetting,
) error {
	db := settingRepository.getDB(ctx)
	return db.Create(token).Error
}

func (settingRepository *SettingRepository) GetAccessTokenByID(
	ctx context.Context,
	id string,
) (model.AccessTokenSetting, error) {
	var token model.AccessTokenSetting
	if err := settingRepository.getDB(ctx).Where("id = ?", id).First(&token).Error; err != nil {
		return model.AccessTokenSetting{}, err
	}
	return token, nil
}

func (settingRepository *SettingRepository) DeleteAccessTokenByID(
	ctx context.Context,
	id string,
) error {
	db := settingRepository.getDB(ctx)
	return db.Where("id = ?", id).Delete(&model.AccessTokenSetting{}).Error
}
