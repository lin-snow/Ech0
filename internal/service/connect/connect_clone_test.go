// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"testing"

	model "github.com/lin-snow/ech0/internal/model/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneConnects(t *testing.T) {
	t.Run("nil input returns non-nil empty slice", func(t *testing.T) {
		got := cloneConnects(nil)
		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("empty input returns non-nil empty slice", func(t *testing.T) {
		got := cloneConnects([]model.Connect{})
		require.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("copies values verbatim", func(t *testing.T) {
		src := []model.Connect{
			{ServerName: "a", ServerURL: "https://a", TotalEchos: 1},
			{ServerName: "b", ServerURL: "https://b", TotalEchos: 2},
		}
		got := cloneConnects(src)
		assert.Equal(t, src, got)
	})

	t.Run("mutating clone does not affect source", func(t *testing.T) {
		src := []model.Connect{{ServerName: "a"}}
		got := cloneConnects(src)
		require.Len(t, got, 1)

		got[0].ServerName = "mutated"
		assert.Equal(t, "a", src[0].ServerName, "mutating the clone must not leak back into source")
	})

	t.Run("append to clone does not grow source", func(t *testing.T) {
		src := []model.Connect{{ServerName: "a"}}
		got := cloneConnects(src)
		got = append(got, model.Connect{ServerName: "b"})

		assert.Len(t, src, 1)
		assert.Len(t, got, 2)
	})
}
