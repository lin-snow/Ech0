// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newConnectRepo(t *testing.T) (*ConnectRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return NewConnectRepository(func() *gorm.DB { return db }), db
}

func TestConnectRepository_GetAllConnects(t *testing.T) {
	t.Run("empty returns non-nil empty slice", func(t *testing.T) {
		repo, _ := newConnectRepo(t)

		got, err := repo.GetAllConnects(context.Background())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("returns all inserted rows", func(t *testing.T) {
		repo, db := newConnectRepo(t)

		seed := []connectModel.Connected{
			{ID: "c-1", ConnectURL: "https://a.example.com"},
			{ID: "c-2", ConnectURL: "https://b.example.com"},
			{ID: "c-3", ConnectURL: "https://c.example.com"},
		}
		require.NoError(t, db.Create(&seed).Error)

		got, err := repo.GetAllConnects(context.Background())
		require.NoError(t, err)
		require.Len(t, got, 3)

		urls := make(map[string]string, len(got))
		for _, c := range got {
			urls[c.ID] = c.ConnectURL
		}
		assert.Equal(t, "https://a.example.com", urls["c-1"])
		assert.Equal(t, "https://b.example.com", urls["c-2"])
		assert.Equal(t, "https://c.example.com", urls["c-3"])
	})
}

func TestConnectRepository_CreateConnect(t *testing.T) {
	t.Run("persists row and keeps provided id", func(t *testing.T) {
		repo, db := newConnectRepo(t)

		c := &connectModel.Connected{ID: "fixed-id", ConnectURL: "https://example.com"}
		require.NoError(t, repo.CreateConnect(context.Background(), c))

		var got connectModel.Connected
		require.NoError(t, db.First(&got, "id = ?", "fixed-id").Error)
		assert.Equal(t, "https://example.com", got.ConnectURL)
	})

	t.Run("BeforeCreate fills uuid when id empty", func(t *testing.T) {
		repo, db := newConnectRepo(t)

		c := &connectModel.Connected{ConnectURL: "https://generated.example.com"}
		require.NoError(t, repo.CreateConnect(context.Background(), c))
		require.NotEmpty(t, c.ID, "BeforeCreate should populate a uuid")

		var got connectModel.Connected
		require.NoError(t, db.First(&got, "id = ?", c.ID).Error)
		assert.Equal(t, "https://generated.example.com", got.ConnectURL)
	})
}

func TestConnectRepository_DeleteConnect(t *testing.T) {
	t.Run("removes existing row", func(t *testing.T) {
		repo, db := newConnectRepo(t)
		require.NoError(t, db.Create(&connectModel.Connected{ID: "del-1", ConnectURL: "https://x"}).Error)

		require.NoError(t, repo.DeleteConnect(context.Background(), "del-1"))

		var count int64
		require.NoError(t, db.Model(&connectModel.Connected{}).Where("id = ?", "del-1").Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("deleting a missing id is a no-op without error", func(t *testing.T) {
		repo, db := newConnectRepo(t)
		require.NoError(t, db.Create(&connectModel.Connected{ID: "keep", ConnectURL: "https://k"}).Error)

		require.NoError(t, repo.DeleteConnect(context.Background(), "does-not-exist"))

		var count int64
		require.NoError(t, db.Model(&connectModel.Connected{}).Count(&count).Error)
		assert.Equal(t, int64(1), count, "unrelated rows must remain")
	})
}

func TestConnectRepository_GetDB_UsesContextTx(t *testing.T) {
	repo, db := newConnectRepo(t)

	tx := db.Begin()
	require.NoError(t, tx.Error)
	ctxTx := context.WithValue(context.Background(), transaction.TxKey, tx)

	require.NoError(t, repo.CreateConnect(ctxTx, &connectModel.Connected{ID: "tx-row", ConnectURL: "https://tx"}))
	require.NoError(t, tx.Rollback().Error)

	got, err := repo.GetAllConnects(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got, "rolled-back tx write must not be visible on the base connection")
}
