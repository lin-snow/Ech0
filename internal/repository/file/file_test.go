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

func refFile(id, uploaderID string, createdAt int64) fileModel.File {
	url := "https://example.com/" + id
	return fileModel.File{
		ID:          id,
		Kind:        string(fileModel.KindRef),
		Status:      string(fileModel.StatusReady),
		Access:      string(fileModel.AccessPublic),
		RefURL:      &url,
		Name:        id,
		ContentType: "image/png",
		UploaderID:  uploaderID,
		CreatedAt:   createdAt,
	}
}

func TestFileRepository_CreateAndGet(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()

	storageID, objectKey := "st-1", "objects/k-1"
	f := &fileModel.File{
		ID:          "f-create",
		Kind:        string(fileModel.KindOwned),
		Status:      string(fileModel.StatusPending),
		Access:      string(fileModel.AccessPrivate),
		StorageID:   &storageID,
		ObjectKey:   &objectKey,
		Name:        "create.png",
		ContentType: "image/png",
		Size:        42,
		UploaderID:  "u-1",
	}
	require.NoError(t, repo.CreateFile(ctx, f))

	got, err := repo.GetFile(ctx, "f-create")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, f.Kind, got.Kind)
	assert.Equal(t, f.Access, got.Access)
	require.NotNil(t, got.StorageID)
	assert.Equal(t, "st-1", *got.StorageID)
	require.NotNil(t, got.ObjectKey)
	assert.Equal(t, "objects/k-1", *got.ObjectKey)

	t.Run("missing id returns record-not-found", func(t *testing.T) {
		got, err := repo.GetFile(ctx, "missing")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, got)
	})

	var count int64
	require.NoError(t, db.Model(&fileModel.File{}).Where("id = ?", "f-create").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestFileRepository_UpdateName(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()
	insertFile(t, db, refFile("f-rename", "u-1", 1))

	require.NoError(t, repo.UpdateName(ctx, "f-rename", "renamed.png"))

	got, err := repo.GetFile(ctx, "f-rename")
	require.NoError(t, err)
	assert.Equal(t, "renamed.png", got.Name)
}

func TestFileRepository_DeleteFile(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()
	insertFile(t, db, refFile("d-1", "u-1", 1))

	removed, err := repo.DeleteFile(ctx, "d-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), removed)

	var count int64
	require.NoError(t, db.Model(&fileModel.File{}).Where("id = ?", "d-1").Count(&count).Error)
	assert.Zero(t, count)

	t.Run("deleting a missing id affects no rows", func(t *testing.T) {
		removed, err := repo.DeleteFile(ctx, "ghost")
		require.NoError(t, err)
		assert.Zero(t, removed)
	})
}

func TestFileRepository_DeletePendingIfStill(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()
	pending := refFile("p-pending", "u-1", 1)
	pending.Status = string(fileModel.StatusPending)
	insertFile(t, db, pending)
	ready := refFile("p-ready", "u-1", 2)
	ready.Status = string(fileModel.StatusReady)
	insertFile(t, db, ready)
	// The sweep must not delete a row a commit already promoted.
	raced := refFile("p-raced", "u-1", 3)
	raced.Status = string(fileModel.StatusPending)
	insertFile(t, db, raced)
	removed, err := repo.DeletePendingIfStill(ctx, "p-pending")
	require.NoError(t, err)
	assert.Equal(t, int64(1), removed)

	// Simulate the commit between the sweep's SELECT and DELETE.
	rows, err := repo.MarkReady(ctx, "p-raced")
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	removed, err = repo.DeletePendingIfStill(ctx, "p-raced")
	require.NoError(t, err)
	assert.Zero(t, removed, "a committed row is no longer pending")

	removed, err = repo.DeletePendingIfStill(ctx, "p-ready")
	require.NoError(t, err)
	assert.Zero(t, removed)

	var count int64
	require.NoError(t, db.Model(&fileModel.File{}).Where("id = ?", "p-raced").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestFileRepository_MarkReady(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()
	insertFile(t, db, refFile("ready-1", "u-1", 1))
	pending := refFile("ready-2", "u-1", 2)
	pending.Status = string(fileModel.StatusPending)
	insertFile(t, db, pending)

	rows, err := repo.MarkReady(ctx, "ready-2")
	require.NoError(t, err)
	assert.Equal(t, int64(1), rows)

	got, err := repo.GetFile(ctx, "ready-2")
	require.NoError(t, err)
	assert.Equal(t, string(fileModel.StatusReady), got.Status)

	t.Run("already ready affects no rows", func(t *testing.T) {
		rows, err := repo.MarkReady(ctx, "ready-1")
		require.NoError(t, err)
		assert.Zero(t, rows)
	})
}

func TestFileRepository_PendingOlderThan(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()
	pending := func(id string, created int64) fileModel.File {
		f := refFile(id, "u-1", created)
		f.Status = string(fileModel.StatusPending)
		return f
	}
	insertFile(t, db, pending("p-1", 10))
	insertFile(t, db, pending("p-2", 20))
	insertFile(t, db, pending("p-3", 30))
	insertFile(t, db, refFile("ready-new", "u-1", 15)) // ready is not a sweep target

	rows, err := repo.PendingOlderThan(ctx, 25, 10)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "p-1", rows[0].ID)
	assert.Equal(t, "p-2", rows[1].ID)

	t.Run("cutoff is strict and limit applies", func(t *testing.T) {
		rows, err := repo.PendingOlderThan(ctx, 10, 10)
		require.NoError(t, err)
		assert.Empty(t, rows, "created_at < cutoff is strict")

		rows, err = repo.PendingOlderThan(ctx, 100, 1)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "p-1", rows[0].ID)
	})
}

func TestFileRepository_ListFiles(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()

	ownedStorage, ownedKey := "st-1", "objects/alpha"
	owned := refFile("l-owned", "u-1", 100)
	owned.Kind = string(fileModel.KindOwned)
	owned.StorageID = &ownedStorage
	owned.ObjectKey = &ownedKey
	owned.Name = "apple"
	owned.ContentType = "image/png"
	owned.Access = string(fileModel.AccessPrivate)
	insertFile(t, db, owned)

	second := refFile("l-second", "u-1", 200)
	second.Name = "banana"
	second.ContentType = "video/mp4"
	insertFile(t, db, second)

	third := refFile("l-third", "u-2", 300)
	third.Name = "cherry"
	third.ContentType = "application/pdf"
	insertFile(t, db, third)

	require.NoError(t, db.Create(&fileModel.Link{
		FileID: "l-owned", OwnerKind: fileModel.OwnerKindEcho, OwnerID: "e-1",
		Role: fileModel.RoleAttachment, SortOrder: 0,
	}).Error)
	require.NoError(t, db.Create(&fileModel.Link{
		FileID: "l-second", OwnerKind: fileModel.OwnerKindUser, OwnerID: "u-1",
		Role: fileModel.RoleAvatar, SortOrder: 0,
	}).Error)

	ids := func(files []fileModel.File) []string {
		out := make([]string, len(files))
		for i, f := range files {
			out[i] = f.ID
		}
		return out
	}

	t.Run("no filter lists everything newest first", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Equal(t, []string{"l-third", "l-second", "l-owned"}, ids(files))
	})

	t.Run("uploader filter", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{UploaderID: "u-2", Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []string{"l-third"}, ids(files))
	})

	t.Run("kind, access and status filters", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{
			Kind: string(fileModel.KindBlob), Page: 1, PageSize: 10,
		})
		require.NoError(t, err)
		assert.Zero(t, total)
		assert.Empty(t, files)

		files, total, err = repo.ListFiles(ctx, fileModel.ListFilter{
			Access: string(fileModel.AccessPrivate), Page: 1, PageSize: 10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []string{"l-owned"}, ids(files))

		files, total, err = repo.ListFiles(ctx, fileModel.ListFilter{
			Status: string(fileModel.StatusPending), Page: 1, PageSize: 10,
		})
		require.NoError(t, err)
		assert.Zero(t, total)
		assert.Empty(t, files)
	})

	t.Run("owner filters through links", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{
			OwnerKind: fileModel.OwnerKindEcho, OwnerID: "e-1", Page: 1, PageSize: 10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []string{"l-owned"}, ids(files))
	})

	t.Run("role filter", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{
			Role: fileModel.RoleAvatar, Page: 1, PageSize: 10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []string{"l-second"}, ids(files))
	})

	t.Run("search matches name", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{Search: "ban", Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []string{"l-second"}, ids(files))
	})

	t.Run("search matches object key", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{Search: "alph", Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, []string{"l-owned"}, ids(files))
	})

	t.Run("no match returns empty non-nil slice with zero total", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{Search: "zzz", Page: 1, PageSize: 10})
		require.NoError(t, err)
		assert.Zero(t, total)
		require.NotNil(t, files)
		assert.Empty(t, files)
	})

	t.Run("pagination respects page and pageSize", func(t *testing.T) {
		files, total, err := repo.ListFiles(ctx, fileModel.ListFilter{Page: 2, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Equal(t, []string{"l-owned"}, ids(files))
	})

	t.Run("page defaults are normalized", func(t *testing.T) {
		files, _, err := repo.ListFiles(ctx, fileModel.ListFilter{Page: 0, PageSize: 0})
		require.NoError(t, err)
		assert.Len(t, files, 3)
	})
}

func TestFileRepository_Blobs(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()

	f := refFile("b-1", "u-1", 1)
	f.Kind = string(fileModel.KindBlob)
	f.RefURL = nil
	blobBytes := int64(5)
	f.BlobBytes = &blobBytes
	insertFile(t, db, f)

	require.NoError(t, repo.PutBlob(ctx, "b-1", []byte("hello")))

	got, err := repo.GetBlob(ctx, "b-1")
	require.NoError(t, err)
	assert.Equal(t, []byte("hello"), got)

	t.Run("missing blob returns record-not-found", func(t *testing.T) {
		got, err := repo.GetBlob(ctx, "ghost")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
		assert.Nil(t, got)
	})
}

func TestFileRepository_LinkWritesAndReads(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()
	insertFile(t, db, refFile("lf-1", "u-1", 1))
	insertFile(t, db, refFile("lf-2", "u-1", 2))

	require.NoError(t, repo.AddLinks(ctx, []fileModel.Link{
		{FileID: "lf-1", OwnerKind: fileModel.OwnerKindEcho, OwnerID: "e-2", Role: fileModel.RoleAttachment, SortOrder: 1},
		{FileID: "lf-2", OwnerKind: fileModel.OwnerKindEcho, OwnerID: "e-2", Role: fileModel.RoleCover, SortOrder: 0},
		{FileID: "lf-1", OwnerKind: fileModel.OwnerKindEcho, OwnerID: "e-1", Role: fileModel.RoleAttachment, SortOrder: 0},
	}))
	require.NoError(t, repo.AddLinks(ctx, nil), "empty batch is a no-op")

	t.Run("re-adding a link is idempotent", func(t *testing.T) {
		require.NoError(t, repo.AddLinks(ctx, []fileModel.Link{
			{FileID: "lf-1", OwnerKind: fileModel.OwnerKindEcho, OwnerID: "e-1", Role: fileModel.RoleAttachment, SortOrder: 0},
		}))
		n, err := repo.CountLinks(ctx, "lf-1")
		require.NoError(t, err)
		assert.Equal(t, int64(2), n)
	})

	t.Run("LinksOfFile", func(t *testing.T) {
		links, err := repo.LinksOfFile(ctx, "lf-1")
		require.NoError(t, err)
		assert.Len(t, links, 2)
	})

	t.Run("LinksOfOwner orders by role then sort order", func(t *testing.T) {
		links, err := repo.LinksOfOwner(ctx, fileModel.OwnerKindEcho, "e-2", "")
		require.NoError(t, err)
		require.Len(t, links, 2)
		assert.Equal(t, fileModel.RoleAttachment, links[0].Role)
		assert.Equal(t, 1, links[0].SortOrder)
		assert.Equal(t, fileModel.RoleCover, links[1].Role)

		links, err = repo.LinksOfOwner(ctx, fileModel.OwnerKindEcho, "e-2", fileModel.RoleCover)
		require.NoError(t, err)
		require.Len(t, links, 1)
		assert.Equal(t, "lf-2", links[0].FileID)
	})

	t.Run("LinksForFiles groups by file id", func(t *testing.T) {
		grouped, err := repo.LinksForFiles(ctx, []string{"lf-1", "lf-2"})
		require.NoError(t, err)
		assert.Len(t, grouped["lf-1"], 2)
		assert.Len(t, grouped["lf-2"], 1)

		empty, err := repo.LinksForFiles(ctx, nil)
		require.NoError(t, err)
		assert.Empty(t, empty)
	})

	t.Run("FilesOfOwner", func(t *testing.T) {
		files, err := repo.FilesOfOwner(ctx, fileModel.OwnerKindEcho, "e-2", "")
		require.NoError(t, err)
		require.Len(t, files, 2)
		assert.Equal(t, "lf-1", files[0].ID)
		assert.Equal(t, "lf-2", files[1].ID)

		files, err = repo.FilesOfOwner(ctx, fileModel.OwnerKindEcho, "e-2", fileModel.RoleCover)
		require.NoError(t, err)
		require.Len(t, files, 1)
		assert.Equal(t, "lf-2", files[0].ID)
	})

	t.Run("DeleteLink removes one slot only", func(t *testing.T) {
		require.NoError(t, repo.DeleteLink(ctx, "lf-1", fileModel.OwnerKindEcho, "e-1", fileModel.RoleAttachment))
		n, err := repo.CountLinks(ctx, "lf-1")
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)
	})

	t.Run("DeleteOwnerLinks with role only clears that role", func(t *testing.T) {
		require.NoError(t, repo.DeleteOwnerLinks(ctx, fileModel.OwnerKindEcho, "e-2", fileModel.RoleCover))
		links, err := repo.LinksOfOwner(ctx, fileModel.OwnerKindEcho, "e-2", "")
		require.NoError(t, err)
		require.Len(t, links, 1)
		assert.Equal(t, fileModel.RoleAttachment, links[0].Role)

		require.NoError(t, repo.DeleteOwnerLinks(ctx, fileModel.OwnerKindEcho, "e-2", ""))
		links, err = repo.LinksOfOwner(ctx, fileModel.OwnerKindEcho, "e-2", "")
		require.NoError(t, err)
		assert.Empty(t, links)
	})
}

