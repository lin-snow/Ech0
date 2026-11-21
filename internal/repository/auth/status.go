// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"

	model "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/gorm"
)

// IsAccountActive answers whether the token owner may still act. It is read on
// every authenticated request so disabling an account invalidates tokens that
// were already issued; a missing row counts as inactive.
func (authRepository *AuthRepository) IsAccountActive(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, nil
	}
	var user model.User
	err := authRepository.getDB(ctx).Select("id", "status").Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return user.IsActive(), nil
}
