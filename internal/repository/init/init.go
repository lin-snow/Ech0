// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	initService "github.com/lin-snow/ech0/internal/service/init"
	"gorm.io/gorm"
)

type InitRepository struct {
	db func() *gorm.DB
}

var _ initService.Repository = (*InitRepository)(nil)

func NewInitRepository(dbProvider func() *gorm.DB) *InitRepository {
	return &InitRepository{db: dbProvider}
}

func (r *InitRepository) IsInitialized() (bool, error) {
	var kv commonModel.KeyValue
	err := r.db().Where("key = ?", commonModel.InstallInitializedKey).First(&kv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return kv.Value == "true", nil
}

func (r *InitRepository) GetPrimaryAdmin() (userModel.User, error) {
	user := userModel.User{}
	if err := r.db().
		Where("is_admin = ?", true).
		Order("id ASC").
		First(&user).Error; err != nil {
		return userModel.User{}, err
	}
	return user, nil
}