func TestFileRepository_UnlinkedFilesOfUploader(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()
	insertFile(t, db, refFile("up-linked", "u-del", 1))
	insertFile(t, db, refFile("up-orphan", "u-del", 2))
	insertFile(t, db, refFile("up-other", "u-keep", 3))
	require.NoError(t, db.Create(&fileModel.Link{
		FileID: "up-linked", OwnerKind: fileModel.OwnerKindEcho, OwnerID: "e-1",
		Role: fileModel.RoleAttachment, SortOrder: 0,
	}).Error)

	rows, err := repo.UnlinkedFilesOfUploader(ctx, "u-del")
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "up-orphan", rows[0].ID)
}

func TestFileRepository_GCTasks(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()

	base := int64(1000)
	require.NoError(t, repo.AddGCTask(ctx, &fileModel.GCTask{
		ID: "gc-2", StorageID: "st-1", ObjectKey: "k-2", NextTryAt: base + 20,
	}))
	require.NoError(t, repo.AddGCTask(ctx, &fileModel.GCTask{
		ID: "gc-1", StorageID: "st-1", ObjectKey: "k-1", NextTryAt: base + 10,
	}))
	require.NoError(t, repo.AddGCTask(ctx, &fileModel.GCTask{
		ID: "gc-3", StorageID: "st-1", ObjectKey: "k-3", NextTryAt: base + 30,
	}))

	t.Run("due tasks ordered by next_try_at and limited", func(t *testing.T) {
		tasks, err := repo.DueGCTasks(ctx, base+20, 10)
		require.NoError(t, err)
		require.Len(t, tasks, 2)
		assert.Equal(t, "gc-1", tasks[0].ID)
		assert.Equal(t, "gc-2", tasks[1].ID)

		tasks, err = repo.DueGCTasks(ctx, base+100, 1)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, "gc-1", tasks[0].ID)
	})

	t.Run("update persists attempts and next try", func(t *testing.T) {
		tasks, err := repo.DueGCTasks(ctx, base+10, 1)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		tasks[0].Attempts = 3
		tasks[0].LastError = "boom"
		tasks[0].NextTryAt = base + 500
		require.NoError(t, repo.UpdateGCTask(ctx, &tasks[0]))

		var fresh fileModel.GCTask
		require.NoError(t, db.First(&fresh, "id = ?", "gc-1").Error)
		assert.Equal(t, 3, fresh.Attempts)
		assert.Equal(t, "boom", fresh.LastError)
		assert.Equal(t, base+500, fresh.NextTryAt)
	})

	t.Run("count and delete", func(t *testing.T) {
		n, err := repo.CountGCTasks(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), n)

		require.NoError(t, repo.DeleteGCTask(ctx, "gc-3"))
		n, err = repo.CountGCTasks(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), n)
	})
}

