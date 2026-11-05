// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoRepository_Reactions(t *testing.T) {
	repo, db := newEchoRepo(t)
	seedEcho(t, db, "e1", "reactions", false, 100)
	seedEcho(t, db, "e2", "other", false, 200)
	ctx := context.Background()

	t.Run("add is idempotent per actor", func(t *testing.T) {
		require.NoError(t, repo.AddReaction(ctx, "e1", "like", "v:a"))
		require.NoError(t, repo.AddReaction(ctx, "e1", "like", "v:a"))

		counts, err := repo.ListReactions(ctx, []string{"e1"}, "v:a")
		require.NoError(t, err)
		require.Len(t, counts["e1"], 1)
		assert.Equal(t, 1, counts["e1"][0].Count)
		assert.True(t, counts["e1"][0].Reacted)
	})

	t.Run("different actors accumulate, each sees own state", func(t *testing.T) {
		require.NoError(t, repo.AddReaction(ctx, "e1", "like", "v:b"))

		counts, err := repo.ListReactions(ctx, []string{"e1"}, "v:b")
		require.NoError(t, err)
		require.Len(t, counts["e1"], 1)
		assert.Equal(t, 2, counts["e1"][0].Count)
		assert.True(t, counts["e1"][0].Reacted)

		other, err := repo.ListReactions(ctx, []string{"e1"}, "v:c")
		require.NoError(t, err)
		require.Len(t, other["e1"], 1)
		assert.False(t, other["e1"][0].Reacted)
	})

	t.Run("same actor can hold several emoji", func(t *testing.T) {
		require.NoError(t, repo.AddReaction(ctx, "e1", "party", "v:a"))

		counts, err := repo.ListReactions(ctx, []string{"e1"}, "v:a")
		require.NoError(t, err)
		require.Len(t, counts["e1"], 2)
		// Display order follows the reaction catalog, not insertion order.
		assert.Equal(t, []string{"like", "party"}, []string{counts["e1"][0].Key, counts["e1"][1].Key})
	})

	t.Run("remove only affects the given actor", func(t *testing.T) {
		require.NoError(t, repo.RemoveReaction(ctx, "e1", "like", "v:a"))
		require.NoError(t, repo.RemoveReaction(ctx, "e1", "like", "v:a"))

		counts, err := repo.ListReactions(ctx, []string{"e1"}, "v:a")
		require.NoError(t, err)
		require.Len(t, counts["e1"], 2)
		assert.Equal(t, "like", counts["e1"][0].Key)
		assert.Equal(t, 1, counts["e1"][0].Count)
		assert.False(t, counts["e1"][0].Reacted, "v:a removed its like")
		assert.Equal(t, "party", counts["e1"][1].Key)
		assert.True(t, counts["e1"][1].Reacted)

		still, err := repo.ListReactions(ctx, []string{"e1"}, "v:b")
		require.NoError(t, err)
		require.Len(t, still["e1"], 2)
		assert.Equal(t, "like", still["e1"][0].Key)
		assert.Equal(t, 1, still["e1"][0].Count)
		assert.True(t, still["e1"][0].Reacted, "v:b's like is untouched")
		assert.Equal(t, "party", still["e1"][1].Key)
		assert.False(t, still["e1"][1].Reacted)
	})

	t.Run("same emoji on different echoes stays separate", func(t *testing.T) {
		require.NoError(t, repo.AddReaction(ctx, "e2", "like", "v:a"))

		counts, err := repo.ListReactions(ctx, []string{"e1", "e2"}, "v:a")
		require.NoError(t, err)
		require.Len(t, counts["e2"], 1)
		assert.Equal(t, 1, counts["e2"][0].Count)
		assert.Equal(t, 1, counts["e1"][0].Count)
	})

	t.Run("echoes without reactions are absent from the map", func(t *testing.T) {
		seedEcho(t, db, "e3", "silent", false, 300)
		counts, err := repo.ListReactions(ctx, []string{"e3"}, "v:a")
		require.NoError(t, err)
		assert.NotContains(t, counts, "e3")
	})
}

func TestReactionCatalogIsNonEmpty(t *testing.T) {
	require.NotEmpty(t, echoModel.ReactionEmojis)
	for _, emoji := range echoModel.ReactionEmojis {
		assert.NotEmpty(t, emoji.Key)
		assert.NotEmpty(t, emoji.Char)
		_, ok := echoModel.ReactionEmojiByKey(emoji.Key)
		assert.True(t, ok)
	}
	_, ok := echoModel.ReactionEmojiByKey("unknown")
	assert.False(t, ok)
}
