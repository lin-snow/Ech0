// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"
	"testing"

	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newFileRepo(t *testing.T) (*FileRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return NewFileRepository(func() *gorm.DB { return db }), db
}

func insertFile(t *testing.T, db *gorm.DB, f fileModel.File) fileModel.File {
	t.Helper()
	require.NoError(t, db.Create(&f).Error)
	return f
}

func TestFileRepository_Create(t *testing.T) {
	repo, db := newFileRepo(t)

	f := &fileModel.File{
		ID:          "f-create",
		Key:         "k-create",
		StorageType: "local",
		URL:         "/files/k-create",
		Name:        "create.png",
		Category:    "image",
		UserID:      "u-1",
	}
	require.NoError(t, repo.Create(context.Background(), f))

	var got fileModel.File
	require.NoError(t, db.First(&got, "id = ?", "f-create").Error)
	assert.Equal(t, "k-create", got.Key)
	assert.Equal(t, "local", got.StorageType)

	t.Run("BeforeCreate fills uuid when id empty", func(t *testing.T) {
		g := &fileModel.File{Key: "k-gen", StorageType: "local", UserID: "u-1"}
		require.NoError(t, repo.Create(context.Background(), g))
		assert.NotEmpty(t, g.ID)
	})
}

func TestFileRepository_GetByID(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "f-1", Key: "k-1", StorageType: "local", UserID: "u-1"})

	t.Run("found", func(t *testing.T) {
		got, err := repo.GetByID(context.Background(), "f-1")
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "k-1", got.Key)
	})

	t.Run("not found returns ErrRecordNotFound", func(t *testing.T) {
		got, err := repo.GetByID(context.Background(), "missing")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, got)
	})
}

func TestFileRepository_GetByKey(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "f-2", Key: "unique-key", StorageType: "local", UserID: "u-1"})

	t.Run("found by key", func(t *testing.T) {
		got, err := repo.GetByKey(context.Background(), "unique-key")
		require.NoError(t, err)
		assert.Equal(t, "f-2", got.ID)
	})

	t.Run("not found", func(t *testing.T) {
		got, err := repo.GetByKey(context.Background(), "nope")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, got)
	})
}

func TestFileRepository_GetByRoute(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{
		ID: "f-route", Key: "obj-key", StorageType: "object",
		Provider: "r2", Bucket: "main", UserID: "u-1",
	})

	t.Run("matches full composite route", func(t *testing.T) {
		got, err := repo.GetByRoute(context.Background(), "object", "r2", "main", "obj-key")
		require.NoError(t, err)
		assert.Equal(t, "f-route", got.ID)
	})

	t.Run("mismatched bucket misses", func(t *testing.T) {
		got, err := repo.GetByRoute(context.Background(), "object", "r2", "other", "obj-key")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, got)
	})
}

func TestFileRepository_ListByStorageTypeAndSearch(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "s-1", Key: "alpha", Name: "apple", StorageType: "local", UserID: "u-1", CreatedAt: 100})
	insertFile(t, db, fileModel.File{ID: "s-2", Key: "beta", Name: "banana", StorageType: "local", UserID: "u-1", CreatedAt: 200})
	insertFile(t, db, fileModel.File{ID: "s-3", Key: "gamma", Name: "cherry", StorageType: "object", Provider: "r2", Bucket: "b", UserID: "u-1", CreatedAt: 300})

	t.Run("filter by storage type, ordered by created_at desc", func(t *testing.T) {
		files, total, err := repo.ListByStorageTypeAndSearch(context.Background(), "local", "", 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		require.Len(t, files, 2)
		assert.Equal(t, "s-2", files[0].ID)
		assert.Equal(t, "s-1", files[1].ID)
	})

	t.Run("empty storage type spans all rows", func(t *testing.T) {
		files, total, err := repo.ListByStorageTypeAndSearch(context.Background(), "", "", 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, files, 3)
	})

	t.Run("search matches name", func(t *testing.T) {
		files, total, err := repo.ListByStorageTypeAndSearch(context.Background(), "", "ban", 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, files, 1)
		assert.Equal(t, "s-2", files[0].ID)
	})

	t.Run("search matches key", func(t *testing.T) {
		files, total, err := repo.ListByStorageTypeAndSearch(context.Background(), "", "gamm", 1, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, files, 1)
		assert.Equal(t, "s-3", files[0].ID)
	})

	t.Run("no match returns empty non-nil slice with zero total", func(t *testing.T) {
		files, total, err := repo.ListByStorageTypeAndSearch(context.Background(), "local", "zzz", 1, 10)
		require.NoError(t, err)
		assert.Zero(t, total)
		require.NotNil(t, files)
		assert.Empty(t, files)
	})

	t.Run("pagination respects page and pageSize", func(t *testing.T) {
		files, total, err := repo.ListByStorageTypeAndSearch(context.Background(), "", "", 2, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		require.Len(t, files, 1)
		assert.Equal(t, "s-1", files[0].ID)
	})
}

func TestFileRepository_ListByStorageTypeAndKeys(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "k-a", Key: "kk-a", StorageType: "local", UserID: "u-1"})
	insertFile(t, db, fileModel.File{ID: "k-b", Key: "kk-b", StorageType: "local", UserID: "u-1"})
	insertFile(t, db, fileModel.File{ID: "k-c", Key: "kk-c", StorageType: "object", Provider: "r2", Bucket: "b", UserID: "u-1"})

	t.Run("empty keys short-circuits to empty slice", func(t *testing.T) {
		files, err := repo.ListByStorageTypeAndKeys(context.Background(), "local", nil)
		require.NoError(t, err)
		require.NotNil(t, files)
		assert.Empty(t, files)
	})

	t.Run("filters by storage type and key set", func(t *testing.T) {
		files, err := repo.ListByStorageTypeAndKeys(context.Background(), "local", []string{"kk-a", "kk-b", "kk-c"})
		require.NoError(t, err)
		require.Len(t, files, 2)
		ids := map[string]bool{files[0].ID: true, files[1].ID: true}
		assert.True(t, ids["k-a"] && ids["k-b"])
	})
}

