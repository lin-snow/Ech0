// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/user"
)

func (userRepository *UserRepository) UpdateUserStatus(ctx context.Context, id string, status model.Status) error {
	existing, err := userRepository.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	if err := userRepository.getDB(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return err
	}

	existing.Status = status
	userRepository.cache.Set(GetUserIDKey(existing.ID), existing, 1)
	userRepository.cache.Set(GetUsernameKey(existing.Username), existing, 1)
	return nil
}
