// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package keyvalue

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newKVRepo(t *testing.T) (*KeyValueRepository, *gorm.DB, cache.ICache[string, any]) {
	t.Helper()
	db := helpers.NewTestDB(t)
	c := helpers.NewTestCache()
	return NewKeyValueRepository(func() *gorm.DB { return db }, c), db, c
}

func cacheGetKV(t *testing.T, c cache.ICache[string, any], key string) (any, bool) {
	t.Helper()
	v, ok, err := c.Get(key)
	require.NoError(t, err)
	return v, ok
}

func countKV(t *testing.T, db *gorm.DB, key string) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&commonModel.KeyValue{}).Where("key = ?", key).Count(&n).Error)
	return n
}

func TestKeyValueRepository_AddOrUpdateKeyValue(t *testing.T) {
	t.Run("creates row when key is absent (RowsAffected==0 path)", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		c.Set(GetKeyValueCacheKey("k"), "stale", 1)

		require.NoError(t, repo.AddOrUpdateKeyValue(context.Background(), "k", "v1"))

		var kv commonModel.KeyValue
		require.NoError(t, db.Where("key = ?", "k").First(&kv).Error)
		assert.Equal(t, "v1", kv.Value)
		assert.Equal(t, int64(1), countKV(t, db, "k"))

		got, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		require.True(t, ok, "cache must be backfilled after upsert")
		assert.Equal(t, "v1", got)
	})

	t.Run("updates existing row in place (update path) without duplicating", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		require.NoError(t, db.Create(&commonModel.KeyValue{Key: "k", Value: "old"}).Error)
		c.Set(GetKeyValueCacheKey("k"), "stale", 1)

		require.NoError(t, repo.AddOrUpdateKeyValue(context.Background(), "k", "v2"))

		var kv commonModel.KeyValue
		require.NoError(t, db.Where("key = ?", "k").First(&kv).Error)
		assert.Equal(t, "v2", kv.Value)
		assert.Equal(t, int64(1), countKV(t, db, "k"), "update must not create a second row")

		got, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		require.True(t, ok)
		assert.Equal(t, "v2", got)
	})
}

func TestKeyValueRepository_GetKeyValue(t *testing.T) {
	t.Run("cache hit returns cached value without touching db", func(t *testing.T) {
		repo, _, c := newKVRepo(t)
		c.Set(GetKeyValueCacheKey("k"), "cached", 1)

		got, err := repo.GetKeyValue(context.Background(), "k")
		require.NoError(t, err)
		assert.Equal(t, "cached", got)
	})

	t.Run("cache miss reads db and backfills", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		require.NoError(t, db.Create(&commonModel.KeyValue{Key: "k", Value: "dbval"}).Error)

		got, err := repo.GetKeyValue(context.Background(), "k")
		require.NoError(t, err)
		assert.Equal(t, "dbval", got)

		cached, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		require.True(t, ok, "miss must backfill cache")
		assert.Equal(t, "dbval", cached)
	})

	t.Run("transaction context bypasses cache and reads db", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		require.NoError(t, db.Create(&commonModel.KeyValue{Key: "k", Value: "dbval"}).Error)
		c.Set(GetKeyValueCacheKey("k"), "stale", 1)

		ctxTx := context.WithValue(context.Background(), transaction.TxKey, db)
		got, err := repo.GetKeyValue(ctxTx, "k")
		require.NoError(t, err)
		assert.Equal(t, "dbval", got, "tx loader must ignore the stale cache")

		cached, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		require.True(t, ok)
		assert.Equal(t, "stale", cached, "tx read must not mutate the cache")
	})

	t.Run("missing key returns record-not-found", func(t *testing.T) {
		repo, _, _ := newKVRepo(t)

		got, err := repo.GetKeyValue(context.Background(), "nope")
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Empty(t, got)
	})
}

func TestKeyValueRepository_AddKeyValue(t *testing.T) {
	t.Run("creates row and backfills cache after invalidating stale", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		c.Set(GetKeyValueCacheKey("k"), "stale", 1)

		require.NoError(t, repo.AddKeyValue(context.Background(), "k", "v"))

		var kv commonModel.KeyValue
		require.NoError(t, db.Where("key = ?", "k").First(&kv).Error)
		assert.Equal(t, "v", kv.Value)

		got, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		require.True(t, ok)
		assert.Equal(t, "v", got)
	})

	t.Run("duplicate key errors and leaves cache invalidated (no stale backfill)", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		require.NoError(t, db.Create(&commonModel.KeyValue{Key: "k", Value: "existing"}).Error)
		c.Set(GetKeyValueCacheKey("k"), "stale", 1)

		err := repo.AddKeyValue(context.Background(), "k", "v")
		require.Error(t, err, "creating a duplicate primary key must fail")

		_, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		assert.False(t, ok, "failed create must not leave a backfilled value")

		var kv commonModel.KeyValue
		require.NoError(t, db.Where("key = ?", "k").First(&kv).Error)
		assert.Equal(t, "existing", kv.Value, "row value must be unchanged")
	})
}

func TestKeyValueRepository_UpdateKeyValue(t *testing.T) {
	t.Run("updates row and backfills cache after invalidating stale", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		require.NoError(t, db.Create(&commonModel.KeyValue{Key: "k", Value: "old"}).Error)
		c.Set(GetKeyValueCacheKey("k"), "stale", 1)

		require.NoError(t, repo.UpdateKeyValue(context.Background(), "k", "new"))

		var kv commonModel.KeyValue
		require.NoError(t, db.Where("key = ?", "k").First(&kv).Error)
		assert.Equal(t, "new", kv.Value)

		got, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		require.True(t, ok)
		assert.Equal(t, "new", got)
	})
}

func TestKeyValueRepository_DeleteKeyValue(t *testing.T) {
	t.Run("removes row and invalidates cache", func(t *testing.T) {
		repo, db, c := newKVRepo(t)
		require.NoError(t, db.Create(&commonModel.KeyValue{Key: "k", Value: "v"}).Error)
		c.Set(GetKeyValueCacheKey("k"), "v", 1)

		require.NoError(t, repo.DeleteKeyValue(context.Background(), "k"))

		assert.Equal(t, int64(0), countKV(t, db, "k"), "row must be deleted")
		_, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
		assert.False(t, ok, "cache must be invalidated on delete")
	})
}

func TestKeyValueRepository_CacheConsistencyFlow(t *testing.T) {
	repo, _, c := newKVRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.AddKeyValue(ctx, "k", "v1"))
	got, err := repo.GetKeyValue(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, "v1", got, "read after add should observe the new value")

	require.NoError(t, repo.UpdateKeyValue(ctx, "k", "v2"))
	got, err = repo.GetKeyValue(ctx, "k")
	require.NoError(t, err)
	assert.Equal(t, "v2", got, "read after update should observe the updated value")

	require.NoError(t, repo.DeleteKeyValue(ctx, "k"))
	_, ok := cacheGetKV(t, c, GetKeyValueCacheKey("k"))
	assert.False(t, ok, "delete should clear the cache entry")
	_, err = repo.GetKeyValue(ctx, "k")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound, "read after delete should miss both cache and db")
}
