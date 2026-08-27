// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	model "github.com/lin-snow/ech0/internal/model/embedding"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newEmbeddingRepo(t *testing.T) (*EmbeddingRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDBWithVec(t)
	return NewEmbeddingRepository(func() *gorm.DB { return db }), db
}

func vec4(x float32) []float32 { return []float32{x, 0, 0, 0} }

func seed(t *testing.T, repo *EmbeddingRepository, ctx context.Context, echoID, username string, x float32) {
	t.Helper()
	meta := &model.EchoEmbedding{
		EchoID:      echoID,
		ContentHash: "h-" + echoID,
		Model:       "test-model",
		Dim:         4,
		Content:     "content-" + echoID,
		Username:    username,
		EchoCreated: 1000,
	}
	require.NoError(t, repo.Upsert(ctx, meta, vec4(x)))
}

func vecRowCount(t *testing.T, db *gorm.DB, echoID string) int {
	t.Helper()
	var n int
	require.NoError(t, db.Raw("SELECT count(*) FROM "+vecTable+" WHERE echo_id = ?", echoID).Scan(&n).Error)
	return n
}

func vecTotal(t *testing.T, db *gorm.DB) int {
	t.Helper()
	var n int
	require.NoError(t, db.Raw("SELECT count(*) FROM "+vecTable).Scan(&n).Error)
	return n
}

func ids(results []model.SearchResult) []string {
	out := make([]string, len(results))
	for i, r := range results {
		out[i] = r.EchoID
	}
	return out
}

func TestEmbeddingRepository_EnsureVecTable(t *testing.T) {
	repo, _ := newEmbeddingRepo(t)
	ctx := context.Background()

	t.Run("invalid dim returns error", func(t *testing.T) {
		for _, dim := range []int{0, -1, -8} {
			err := repo.EnsureVecTable(ctx, dim)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid vector dim")
		}
	})

	t.Run("valid dim is idempotent", func(t *testing.T) {
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		seed(t, repo, ctx, "e-ensure", "u", 1)
		assert.Equal(t, 1, vecRowCount(t, repo.db(), "e-ensure"))
	})
}

func TestEmbeddingRepository_DropVecTable(t *testing.T) {
	repo, db := newEmbeddingRepo(t)
	ctx := context.Background()

	t.Run("drop when not created is no-op", func(t *testing.T) {
		require.NoError(t, repo.DropVecTable(ctx))
	})

	t.Run("create then drop removes the virtual table", func(t *testing.T) {
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		seed(t, repo, ctx, "e-drop", "u", 1)
		require.NoError(t, repo.DropVecTable(ctx))

		var n int
		err := db.Raw("SELECT count(*) FROM " + vecTable).Scan(&n).Error
		require.Error(t, err)

		require.NoError(t, repo.DropVecTable(ctx))
	})
}

func TestEmbeddingRepository_Upsert(t *testing.T) {
	repo, db := newEmbeddingRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.EnsureVecTable(ctx, 4))

	t.Run("insert writes meta and vector", func(t *testing.T) {
		seed(t, repo, ctx, "e-1", "alice", 1)

		got, ok, err := repo.GetMeta(ctx, "e-1")
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "alice", got.Username)
		assert.Equal(t, "content-e-1", got.Content)
		assert.Equal(t, 4, got.Dim)
		assert.Equal(t, 1, vecRowCount(t, db, "e-1"))
	})

	t.Run("re-upsert updates meta (OnConflict UpdateAll) and replaces vector idempotently", func(t *testing.T) {
		meta := &model.EchoEmbedding{
			EchoID:      "e-1",
			ContentHash: "h-new",
			Model:       "test-model",
			Dim:         4,
			Content:     "updated",
			Username:    "bob",
			EchoCreated: 2000,
		}
		require.NoError(t, repo.Upsert(ctx, meta, vec4(9)))

		got, ok, err := repo.GetMeta(ctx, "e-1")
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "updated", got.Content)
		assert.Equal(t, "bob", got.Username)
		assert.Equal(t, "h-new", got.ContentHash)

		cnt, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), cnt)

		assert.Equal(t, 1, vecRowCount(t, db, "e-1"))
		assert.Equal(t, 1, vecTotal(t, db))

		res, err := repo.Search(ctx, vec4(0), 5, "")
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, "e-1", res[0].EchoID)
		assert.InDelta(t, 9.0, res[0].Distance, 0.001)
	})

	t.Run("propagates vec write error when vec table is absent", func(t *testing.T) {
		repo2, _ := newEmbeddingRepo(t)
		ctx2 := context.Background()
		meta := &model.EchoEmbedding{EchoID: "e-novec", Username: "u", Dim: 4}
		err := repo2.Upsert(ctx2, meta, vec4(1))
		require.Error(t, err)

		_, ok, gerr := repo2.GetMeta(ctx2, "e-novec")
		require.NoError(t, gerr)
		assert.True(t, ok)
	})
}

func TestEmbeddingRepository_TxContext(t *testing.T) {
	repo, db := newEmbeddingRepo(t)
	require.NoError(t, repo.EnsureVecTable(context.Background(), 4))

	err := db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(context.Background(), transaction.TxKey, tx)
		seed(t, repo, txCtx, "e-tx", "u", 1)

		got, ok, err := repo.GetMeta(txCtx, "e-tx")
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "u", got.Username)
		return assert.AnError
	})
	require.Error(t, err)

	_, ok, gerr := repo.GetMeta(context.Background(), "e-tx")
	require.NoError(t, gerr)
	assert.False(t, ok)
}

