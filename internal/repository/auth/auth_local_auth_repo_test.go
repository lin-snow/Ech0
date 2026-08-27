// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuthRepository_GetLocalAuthByUserID(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	ctx := context.Background()

	require.NoError(t, db.Create(&userModel.UserLocalAuth{
		UserID: "u1", PasswordHash: "hash", PasswordAlgo: "md5",
	}).Error)

	got, err := repo.GetLocalAuthByUserID(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "hash", got.PasswordHash)
	assert.Equal(t, "md5", got.PasswordAlgo)

	_, err = repo.GetLocalAuthByUserID(ctx, "ghost")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestAuthRepository_UpdateLocalAuthPassword(t *testing.T) {
	repo, db, _ := newAuthRepo(t)
	ctx := context.Background()

	require.NoError(t, db.Create(&userModel.UserLocalAuth{
		UserID: "u1", PasswordHash: "old-md5", PasswordAlgo: "md5",
	}).Error)

	require.NoError(t, repo.UpdateLocalAuthPassword(ctx, "u1", "new-bcrypt", "bcrypt"))

	var row userModel.UserLocalAuth
	require.NoError(t, db.Where("user_id = ?", "u1").First(&row).Error)
	assert.Equal(t, "new-bcrypt", row.PasswordHash)
	assert.Equal(t, "bcrypt", row.PasswordAlgo)
}
