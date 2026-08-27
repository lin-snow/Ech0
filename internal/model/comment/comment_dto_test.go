// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFullComment() Comment {
	parent := "parent-1"
	uid := "user-1"
	return Comment{
		ID:        "c1",
		EchoID:    "e1",
		ParentID:  &parent,
		UserID:    &uid,
		Nickname:  "alice",
		Email:     "alice@example.com",
		Website:   "https://alice.dev",
		Content:   "hello",
		Status:    StatusApproved,
		Hot:       true,
		IPHash:    "deadbeef",
		UserAgent: "curl/8",
		Source:    SourceGuest,
		CreatedAt: 100,
		UpdatedAt: 200,
	}
}

func TestToPublicComment(t *testing.T) {
	t.Run("projects_public_fields_only", func(t *testing.T) {
		c := newFullComment()
		got := ToPublicComment(c)

		assert.Equal(t, c.ID, got.ID)
		assert.Equal(t, c.EchoID, got.EchoID)
		assert.Equal(t, c.ParentID, got.ParentID)
		assert.Equal(t, c.Nickname, got.Nickname)
		assert.Equal(t, c.Website, got.Website)
		assert.Equal(t, c.Content, got.Content)
		assert.Equal(t, c.Status, got.Status)
		assert.Equal(t, c.Hot, got.Hot)
		assert.Equal(t, c.Source, got.Source)
		assert.Equal(t, c.CreatedAt, got.CreatedAt)
		assert.Equal(t, c.UpdatedAt, got.UpdatedAt)
	})

	t.Run("parent_id_pointer_is_carried_through", func(t *testing.T) {
		c := newFullComment()
		got := ToPublicComment(c)
		require.NotNil(t, got.ParentID)
		assert.Same(t, c.ParentID, got.ParentID)
	})

	t.Run("nil_parent_id_stays_nil", func(t *testing.T) {
		c := newFullComment()
		c.ParentID = nil
		got := ToPublicComment(c)
		assert.Nil(t, got.ParentID)
	})
}

func TestToPublicComments(t *testing.T) {
	t.Run("nil_input_returns_non_nil_empty_slice", func(t *testing.T) {
		got := ToPublicComments(nil)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})

	t.Run("empty_input_returns_empty_slice", func(t *testing.T) {
		got := ToPublicComments([]Comment{})
		assert.NotNil(t, got)
		assert.Len(t, got, 0)
	})

	t.Run("preserves_length_and_order", func(t *testing.T) {
		in := []Comment{
			{ID: "a", Nickname: "n-a", Content: "ca", Status: StatusPending, Source: SourceGuest},
			{ID: "b", Nickname: "n-b", Content: "cb", Status: StatusApproved, Source: SourceSystem},
			{ID: "c", Nickname: "n-c", Content: "cc", Status: StatusRejected, Source: SourceIntegration},
		}
		got := ToPublicComments(in)
		require.Len(t, got, len(in))
		for i := range in {
			assert.Equal(t, in[i].ID, got[i].ID)
			assert.Equal(t, in[i].Nickname, got[i].Nickname)
			assert.Equal(t, in[i].Content, got[i].Content)
			assert.Equal(t, in[i].Status, got[i].Status)
			assert.Equal(t, in[i].Source, got[i].Source)
		}
	})

	t.Run("each_element_matches_to_public_comment", func(t *testing.T) {
		in := []Comment{newFullComment()}
		got := ToPublicComments(in)
		require.Len(t, got, 1)
		assert.Equal(t, ToPublicComment(in[0]), got[0])
	})
}