func TestEmbeddingRepository_GetMeta(t *testing.T) {
	repo, _ := newEmbeddingRepo(t)
	ctx := context.Background()

	t.Run("missing returns ok=false without error and without vec table", func(t *testing.T) {
		got, ok, err := repo.GetMeta(ctx, "nope")
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("present returns the row", func(t *testing.T) {
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		seed(t, repo, ctx, "e-get", "carol", 2)
		got, ok, err := repo.GetMeta(ctx, "e-get")
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "carol", got.Username)
	})
}

func TestEmbeddingRepository_Count(t *testing.T) {
	repo, db := newEmbeddingRepo(t)
	ctx := context.Background()

	got, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), got)

	for _, id := range []string{"a", "b", "c"} {
		require.NoError(t, db.Create(&model.EchoEmbedding{EchoID: id, Username: "u", Dim: 4}).Error)
	}
	got, err = repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), got)
}

func TestEmbeddingRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("removes meta and vector", func(t *testing.T) {
		repo, db := newEmbeddingRepo(t)
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		seed(t, repo, ctx, "e-del", "u", 1)
		seed(t, repo, ctx, "e-keep", "u", 2)

		require.NoError(t, repo.Delete(ctx, "e-del"))

		_, ok, err := repo.GetMeta(ctx, "e-del")
		require.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, 0, vecRowCount(t, db, "e-del"))

		_, ok, err = repo.GetMeta(ctx, "e-keep")
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, 1, vecRowCount(t, db, "e-keep"))
	})

	t.Run("tolerates missing vec table (ignores vec delete error)", func(t *testing.T) {
		repo, db := newEmbeddingRepo(t)
		require.NoError(t, db.Create(&model.EchoEmbedding{EchoID: "e-nv", Username: "u", Dim: 4}).Error)

		require.NoError(t, repo.Delete(ctx, "e-nv"))
		_, ok, err := repo.GetMeta(ctx, "e-nv")
		require.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestEmbeddingRepository_ClearAll(t *testing.T) {
	repo, db := newEmbeddingRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.EnsureVecTable(ctx, 4))

	for i, id := range []string{"c1", "c2", "c3"} {
		seed(t, repo, ctx, id, "u", float32(i+1))
	}
	require.Equal(t, 3, vecTotal(t, db))

	require.NoError(t, repo.ClearAll(ctx))

	cnt, err := repo.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), cnt)
	assert.Equal(t, 0, vecTotal(t, db))
}

func TestEmbeddingRepository_Search(t *testing.T) {
	t.Run("empty index returns nil", func(t *testing.T) {
		repo, _ := newEmbeddingRepo(t)
		ctx := context.Background()
		require.NoError(t, repo.EnsureVecTable(ctx, 4))

		res, err := repo.Search(ctx, vec4(1), 5, "")
		require.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("no author returns k nearest in distance order", func(t *testing.T) {
		repo, _ := newEmbeddingRepo(t)
		ctx := context.Background()
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		for i := 1; i <= 5; i++ {
			seed(t, repo, ctx, idAt(i), "u", float32(i))
		}

		res, err := repo.Search(ctx, vec4(0), 3, "")
		require.NoError(t, err)
		require.Len(t, res, 3)
		assert.Equal(t, []string{idAt(1), idAt(2), idAt(3)}, ids(res))
		assertAscending(t, res)
	})

	t.Run("k<=0 defaults to 6", func(t *testing.T) {
		repo, _ := newEmbeddingRepo(t)
		ctx := context.Background()
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		for i := 1; i <= 8; i++ {
			seed(t, repo, ctx, idAt(i), "u", float32(i))
		}

		res, err := repo.Search(ctx, vec4(0), 0, "")
		require.NoError(t, err)
		require.Len(t, res, 6)
		assert.Equal(t, idAt(1), res[0].EchoID)
		assert.Equal(t, idAt(6), res[5].EchoID)
	})

	t.Run("author scoping overfetches then filters and truncates to k", func(t *testing.T) {
		repo, _ := newEmbeddingRepo(t)
		ctx := context.Background()
		require.NoError(t, repo.EnsureVecTable(ctx, 4))

		seed(t, repo, ctx, "bob-1", "bob", 1)
		seed(t, repo, ctx, "bob-2", "bob", 2)
		seed(t, repo, ctx, "bob-3", "bob", 3)
		seed(t, repo, ctx, "alice-4", "alice", 4)
		seed(t, repo, ctx, "alice-5", "alice", 5)
		seed(t, repo, ctx, "alice-6", "alice", 6)

		res, err := repo.Search(ctx, vec4(0), 2, "alice")
		require.NoError(t, err)
		require.Len(t, res, 2)
		assert.Equal(t, []string{"alice-4", "alice-5"}, ids(res))
		for _, r := range res {
			assert.Equal(t, "alice", r.Username)
		}
		assertAscending(t, res)
	})

	t.Run("author with no hits returns empty", func(t *testing.T) {
		repo, _ := newEmbeddingRepo(t)
		ctx := context.Background()
		require.NoError(t, repo.EnsureVecTable(ctx, 4))
		seed(t, repo, ctx, "bob-1", "bob", 1)

		res, err := repo.Search(ctx, vec4(0), 3, "ghost")
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func idAt(i int) string {
	return "e-" + string(rune('a'+i-1))
}

func assertAscending(t *testing.T, res []model.SearchResult) {
	t.Helper()
	for i := 1; i < len(res); i++ {
		assert.LessOrEqual(t, res[i-1].Distance, res[i].Distance, "results must be in ascending distance order")
	}
}
