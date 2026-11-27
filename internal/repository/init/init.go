// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
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

// IsInitialized answers whether the space has been claimed, which is exactly
// "an owner exists". There is no separate install marker to keep in sync.
func (r *InitRepository) IsInitialized() (bool, error) {
	var count int64
	err := r.db().
		Model(&userModel.User{}).
		Where("role = ?", userModel.RoleOwner).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetOwner returns the account the space belongs to. A missing row means the
// instance is not initialized yet.
func (r *InitRepository) GetOwner() (userModel.User, error) {
	user := userModel.User{}
	if err := r.db().
		Where("role = ?", userModel.RoleOwner).
		First(&user).Error; err != nil {
		return userModel.User{}, err
	}
	return user, nil
}