func TestFileRepository_ListByStorageTypeAndURLs(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "u-a", Key: "ua", URL: "/files/ua", StorageType: "local", UserID: "u-1"})
	insertFile(t, db, fileModel.File{ID: "u-b", Key: "ub", URL: "/files/ub", StorageType: "local", UserID: "u-1"})

	t.Run("empty urls short-circuits", func(t *testing.T) {
		files, err := repo.ListByStorageTypeAndURLs(context.Background(), "local", []string{})
		require.NoError(t, err)
		require.NotNil(t, files)
		assert.Empty(t, files)
	})

	t.Run("filters by storage type and url set", func(t *testing.T) {
		files, err := repo.ListByStorageTypeAndURLs(context.Background(), "local", []string{"/files/ua", "/files/missing"})
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "u-a", files[0].ID)
	})
}

func TestFileRepository_UpdateMetaByID(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "m-1", Key: "mk", StorageType: "local", UserID: "u-1", Size: 1})

	t.Run("updates size only when optional fields nil", func(t *testing.T) {
		got, err := repo.UpdateMetaByID(context.Background(), "m-1", 4096, nil, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(4096), got.Size)
		assert.Zero(t, got.Width)
		assert.Zero(t, got.Height)
		assert.Empty(t, got.ContentType)
	})

	t.Run("updates optional fields when provided", func(t *testing.T) {
		w, h := 800, 600
		ct := "image/png"
		got, err := repo.UpdateMetaByID(context.Background(), "m-1", 2048, &w, &h, &ct)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(2048), got.Size)
		assert.Equal(t, 800, got.Width)
		assert.Equal(t, 600, got.Height)
		assert.Equal(t, "image/png", got.ContentType)

		var fresh fileModel.File
		require.NoError(t, db.First(&fresh, "id = ?", "m-1").Error)
		assert.Equal(t, 800, fresh.Width)
	})
}

func TestFileRepository_Delete(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "d-1", Key: "dk", StorageType: "local", UserID: "u-1"})

	require.NoError(t, repo.Delete(context.Background(), "d-1"))

	var count int64
	require.NoError(t, db.Model(&fileModel.File{}).Where("id = ?", "d-1").Count(&count).Error)
	assert.Zero(t, count)

	t.Run("deleting a missing id is not an error", func(t *testing.T) {
		require.NoError(t, repo.Delete(context.Background(), "ghost"))
	})
}

