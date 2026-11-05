// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	actorUtil "github.com/lin-snow/ech0/internal/util/actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestReactionHydration_IsPerActor is the regression test for the shared echo
// cache: the cached echo must stay actor-neutral while each request sees its
// own reacted flags.
func TestReactionHydration_IsPerActor(t *testing.T) {
	db := helpers.NewTestDB(t)
	repo := echoRepository.NewEchoRepository(func() *gorm.DB { return db }, helpers.NewTestCache())
	svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)

	require.NoError(t, db.Create(&echoModel.Echo{
		ID: echoID, Content: "public post", UserID: "u1",
	}).Error)

	ctxA := actorUtil.WithContext(helpers.CtxAnonymous(), "v:actor-a")
	ctxB := actorUtil.WithContext(helpers.CtxAnonymous(), "v:actor-b")

	counts, err := svc.ReactToEcho(ctxA, echoID, "like")
	require.NoError(t, err)
	require.Len(t, counts, 1)
	assert.Equal(t, 1, counts[0].Count)
	assert.True(t, counts[0].Reacted)

	t.Run("reacting twice does not double count", func(t *testing.T) {
		counts, err := svc.ReactToEcho(ctxA, echoID, "like")
		require.NoError(t, err)
		require.Len(t, counts, 1)
		assert.Equal(t, 1, counts[0].Count)
	})

	t.Run("each actor sees its own reacted flag", func(t *testing.T) {
		for range 2 {
			got, err := svc.GetEchoById(ctxA, echoID)
			require.NoError(t, err)
			require.Len(t, got.Reactions, 1)
			assert.True(t, got.Reactions[0].Reacted)
		}

		got, err := svc.GetEchoById(ctxB, echoID)
		require.NoError(t, err)
		require.Len(t, got.Reactions, 1)
		assert.Equal(t, 1, got.Reactions[0].Count)
		assert.False(t, got.Reactions[0].Reacted)
	})

	t.Run("cache holds no reaction state", func(t *testing.T) {
		cached, err := repo.GetEchosById(ctxA, echoID)
		require.NoError(t, err)
		require.NotNil(t, cached)
		assert.Empty(t, cached.Reactions)
	})

	t.Run("removing clears the aggregate", func(t *testing.T) {
		counts, err := svc.UnreactEcho(ctxA, echoID, "like")
		require.NoError(t, err)
		assert.Empty(t, counts)

		got, err := svc.GetEchoById(ctxA, echoID)
		require.NoError(t, err)
		assert.Empty(t, got.Reactions)
	})
}
