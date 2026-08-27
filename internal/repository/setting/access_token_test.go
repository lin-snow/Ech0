// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"
	"testing"

	model "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newSettingRepo(t *testing.T) (*SettingRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return NewSettingRepository(func() *gorm.DB { return db }), db
}

func newAccessToken(id, userID, suffix string) *model.AccessTokenSetting {
	return &model.AccessTokenSetting{
		ID:        id,
		UserID:    userID,
		Token:     "tok-" + suffix,
		Name:      "name-" + suffix,
		TokenType: "access",
		Scopes:    `["echo:read"]`,
		Audience:  "api",
		JTI:       "jti-" + suffix,
	}
}

func TestSettingRepository_CreateAndGetAccessTokenByID(t *testing.T) {
	repo, _ := newSettingRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateAccessToken(ctx, newAccessToken("at-1", "u1", "1")))

	t.Run("hit returns the stored token", func(t *testing.T) {
		got, err := repo.GetAccessTokenByID(ctx, "at-1")
		require.NoError(t, err)
		assert.Equal(t, "at-1", got.ID)
		assert.Equal(t, "u1", got.UserID)
		assert.Equal(t, "jti-1", got.JTI)
		assert.NotZero(t, got.CreatedAt, "autoCreateTime 应填充")
	})

	t.Run("miss returns record-not-found", func(t *testing.T) {
		_, err := repo.GetAccessTokenByID(ctx, "nope")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}

func TestSettingRepository_CreateAccessToken_GeneratesIDWhenEmpty(t *testing.T) {
	repo, _ := newSettingRepo(t)
	ctx := context.Background()

	tok := newAccessToken("", "u1", "auto")
	require.NoError(t, repo.CreateAccessToken(ctx, tok))
	assert.NotEmpty(t, tok.ID)
}

func TestSettingRepository_ListAccessTokens_FiltersByUser(t *testing.T) {
	repo, _ := newSettingRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateAccessToken(ctx, newAccessToken("a1", "u1", "a1")))
	require.NoError(t, repo.CreateAccessToken(ctx, newAccessToken("a2", "u1", "a2")))
	require.NoError(t, repo.CreateAccessToken(ctx, newAccessToken("b1", "u2", "b1")))

	t.Run("returns only the owner's tokens", func(t *testing.T) {
		tokens, err := repo.ListAccessTokens(ctx, "u1")
		require.NoError(t, err)
		require.Len(t, tokens, 2)
		ids := []string{tokens[0].ID, tokens[1].ID}
		assert.ElementsMatch(t, []string{"a1", "a2"}, ids)
	})

	t.Run("user with no tokens returns empty slice", func(t *testing.T) {
		tokens, err := repo.ListAccessTokens(ctx, "ghost")
		require.NoError(t, err)
		assert.Empty(t, tokens)
	})
}

func TestSettingRepository_DeleteAccessTokenByID(t *testing.T) {
	repo, _ := newSettingRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.CreateAccessToken(ctx, newAccessToken("del-1", "u1", "del1")))
	require.NoError(t, repo.CreateAccessToken(ctx, newAccessToken("keep-1", "u1", "keep1")))

	require.NoError(t, repo.DeleteAccessTokenByID(ctx, "del-1"))

	t.Run("deleted token is gone", func(t *testing.T) {
		_, err := repo.GetAccessTokenByID(ctx, "del-1")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("sibling token survives", func(t *testing.T) {
		got, err := repo.GetAccessTokenByID(ctx, "keep-1")
		require.NoError(t, err)
		assert.Equal(t, "keep-1", got.ID)
	})

	t.Run("deleting a missing id is a no-op error-free", func(t *testing.T) {
		require.NoError(t, repo.DeleteAccessTokenByID(ctx, "nonexistent"))
	})
}
