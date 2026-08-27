// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"errors"
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEchoRepository_IncrementTagUsageCount(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedTag(t, db, "t1", "golang")

	readUsage := func(id string) int {
		t.Helper()
		var tag echoModel.Tag
		require.NoError(t, db.First(&tag, "id = ?", id).Error)
		return tag.UsageCount
	}

	t.Run("increments by one each call", func(t *testing.T) {
		require.NoError(t, repo.IncrementTagUsageCount(context.Background(), "t1"))
		assert.Equal(t, 1, readUsage("t1"))

		require.NoError(t, repo.IncrementTagUsageCount(context.Background(), "t1"))
		assert.Equal(t, 2, readUsage("t1"))
	})

	t.Run("missing tag is a no-op without error", func(t *testing.T) {
		err := repo.IncrementTagUsageCount(context.Background(), "does-not-exist")
		require.NoError(t, err)
	})
}

func TestEchoRepository_GetTagsByNames(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedTag(t, db, "t1", "alpha")
	seedTag(t, db, "t2", "beta")
	seedTag(t, db, "t3", "gamma")

	t.Run("returns matching tags only", func(t *testing.T) {
		tags, err := repo.GetTagsByNames(context.Background(), []string{"alpha", "gamma"})
		require.NoError(t, err)
		require.Len(t, tags, 2)
		names := []string{tags[0].Name, tags[1].Name}
		assert.ElementsMatch(t, []string{"alpha", "gamma"}, names)
	})

	t.Run("unknown names yield empty result", func(t *testing.T) {
		tags, err := repo.GetTagsByNames(context.Background(), []string{"nope"})
		require.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("empty name list yields empty result", func(t *testing.T) {
		tags, err := repo.GetTagsByNames(context.Background(), []string{})
		require.NoError(t, err)
		assert.Empty(t, tags)
	})
}

func TestEchoRepository_DeleteTagById(t *testing.T) {
	t.Run("deletes tag and its echo_tags join rows", func(t *testing.T) {
		repo, db := newEchoRepo(t)
		seedEcho(t, db, "e1", "tagged", false, 0, 100)
		seedTag(t, db, "t1", "alpha")
		linkTag(t, db, "e1", "t1")

		require.NoError(t, repo.DeleteTagById(context.Background(), "t1"))

		var tagCount int64
		require.NoError(t, db.Model(&echoModel.Tag{}).Where("id = ?", "t1").Count(&tagCount).Error)
		assert.Equal(t, int64(0), tagCount, "tag row should be gone")

		var joinCount int64
		require.NoError(t, db.Model(&echoModel.EchoTag{}).Where("tag_id = ?", "t1").Count(&joinCount).Error)
		assert.Equal(t, int64(0), joinCount, "echo_tags rows for the tag should be gone")
	})

	t.Run("deleting a missing tag returns ErrRecordNotFound", func(t *testing.T) {
		repo, _ := newEchoRepo(t)
		err := repo.DeleteTagById(context.Background(), "ghost")
		require.Error(t, err)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}
