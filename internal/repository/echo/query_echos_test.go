// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newEchoRepo(t *testing.T) (*EchoRepository, *gorm.DB) {
	t.Helper()
	db := helpers.NewTestDB(t)
	return NewEchoRepository(func() *gorm.DB { return db }, helpers.NewTestCache()), db
}

func seedEcho(
	t *testing.T,
	db *gorm.DB,
	id, content string,
	private bool,
	favCount int,
	createdAt int64,
) echoModel.Echo {
	t.Helper()
	e := echoModel.Echo{
		ID:        id,
		Content:   content,
		UserID:    "u1",
		Private:   private,
		FavCount:  favCount,
		CreatedAt: createdAt,
	}
	require.NoError(t, db.Create(&e).Error)
	return e
}

func seedTag(t *testing.T, db *gorm.DB, id, name string) {
	t.Helper()
	require.NoError(t, db.Create(&echoModel.Tag{ID: id, Name: name}).Error)
}

func linkTag(t *testing.T, db *gorm.DB, echoID, tagID string) {
	t.Helper()
	require.NoError(t, db.Create(&echoModel.EchoTag{EchoID: echoID, TagID: tagID}).Error)
}

func echoIDs(echos []echoModel.Echo) []string {
	ids := make([]string, len(echos))
	for i, e := range echos {
		ids[i] = e.ID
	}
	return ids
}

func TestEchoRepository_QueryEchos_PrivateFilter(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e-pub", "public one", false, 0, 100)
	seedEcho(t, db, "e-prv", "private one", true, 0, 200)

	t.Run("showPrivate=false excludes private", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(commonModel.EchoQueryDto{Page: 1, PageSize: 10}, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, echos, 1)
		assert.Equal(t, "e-pub", echos[0].ID)
	})

	t.Run("showPrivate=true includes private", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(commonModel.EchoQueryDto{Page: 1, PageSize: 10}, true)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, echos, 2)
	})

	ptr := func(b bool) *bool { return &b }

	t.Run("showPrivate=true + Private=true narrows to private only", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 1, PageSize: 10, Private: ptr(true)},
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, echos, 1)
		assert.Equal(t, "e-prv", echos[0].ID)
	})

	t.Run("showPrivate=true + Private=false narrows to public only", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 1, PageSize: 10, Private: ptr(false)},
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, echos, 1)
		assert.Equal(t, "e-pub", echos[0].ID)
	})

	t.Run("showPrivate=false ignores Private=true (anti-leak)", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 1, PageSize: 10, Private: ptr(true)},
			false,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, echos, 1)
		assert.Equal(t, "e-pub", echos[0].ID)
	})
}

func TestEchoRepository_QueryEchos_TagJoinDistinctCount(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e1", "tagged twice", false, 0, 100)
	seedEcho(t, db, "e2", "untagged", false, 0, 200)
	seedTag(t, db, "t1", "alpha")
	seedTag(t, db, "t2", "beta")
	linkTag(t, db, "e1", "t1")
	linkTag(t, db, "e1", "t2")

	echos, total, err := repo.QueryEchos(
		commonModel.EchoQueryDto{Page: 1, PageSize: 10, TagIDs: []string{"t1", "t2"}},
		true,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total, "DISTINCT 应把同一 echo 的多标签命中收敛为 1")
	require.Len(t, echos, 1)
	assert.Equal(t, "e1", echos[0].ID)
	assert.Len(t, echos[0].Tags, 2)
}

func TestEchoRepository_QueryEchos_SearchLike(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e1", "golang is great", false, 0, 100)
	seedEcho(t, db, "e2", "vue is nice", false, 0, 200)

	echos, total, err := repo.QueryEchos(
		commonModel.EchoQueryDto{Page: 1, PageSize: 10, Search: "golang"},
		true,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, echos, 1)
	assert.Equal(t, "e1", echos[0].ID)
}

func TestEchoRepository_QueryEchos_DateRange(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e-old", "old", false, 0, 1000)
	seedEcho(t, db, "e-mid", "mid", false, 0, 2000)
	seedEcho(t, db, "e-new", "new", false, 0, 3000)

	t.Run("DateFrom inclusive lower bound", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 1, PageSize: 10, DateFrom: 2000},
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.ElementsMatch(t, []string{"e-mid", "e-new"}, echoIDs(echos))
	})

	t.Run("DateTo inclusive upper bound", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 1, PageSize: 10, DateTo: 2000},
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.ElementsMatch(t, []string{"e-old", "e-mid"}, echoIDs(echos))
	})

	t.Run("DateFrom and DateTo bound a closed window", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 1, PageSize: 10, DateFrom: 2000, DateTo: 2000},
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, echos, 1)
		assert.Equal(t, "e-mid", echos[0].ID)
	})
}

