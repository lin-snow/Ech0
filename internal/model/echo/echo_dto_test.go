// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoUpsertDto_ToModel(t *testing.T) {
	ts := int64(1_700_000_000)

	t.Run("nil_created_at_leaves_zero_value", func(t *testing.T) {
		dto := &EchoUpsertDto{ID: "e1", Content: "hi", CreatedAt: nil}
		got := dto.ToModel()
		require.NotNil(t, got)
		assert.Equal(t, int64(0), got.CreatedAt)
	})

	t.Run("non_nil_created_at_is_dereferenced", func(t *testing.T) {
		dto := &EchoUpsertDto{ID: "e1", Content: "hi", CreatedAt: &ts}
		got := dto.ToModel()
		assert.Equal(t, ts, got.CreatedAt)
	})

	t.Run("nil_extension_stays_nil", func(t *testing.T) {
		dto := &EchoUpsertDto{ID: "e1", Extension: nil}
		got := dto.ToModel()
		assert.Nil(t, got.Extension)
	})

	t.Run("non_nil_extension_is_projected", func(t *testing.T) {
		payload := map[string]any{"url": "https://x", "n": 1}
		dto := &EchoUpsertDto{
			ID:        "e1",
			Extension: &EchoExtensionDto{Type: Extension_MUSIC, Payload: payload},
		}
		got := dto.ToModel()
		require.NotNil(t, got.Extension)
		assert.Equal(t, Extension_MUSIC, got.Extension.Type)
		assert.Equal(t, payload, got.Extension.Payload)
		assert.Equal(t, "", got.Extension.ID)
		assert.Equal(t, "", got.Extension.EchoID)
		assert.Equal(t, int64(0), got.Extension.CreatedAt)
	})

	t.Run("scalar_and_slice_fields_pass_through", func(t *testing.T) {
		files := []EchoFile{{ID: "f1", FileID: "file-1", SortOrder: 2}}
		tags := []Tag{{ID: "t1", Name: "go"}}
		dto := &EchoUpsertDto{
			ID:        "e9",
			Content:   "body",
			EchoFiles: files,
			Layout:    LayoutGrid,
			Private:   true,
			Tags:      tags,
		}
		got := dto.ToModel()
		assert.Equal(t, "e9", got.ID)
		assert.Equal(t, "body", got.Content)
		assert.Equal(t, LayoutGrid, got.Layout)
		assert.True(t, got.Private)
		assert.Equal(t, files, got.EchoFiles)
		assert.Equal(t, tags, got.Tags)
	})

	t.Run("zero_value_dto_maps_to_empty_model", func(t *testing.T) {
		got := (&EchoUpsertDto{}).ToModel()
		require.NotNil(t, got)
		assert.Equal(t, "", got.ID)
		assert.Equal(t, "", got.Content)
		assert.False(t, got.Private)
		assert.Equal(t, int64(0), got.CreatedAt)
		assert.Nil(t, got.Extension)
		assert.Nil(t, got.EchoFiles)
		assert.Nil(t, got.Tags)
	})
}