func TestFileRepository_DeleteByRoute(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{
		ID: "dr-1", Key: "drk", StorageType: "object",
		Provider: "r2", Bucket: "main", UserID: "u-1",
	})

	t.Run("mismatched route deletes nothing", func(t *testing.T) {
		require.NoError(t, repo.DeleteByRoute(context.Background(), "object", "r2", "wrong", "drk"))
		var count int64
		require.NoError(t, db.Model(&fileModel.File{}).Where("id = ?", "dr-1").Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("full route match deletes the row", func(t *testing.T) {
		require.NoError(t, repo.DeleteByRoute(context.Background(), "object", "r2", "main", "drk"))
		var count int64
		require.NoError(t, db.Model(&fileModel.File{}).Where("id = ?", "dr-1").Count(&count).Error)
		assert.Zero(t, count)
	})
}

func TestFileRepository_TempLifecycle(t *testing.T) {
	repo, db := newFileRepo(t)

	t.Run("create then delete by file id", func(t *testing.T) {
		require.NoError(t, repo.CreateTemp(context.Background(), &fileModel.TempFile{
			ID: "t-1", FileID: "file-1", UploaderID: "u-1", ExpireAt: 9999,
		}))

		var count int64
		require.NoError(t, db.Model(&fileModel.TempFile{}).Where("file_id = ?", "file-1").Count(&count).Error)
		assert.Equal(t, int64(1), count)

		require.NoError(t, repo.DeleteTempByFileID(context.Background(), "file-1"))
		require.NoError(t, db.Model(&fileModel.TempFile{}).Where("file_id = ?", "file-1").Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("delete by id", func(t *testing.T) {
		require.NoError(t, repo.CreateTemp(context.Background(), &fileModel.TempFile{
			ID: "t-2", FileID: "file-2", UploaderID: "u-1", ExpireAt: 9999,
		}))
		require.NoError(t, repo.DeleteTempByID(context.Background(), "t-2"))

		var count int64
		require.NoError(t, db.Model(&fileModel.TempFile{}).Where("id = ?", "t-2").Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("BeforeCreate fills temp uuid", func(t *testing.T) {
		temp := &fileModel.TempFile{FileID: "file-gen", UploaderID: "u-1", ExpireAt: 1}
		require.NoError(t, repo.CreateTemp(context.Background(), temp))
		assert.NotEmpty(t, temp.ID)
	})
}

func TestFileRepository_ListExpiredTemps(t *testing.T) {
	repo, db := newFileRepo(t)
	require.NoError(t, db.Create(&fileModel.TempFile{ID: "e-1", FileID: "ef-1", UploaderID: "u-1", ExpireAt: 10, CreatedAt: 1}).Error)
	require.NoError(t, db.Create(&fileModel.TempFile{ID: "e-2", FileID: "ef-2", UploaderID: "u-1", ExpireAt: 20, CreatedAt: 2}).Error)
	require.NoError(t, db.Create(&fileModel.TempFile{ID: "e-3", FileID: "ef-3", UploaderID: "u-1", ExpireAt: 30, CreatedAt: 3}).Error)

	t.Run("returns only temps with expire_at strictly below cutoff, ordered created_at asc", func(t *testing.T) {
		got, err := repo.ListExpiredTemps(context.Background(), 25)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "e-1", got[0].ID)
		assert.Equal(t, "e-2", got[1].ID)
	})

	t.Run("cutoff equal to expire_at excludes the boundary row", func(t *testing.T) {
		got, err := repo.ListExpiredTemps(context.Background(), 10)
		require.NoError(t, err)
		assert.Empty(t, got, "expire_at < cutoff is strict")
	})

	t.Run("none expired returns empty", func(t *testing.T) {
		got, err := repo.ListExpiredTemps(context.Background(), 1)
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestFileRepository_GetByCategory(t *testing.T) {
	repo, db := newFileRepo(t)
	insertFile(t, db, fileModel.File{ID: "cat-1", Key: "ck1", StorageType: "local", Category: "image", UserID: "u-1"})
	insertFile(t, db, fileModel.File{ID: "cat-2", Key: "ck2", StorageType: "local", Category: "image", UserID: "u-1"})
	insertFile(t, db, fileModel.File{ID: "cat-3", Key: "ck3", StorageType: "local", Category: "video", UserID: "u-1"})

	t.Run("returns matching category", func(t *testing.T) {
		got, err := repo.GetByCategory(context.Background(), "image")
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("unknown category returns empty", func(t *testing.T) {
		got, err := repo.GetByCategory(context.Background(), "audio")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestFileRepository_GetDB_UsesContextTx(t *testing.T) {
	repo, db := newFileRepo(t)

	tx := db.Begin()
	require.NoError(t, tx.Error)
	ctxTx := context.WithValue(context.Background(), transaction.TxKey, tx)

	require.NoError(t, repo.Create(ctxTx, &fileModel.File{
		ID: "tx-file", Key: "tx-key", StorageType: "local", UserID: "u-1",
	}))
	require.NoError(t, tx.Rollback().Error)

	got, err := repo.GetByID(context.Background(), "tx-file")
	require.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	assert.Nil(t, got)
}
