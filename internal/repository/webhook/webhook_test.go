// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository_test

import (
	"context"
	"errors"
	"testing"

	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	webhookRepository "github.com/lin-snow/ech0/internal/repository/webhook"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func closedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:closed?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return db
}

func newWebhookRepo(t *testing.T) (*webhookRepository.WebhookRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return webhookRepository.NewWebhookRepository(func() *gorm.DB { return db }), db
}

func makeWebhook(t *testing.T, repo *webhookRepository.WebhookRepository, name string) string {
	t.Helper()
	wh := &webhookModel.Webhook{Name: name, URL: "https://example.com/" + name, Secret: "s-" + name}
	require.NoError(t, repo.CreateWebhook(context.Background(), wh))
	require.NotEmpty(t, wh.ID, "BeforeCreate 应自动生成 ID")
	return wh.ID
}

func TestWebhookRepository_CreateAndGet(t *testing.T) {
	repo, db := newWebhookRepo(t)
	ctx := context.Background()

	id := makeWebhook(t, repo, "alpha")

	got, err := repo.GetWebhookByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, "alpha", got.Name)
	assert.True(t, got.IsActive, "model default:true 应使新建 webhook 默认激活")

	var count int64
	require.NoError(t, db.Model(&webhookModel.Webhook{}).Where("id = ?", id).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	all, err := repo.GetAllWebhooks(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestWebhookRepository_GetWebhookByID_NotFound(t *testing.T) {
	repo, _ := newWebhookRepo(t)
	_, err := repo.GetWebhookByID(context.Background(), "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestWebhookRepository_ListActiveWebhooks(t *testing.T) {
	repo, _ := newWebhookRepo(t)
	ctx := context.Background()

	t.Run("empty db returns empty slice", func(t *testing.T) {
		got, err := repo.ListActiveWebhooks(ctx)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	active1 := makeWebhook(t, repo, "active1")
	active2 := makeWebhook(t, repo, "active2")
	inactive := makeWebhook(t, repo, "inactive")
	require.NoError(t, repo.UpdateWebhookByID(ctx, inactive, &webhookModel.Webhook{
		Name: "inactive", URL: "https://example.com/inactive", Secret: "s", IsActive: false,
	}))

	t.Run("returns only active webhooks", func(t *testing.T) {
		got, err := repo.ListActiveWebhooks(ctx)
		require.NoError(t, err)
		require.Len(t, got, 2)
		ids := map[string]bool{}
		for _, w := range got {
			ids[w.ID] = true
			assert.True(t, w.IsActive)
		}
		assert.True(t, ids[active1])
		assert.True(t, ids[active2])
		assert.False(t, ids[inactive], "禁用的 webhook 不应出现")
	})
}

func TestWebhookRepository_UpdateWebhookByID(t *testing.T) {
	repo, _ := newWebhookRepo(t)
	ctx := context.Background()

	t.Run("updates mutable fields", func(t *testing.T) {
		id := makeWebhook(t, repo, "before")
		err := repo.UpdateWebhookByID(ctx, id, &webhookModel.Webhook{
			Name: "after", URL: "https://new.example.com", Secret: "new-secret", IsActive: false,
		})
		require.NoError(t, err)

		got, err := repo.GetWebhookByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "after", got.Name)
		assert.Equal(t, "https://new.example.com", got.URL)
		assert.Equal(t, "new-secret", got.Secret)
		assert.False(t, got.IsActive)
	})

	t.Run("not found returns error", func(t *testing.T) {
		err := repo.UpdateWebhookByID(ctx, "missing", &webhookModel.Webhook{Name: "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "webhook not found")
	})
}

func TestWebhookRepository_UpdateWebhookDeliveryStatus(t *testing.T) {
	repo, db := newWebhookRepo(t)
	ctx := context.Background()

	t.Run("writes last status and trigger", func(t *testing.T) {
		id := makeWebhook(t, repo, "deliver")
		require.NoError(t, repo.UpdateWebhookDeliveryStatus(ctx, id, "success", 1717000000))

		got, err := repo.GetWebhookByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "success", got.LastStatus)
		assert.Equal(t, int64(1717000000), got.LastTrigger)

		require.NoError(t, repo.UpdateWebhookDeliveryStatus(ctx, id, "failed", 1717000999))
		got, err = repo.GetWebhookByID(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, "failed", got.LastStatus)
		assert.Equal(t, int64(1717000999), got.LastTrigger)
	})

	t.Run("does not touch other rows", func(t *testing.T) {
		other := makeWebhook(t, repo, "untouched")
		target := makeWebhook(t, repo, "target")
		require.NoError(t, repo.UpdateWebhookDeliveryStatus(ctx, target, "success", 42))

		got, err := repo.GetWebhookByID(ctx, other)
		require.NoError(t, err)
		assert.Empty(t, got.LastStatus, "未指定的 webhook 状态不应被改动")
		assert.Equal(t, int64(0), got.LastTrigger)
	})

	t.Run("nonexistent id is a no-op without error", func(t *testing.T) {
		err := repo.UpdateWebhookDeliveryStatus(ctx, "missing", "success", 99)
		require.NoError(t, err)

		var count int64
		require.NoError(t, db.Model(&webhookModel.Webhook{}).Where("id = ?", "missing").Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})
}

func TestWebhookRepository_DeleteWebhookByID(t *testing.T) {
	repo, _ := newWebhookRepo(t)
	ctx := context.Background()

	id := makeWebhook(t, repo, "doomed")
	require.NoError(t, repo.DeleteWebhookByID(ctx, id))

	_, err := repo.GetWebhookByID(ctx, id)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	require.NoError(t, repo.DeleteWebhookByID(ctx, "missing"))
}

func TestWebhookRepository_DBErrorsPropagate(t *testing.T) {
	repo := webhookRepository.NewWebhookRepository(func() *gorm.DB { return closedDB(t) })
	ctx := context.Background()

	t.Run("create", func(t *testing.T) {
		assert.Error(t, repo.CreateWebhook(ctx, &webhookModel.Webhook{ID: "x", Name: "n"}))
	})
	t.Run("get all", func(t *testing.T) {
		_, err := repo.GetAllWebhooks(ctx)
		assert.Error(t, err)
	})
	t.Run("get by id", func(t *testing.T) {
		_, err := repo.GetWebhookByID(ctx, "x")
		assert.Error(t, err)
	})
	t.Run("list active", func(t *testing.T) {
		_, err := repo.ListActiveWebhooks(ctx)
		assert.Error(t, err)
	})
	t.Run("update by id", func(t *testing.T) {
		assert.Error(t, repo.UpdateWebhookByID(ctx, "x", &webhookModel.Webhook{Name: "n"}))
	})
	t.Run("update delivery status", func(t *testing.T) {
		assert.Error(t, repo.UpdateWebhookDeliveryStatus(ctx, "x", "success", 1))
	})
	t.Run("delete", func(t *testing.T) {
		assert.Error(t, repo.DeleteWebhookByID(ctx, "x"))
	})
}

func TestWebhookRepository_TxContext(t *testing.T) {
	repo, db := newWebhookRepo(t)
	transactor := transaction.NewGormTransactor(func() *gorm.DB { return db })

	t.Run("commit persists", func(t *testing.T) {
		var id string
		err := transactor.Run(context.Background(), func(ctx context.Context) error {
			wh := &webhookModel.Webhook{Name: "tx-commit", URL: "https://example.com/c"}
			if err := repo.CreateWebhook(ctx, wh); err != nil {
				return err
			}
			id = wh.ID
			return nil
		})
		require.NoError(t, err)
		got, err := repo.GetWebhookByID(context.Background(), id)
		require.NoError(t, err)
		assert.Equal(t, "tx-commit", got.Name)
	})

	t.Run("rollback discards", func(t *testing.T) {
		sentinel := errors.New("boom")
		var id string
		err := transactor.Run(context.Background(), func(ctx context.Context) error {
			wh := &webhookModel.Webhook{Name: "tx-rollback", URL: "https://example.com/r"}
			if err := repo.CreateWebhook(ctx, wh); err != nil {
				return err
			}
			id = wh.ID
			return sentinel
		})
		require.ErrorIs(t, err, sentinel)
		_, err = repo.GetWebhookByID(context.Background(), id)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "回滚后该行不应可见")
	})
}
