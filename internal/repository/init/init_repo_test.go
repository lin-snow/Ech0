// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
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
	cases := []struct {
		name  string
		seed  bool
		value string
		want  bool
	}{
		{name: "missing key is treated as not initialized", seed: false, want: false},
		{name: "value 'true' is initialized", seed: true, value: "true", want: true},
		{name: "any other value is not initialized", seed: true, value: "false", want: false},
		{name: "garbage value is not initialized", seed: true, value: "yes", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo, db := newInitRepo(t)
			if tc.seed {
				require.NoError(t, db.Create(&commonModel.KeyValue{
					Key:   commonModel.InstallInitializedKey,
					Value: tc.value,
				}).Error)
			}

			got, err := repo.IsInitialized()
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInitRepository_GetOwner(t *testing.T) {
	repo, db := newInitRepo(t)

	t.Run("no owner returns record-not-found", func(t *testing.T) {
		_, err := repo.GetOwner()
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})

	t.Run("returns the owner ignoring non-owner users", func(t *testing.T) {
		require.NoError(t, db.Create(&userModel.User{ID: "u-normal", Username: "normal"}).Error)
		require.NoError(t, db.Create(&userModel.User{
			ID:       "u-owner",
			Username: "owner",
			IsOwner:  true,
			IsAdmin:  true,
		}).Error)

		owner, err := repo.GetOwner()
		require.NoError(t, err)
		assert.Equal(t, "u-owner", owner.ID)
		assert.True(t, owner.IsOwner)
	})
}
