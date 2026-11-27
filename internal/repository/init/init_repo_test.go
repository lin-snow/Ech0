// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"errors"
	"testing"

	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newInitRepo(t *testing.T) (*InitRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return NewInitRepository(func() *gorm.DB { return db }), db
}

func TestInitRepository_IsInitialized(t *testing.T) {
	t.Run("no owner is not initialized", func(t *testing.T) {
		repo, db := newInitRepo(t)
		require.NoError(t, db.Create(&userModel.User{ID: "u-member", Username: "member"}).Error)

		got, err := repo.IsInitialized()
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("owner presence is initialization", func(t *testing.T) {
		repo, db := newInitRepo(t)
		require.NoError(t, db.Create(&userModel.User{
			ID: "u-owner", Username: "owner", Role: userModel.RoleOwner,
		}).Error)

		got, err := repo.IsInitialized()
		require.NoError(t, err)
		assert.True(t, got)
	})
}

func TestInitRepository_GetOwner(t *testing.T) {
	repo, db := newInitRepo(t)

	t.Run("no owner returns record-not-found", func(t *testing.T) {
		_, err := repo.GetOwner()
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("returns the owner ignoring regular users", func(t *testing.T) {
		require.NoError(t, db.Create(&userModel.User{ID: "u-normal", Username: "normal"}).Error)
		require.NoError(t, db.Create(&userModel.User{
			ID:       "u-owner",
			Username: "owner",
			Role:     userModel.RoleOwner,
		}).Error)

		owner, err := repo.GetOwner()
		require.NoError(t, err)
		assert.Equal(t, "u-owner", owner.ID)
		assert.Equal(t, userModel.RoleOwner, owner.Role)
	})
}