func TestEchoRepository_QueryEchos_Pagination(t *testing.T) {
	repo, db := newEchoRepo(t)
	for i := 1; i <= 5; i++ {
		seedEcho(t, db, "e"+string(rune('0'+i)), "c", false, 0, int64(i*100))
	}

	t.Run("page 1 size 2", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(commonModel.EchoQueryDto{Page: 1, PageSize: 2}, true)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		require.Len(t, echos, 2)
		assert.Equal(t, []string{"e5", "e4"}, echoIDs(echos))
	})

	t.Run("page 2 size 2 applies offset", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(commonModel.EchoQueryDto{Page: 2, PageSize: 2}, true)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		require.Len(t, echos, 2)
		assert.Equal(t, []string{"e3", "e2"}, echoIDs(echos))
	})

	t.Run("page past the end returns empty but real total", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(commonModel.EchoQueryDto{Page: 4, PageSize: 2}, true)
		require.NoError(t, err)
		assert.Equal(t, int64(5), total)
		assert.Empty(t, echos)
	})
}

func TestEchoRepository_QueryEchos_SortWhitelist(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e-low-fav-new", "a", false, 1, 300)
	seedEcho(t, db, "e-mid-fav-mid", "b", false, 5, 200)
	seedEcho(t, db, "e-high-fav-old", "c", false, 9, 100)

	cases := []struct {
		name      string
		sortBy    string
		sortOrder string
		want      []string
	}{
		{"created_at desc (default)", "", "", []string{"e-low-fav-new", "e-mid-fav-mid", "e-high-fav-old"}},
		{"created_at asc", "created_at", "asc", []string{"e-high-fav-old", "e-mid-fav-mid", "e-low-fav-new"}},
		{"fav_count desc", "fav_count", "desc", []string{"e-high-fav-old", "e-mid-fav-mid", "e-low-fav-new"}},
		{"fav_count asc", "fav_count", "asc", []string{"e-low-fav-new", "e-mid-fav-mid", "e-high-fav-old"}},
		{"unknown sortBy falls back to created_at", "bogus_column", "desc", []string{"e-low-fav-new", "e-mid-fav-mid", "e-high-fav-old"}},
		{"unknown sortOrder falls back to desc", "fav_count", "sideways", []string{"e-high-fav-old", "e-mid-fav-mid", "e-low-fav-new"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			echos, _, err := repo.QueryEchos(
				commonModel.EchoQueryDto{Page: 1, PageSize: 10, SortBy: tc.sortBy, SortOrder: tc.sortOrder},
				true,
			)
			require.NoError(t, err)
			assert.Equal(t, tc.want, echoIDs(echos))
		})
	}
}

func TestEchoRepository_QueryEchos_EmptyTagShortCircuit(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e1", "no tags here", false, 0, 100)
	seedTag(t, db, "t-real", "real")

	t.Run("tag filter with zero matches returns empty non-nil slice", func(t *testing.T) {
		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 1, PageSize: 10, TagIDs: []string{"t-real"}},
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		require.NotNil(t, echos)
		assert.Empty(t, echos)
	})

	t.Run("tag filter with offset past results short-circuits to empty with real total", func(t *testing.T) {
		seedEcho(t, db, "e2", "tagged", false, 0, 200)
		linkTag(t, db, "e2", "t-real")

		echos, total, err := repo.QueryEchos(
			commonModel.EchoQueryDto{Page: 2, PageSize: 10, TagIDs: []string{"t-real"}},
			true,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Empty(t, echos)
	})
}

func TestEchoRepository_QueryEchos_UserIDFilter(t *testing.T) {
	repo, db := newEchoRepo(t)
	require.NoError(t, db.Create(&echoModel.Echo{ID: "e-alice", Content: "a", UserID: "alice", CreatedAt: 100}).Error)
	require.NoError(t, db.Create(&echoModel.Echo{ID: "e-bob", Content: "b", UserID: "bob", CreatedAt: 200}).Error)

	echos, total, err := repo.QueryEchos(
		commonModel.EchoQueryDto{Page: 1, PageSize: 10, UserID: "alice"},
		true,
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, echos, 1)
	assert.Equal(t, "e-alice", echos[0].ID)
}
