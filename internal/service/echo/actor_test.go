// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/internal/test/helpers"
	actorUtil "github.com/lin-snow/ech0/internal/util/actor"
	"github.com/stretchr/testify/assert"
)

func TestReactionActor(t *testing.T) {
	t.Run("authenticated user wins over the visitor id", func(t *testing.T) {
		ctx := actorUtil.WithContext(helpers.CtxAsUser("u1"), "v:anon")
		assert.Equal(t, "u:u1", reactionActor(ctx))
	})

	t.Run("visitor id is used for anonymous requests", func(t *testing.T) {
		ctx := actorUtil.WithContext(helpers.CtxAnonymous(), "v:anon")
		assert.Equal(t, "v:anon", reactionActor(ctx))
	})

	t.Run("contexts without identity share the fallback actor", func(t *testing.T) {
		assert.Equal(t, "ip:unknown", reactionActor(context.Background()))
	})
}
