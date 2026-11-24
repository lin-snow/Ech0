// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"

	model "github.com/lin-snow/ech0/internal/model/user"
	"gorm.io/gorm"
)

// AccountState returns the live role and status of a token owner. It is read on
// every authenticated request so that a disable, a demotion or an ownership
// transfer invalidates tokens that were already issued; a missing row counts as
// absent, not as active.
func (authRepository *AuthRepository) AccountState(ctx context.Context, userID string) (model.AccountState, error) {
	if userID == "" {
		return model.AccountState{}, nil
	}
	var user model.User
	err := authRepository.getDB(ctx).
		Select("role", "status").
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.AccountState{}, nil
		}
		return model.AccountState{}, err
	}
	return model.AccountState{Found: true, Role: user.Role, Status: user.Status}, nil
}
