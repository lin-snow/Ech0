// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	embeddingmock "github.com/lin-snow/ech0/internal/test/mocks/embeddingmock"
	kvmock "github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func TestBuildText(t *testing.T) {
	cases := []struct {
		name string
		echo echoModel.Echo
		want string
	}{
		{"content only", echoModel.Echo{Content: "hello"}, "hello"},
		{"content trimmed", echoModel.Echo{Content: "  hi  "}, "hi"},
		{"empty no tags", echoModel.Echo{}, ""},
		{"whitespace no tags", echoModel.Echo{Content: "   "}, ""},
		{
			"content with tags",
			echoModel.Echo{Content: "hello", Tags: []echoModel.Tag{{Name: "go"}, {Name: "vue"}}},
			"hello go, vue",
		},
		{
			"empty content with tags falls back to tags",
			echoModel.Echo{Tags: []echoModel.Tag{{Name: "go"}}},
			"go",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, buildText(tc.echo))
		})
	}
}

func TestHashContent(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		assert.Equal(t, hashContent("same"), hashContent("same"))
	})
	t.Run("distinct inputs differ", func(t *testing.T) {
		assert.NotEqual(t, hashContent("a"), hashContent("b"))
	})
	t.Run("matches stdlib sha256 hex", func(t *testing.T) {
		sum := sha256.Sum256([]byte("hello world"))
		assert.Equal(t, hex.EncodeToString(sum[:]), hashContent("hello world"))
	})
}

func newSvcForEnsure(t *testing.T) (*EmbeddingService, *embeddingmock.MockRepository, *kvmock.MockStore) {
	t.Helper()
	repo := embeddingmock.NewMockRepository(t)
	kv := kvmock.NewMockStore(t)
	return NewEmbeddingService(repo, kv, nil), repo, kv
}

func TestEnsureReady_AlreadyReady(t *testing.T) {
	ctx := context.Background()
	setting := settingModel.EmbeddingSetting{Model: "m1", Dim: 768}
	stateJSON := mustJSON(t, embModel.IndexState{Model: "m1", Dim: 768})

	t.Run("ensure table ok", func(t *testing.T) {
		svc, repo, kv := newSvcForEnsure(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return(stateJSON, nil).Once()
		repo.EXPECT().EnsureVecTable(ctx, 768).Return(nil).Once()
		require.NoError(t, svc.ensureReady(ctx, setting))
	})

	t.Run("ensure table error propagates", func(t *testing.T) {
		svc, repo, kv := newSvcForEnsure(t)
		boom := errors.New("ensure boom")
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return(stateJSON, nil).Once()
		repo.EXPECT().EnsureVecTable(ctx, 768).Return(boom).Once()
		require.ErrorIs(t, svc.ensureReady(ctx, setting), boom)
	})
}

func TestEnsureReady_Rebuild(t *testing.T) {
	ctx := context.Background()
	setting := settingModel.EmbeddingSetting{Model: "m1", Dim: 768}
	newState := mustJSON(t, embModel.IndexState{Model: "m1", Dim: 768})

	cases := []struct {
		name     string
		stateRaw string
		stateErr error
	}{
		{"missing state", "", kvstore.ErrNotFound},
		{"corrupt state json", "not-json", nil},
		{"dim changed", mustJSON(t, embModel.IndexState{Model: "m1", Dim: 512}), nil},
		{"model changed", mustJSON(t, embModel.IndexState{Model: "old", Dim: 768}), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, kv := newSvcForEnsure(t)
			kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return(tc.stateRaw, tc.stateErr).Once()
			repo.EXPECT().DropVecTable(ctx).Return(nil).Once()
			repo.EXPECT().ClearAll(ctx).Return(nil).Once()
			repo.EXPECT().EnsureVecTable(ctx, 768).Return(nil).Once()
			kv.EXPECT().Set(ctx, commonModel.EmbeddingIndexStateKey, newState).Return(nil).Once()
			require.NoError(t, svc.ensureReady(ctx, setting))
		})
	}
}

func TestEnsureReady_RebuildErrors(t *testing.T) {
	ctx := context.Background()
	setting := settingModel.EmbeddingSetting{Model: "m1", Dim: 768}
	boom := errors.New("step boom")

	t.Run("drop fails", func(t *testing.T) {
		svc, repo, kv := newSvcForEnsure(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return("", kvstore.ErrNotFound).Once()
		repo.EXPECT().DropVecTable(ctx).Return(boom).Once()
		require.ErrorIs(t, svc.ensureReady(ctx, setting), boom)
	})

	t.Run("clear all fails", func(t *testing.T) {
		svc, repo, kv := newSvcForEnsure(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return("", kvstore.ErrNotFound).Once()
		repo.EXPECT().DropVecTable(ctx).Return(nil).Once()
		repo.EXPECT().ClearAll(ctx).Return(boom).Once()
		require.ErrorIs(t, svc.ensureReady(ctx, setting), boom)
	})

	t.Run("ensure table fails", func(t *testing.T) {
		svc, repo, kv := newSvcForEnsure(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return("", kvstore.ErrNotFound).Once()
		repo.EXPECT().DropVecTable(ctx).Return(nil).Once()
		repo.EXPECT().ClearAll(ctx).Return(nil).Once()
		repo.EXPECT().EnsureVecTable(ctx, 768).Return(boom).Once()
		require.ErrorIs(t, svc.ensureReady(ctx, setting), boom)
	})

	t.Run("persist state fails", func(t *testing.T) {
		svc, repo, kv := newSvcForEnsure(t)
		newState := mustJSON(t, embModel.IndexState{Model: "m1", Dim: 768})
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return("", kvstore.ErrNotFound).Once()
		repo.EXPECT().DropVecTable(ctx).Return(nil).Once()
		repo.EXPECT().ClearAll(ctx).Return(nil).Once()
		repo.EXPECT().EnsureVecTable(ctx, 768).Return(nil).Once()
		kv.EXPECT().Set(ctx, commonModel.EmbeddingIndexStateKey, newState).Return(boom).Once()
		require.ErrorIs(t, svc.ensureReady(ctx, setting), boom)
	})
}
