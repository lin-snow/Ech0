// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_UpsertLocalAuth(t *testing.T) {
	repo, db, _ := newUserRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.UpsertLocalAuth(ctx, &userModel.UserLocalAuth{
		UserID: "u1", PasswordHash: "hash-md5", PasswordAlgo: "md5",
	}))
	var row userModel.UserLocalAuth
	require.NoError(t, db.Where("user_id = ?", "u1").First(&row).Error)
	assert.Equal(t, "hash-md5", row.PasswordHash)
	assert.Equal(t, "md5", row.PasswordAlgo)
	assert.NotZero(t, row.UpdatedAt, "autoUpdateTime 应被填充")

	require.NoError(t, repo.UpsertLocalAuth(ctx, &userModel.UserLocalAuth{
		UserID: "u1", PasswordHash: "hash-bcrypt", PasswordAlgo: "bcrypt",
	}))
	var count int64
	require.NoError(t, db.Model(&userModel.UserLocalAuth{}).Where("user_id = ?", "u1").Count(&count).Error)
	assert.Equal(t, int64(1), count, "冲突应更新而非新增")
	require.NoError(t, db.Where("user_id = ?", "u1").First(&row).Error)
	assert.Equal(t, "hash-bcrypt", row.PasswordHash)
	assert.Equal(t, "bcrypt", row.PasswordAlgo)
}

func TestUserRepository_DeleteUser_RemovesLocalAuth(t *testing.T) {
	repo, db, _ := newUserRepo(t)
	ctx := context.Background()

	seedUser(t, db, userModel.User{ID: "u1", Username: "alice"})
	require.NoError(t, repo.UpsertLocalAuth(ctx, &userModel.UserLocalAuth{
		UserID: "u1", PasswordHash: "h", PasswordAlgo: "bcrypt",
	}))

	require.NoError(t, repo.DeleteUser(ctx, "u1"))

	var count int64
	require.NoError(t, db.Model(&userModel.UserLocalAuth{}).Where("user_id = ?", "u1").Count(&count).Error)
	assert.Equal(t, int64(0), count, "删除用户应一并清理其本地认证行")
}