func TestFileRepository_Storages(t *testing.T) {
	repo, db := newFileRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SaveStorage(ctx, &fileModel.StorageConfig{
		ID: "st-local", Driver: "local", Name: "local", IsDefault: true,
		Config: map[string]string{"root": "./data"},
	}))
	require.NoError(t, repo.SaveStorage(ctx, &fileModel.StorageConfig{
		ID: "st-s3", Driver: "s3", Name: "s3", Config: map[string]string{"bucket": "b"},
	}))

	t.Run("list is ordered by created_at", func(t *testing.T) {
		rows, err := repo.ListStorages(ctx)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, "st-local", rows[0].ID)
		assert.Equal(t, "st-s3", rows[1].ID)
	})

	t.Run("save upserts by id", func(t *testing.T) {
		require.NoError(t, repo.SaveStorage(ctx, &fileModel.StorageConfig{
			ID: "st-s3", Driver: "s3", Name: "renamed", Config: map[string]string{"bucket": "b2"},
		}))
		var fresh fileModel.StorageConfig
		require.NoError(t, db.First(&fresh, "id = ?", "st-s3").Error)
		assert.Equal(t, "renamed", fresh.Name)
		assert.Equal(t, "b2", fresh.Config["bucket"])
	})

	t.Run("set default moves the flag", func(t *testing.T) {
		require.NoError(t, repo.SetDefaultStorage(ctx, "st-s3"))

		var defaults int64
		require.NoError(t, db.Model(&fileModel.StorageConfig{}).Where("is_default = ?", true).Count(&defaults).Error)
		assert.Equal(t, int64(1), defaults, "a partial unique index allows one default")

		var fresh fileModel.StorageConfig
		require.NoError(t, db.First(&fresh, "id = ?", "st-s3").Error)
		assert.True(t, fresh.IsDefault)
	})

	t.Run("delete removes the row", func(t *testing.T) {
		require.NoError(t, repo.DeleteStorage(ctx, "st-local"))
		rows, err := repo.ListStorages(ctx)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "st-s3", rows[0].ID)
	})
}

func TestFileRepository_GetDB_UsesContextTx(t *testing.T) {
	repo, db := newFileRepo(t)

	tx := db.Begin()
	require.NoError(t, tx.Error)
	ctxTx := context.WithValue(context.Background(), transaction.TxKey, tx)

	require.NoError(t, repo.CreateFile(ctxTx, &fileModel.File{
		ID: "tx-file", Kind: string(fileModel.KindRef), Status: string(fileModel.StatusReady),
		Access: string(fileModel.AccessPublic), ContentType: "text/plain", UploaderID: "u-1",
	}))
	require.NoError(t, tx.Rollback().Error)

	got, err := repo.GetFile(context.Background(), "tx-file")
	require.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	assert.Nil(t, got)
}
