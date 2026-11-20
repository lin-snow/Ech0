// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

func (r *AuthRepository) CreateDevice(ctx context.Context, g *authModel.DeviceGrant) error {
	return r.getDB(ctx).WithContext(ctx).Create(g).Error
}

func (r *AuthRepository) DeviceInfo(ctx context.Context, userKey string) (authModel.DeviceGrant, error) {
	var g authModel.DeviceGrant
	err := r.getDB(ctx).WithContext(ctx).Where("user_key=?", userKey).Take(&g).Error
	return g, err
}

// MutateDevice serializes approval, polling and redemption. A true callback
// result consumes the grant in the same transaction as token issuance or lookup.
func (r *AuthRepository) MutateDevice(ctx context.Context, key string, byUser bool, fn func(context.Context, *authModel.DeviceGrant) (bool, error)) error {
	return r.getDB(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		match := "id=?"
		if byUser {
			match = "user_key=?"
		}
		locked := tx.Model(&authModel.DeviceGrant{}).Where(match, key).UpdateColumn("last_poll", gorm.Expr("last_poll"))
		if locked.Error != nil {
			return locked.Error
		}
		if locked.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		var grant authModel.DeviceGrant
		if err := tx.Where(match, key).Take(&grant).Error; err != nil {
			return err
		}
		consume, err := fn(context.WithValue(ctx, transaction.TxKey, tx), &grant)
		if err != nil {
			return err
		}
		if consume {
			return tx.Delete(&grant).Error
		}
		return tx.Model(&grant).Select("last_poll", "jwt_id").Updates(&grant).Error
	})
}
func (r *AuthRepository) GrantedToken(ctx context.Context, jti string) (string, error) {
	var token settingModel.AccessTokenSetting
	err := r.getDB(ctx).WithContext(ctx).Select("token").Where("jti=?", jti).Take(&token).Error
	return token.Token, err
}
