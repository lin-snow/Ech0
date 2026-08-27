// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/embedding"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	embeddingmock "github.com/lin-snow/ech0/internal/test/mocks/embeddingmock"
	kvmock "github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testModel = "text-embedding-3-small"
	testDim   = 1536
)

func enabledSettingJSON(t *testing.T) string {
	t.Helper()
	b, err := json.Marshal(settingModel.EmbeddingSetting{
		Enable: true,
		Model:  testModel,
		Dim:    testDim,
	})
	require.NoError(t, err)
	return string(b)
}

func contentHash(t *testing.T, content string) string {
	t.Helper()
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func newSvc(t *testing.T) (*embeddingService.EmbeddingService, *embeddingmock.MockRepository, *kvmock.MockStore) {
	t.Helper()
	repo := embeddingmock.NewMockRepository(t)
	kv := kvmock.NewMockStore(t)
	svc := embeddingService.NewEmbeddingService(repo, kv, nil)
	return svc, repo, kv
}

func TestEnabled(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name   string
		raw    string
		getErr error
		want   bool
	}{
		{"fully configured", enabledSettingJSON(t), nil, true},
		{"setting missing -> default disabled", "", kvstore.ErrNotFound, false},
		{"kv backend error -> false", "", errors.New("kv boom"), false},
		{
			"enabled but model empty -> false",
			mustSettingJSON(t, settingModel.EmbeddingSetting{Enable: true, Model: "", Dim: testDim}),
			nil,
			false,
		},
		{
			"enabled but dim zero -> false",
			mustSettingJSON(t, settingModel.EmbeddingSetting{Enable: true, Model: testModel, Dim: 0}),
			nil,
			false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, kv := newSvc(t)
			kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(tc.raw, tc.getErr).Once()
			assert.Equal(t, tc.want, svc.Enabled(ctx))
		})
	}
}

func mustSettingJSON(t *testing.T, s settingModel.EmbeddingSetting) string {
	t.Helper()
	b, err := json.Marshal(s)
	require.NoError(t, err)
	return string(b)
}

func TestIndexEcho_GetSettingError(t *testing.T) {
	ctx := context.Background()
	svc, _, kv := newSvc(t)
	boom := errors.New("kv backend down")
	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return("", boom).Once()

	err := svc.IndexEcho(ctx, echoModel.Echo{ID: "e1", Content: "hello"})
	require.ErrorIs(t, err, boom)
}

func TestIndexEcho_NotEnabled(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name   string
		raw    string
		getErr error
	}{
		{"missing setting", "", kvstore.ErrNotFound},
		{"explicitly disabled", mustSettingJSON(t, settingModel.EmbeddingSetting{Enable: false, Model: testModel, Dim: testDim}), nil},
		{"model unset", mustSettingJSON(t, settingModel.EmbeddingSetting{Enable: true, Model: "", Dim: testDim}), nil},
		{"dim non-positive", mustSettingJSON(t, settingModel.EmbeddingSetting{Enable: true, Model: testModel, Dim: 0}), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, kv := newSvc(t)
			kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(tc.raw, tc.getErr).Once()
			require.NoError(t, svc.IndexEcho(ctx, echoModel.Echo{ID: "e1", Content: "hello"}))
		})
	}
}

func TestIndexEcho_EmptyText_Deletes(t *testing.T) {
	ctx := context.Background()

	t.Run("empty content with no tags deletes existing index", func(t *testing.T) {
		svc, repo, kv := newSvc(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
		repo.EXPECT().Delete(ctx, "e-empty").Return(nil).Once()
		require.NoError(t, svc.IndexEcho(ctx, echoModel.Echo{ID: "e-empty", Content: "   "}))
	})

	t.Run("delete error propagates", func(t *testing.T) {
		svc, repo, kv := newSvc(t)
		boom := errors.New("delete boom")
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
		repo.EXPECT().Delete(ctx, "e-empty").Return(boom).Once()
		require.ErrorIs(t, svc.IndexEcho(ctx, echoModel.Echo{ID: "e-empty", Content: ""}), boom)
	})
}

func TestIndexEcho_DedupSkip(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv := newSvc(t)

	const content = "stable content"
	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	repo.EXPECT().GetMeta(ctx, "e1").Return(&embModel.EchoEmbedding{
		EchoID:      "e1",
		ContentHash: contentHash(t, content),
		Model:       testModel,
		Dim:         testDim,
	}, true, nil).Once()

	require.NoError(t, svc.IndexEcho(ctx, echoModel.Echo{ID: "e1", Content: content}))
}

func TestIndexEcho_ContentChanged_ProceedsPastDedup(t *testing.T) {
	ctx := context.Background()
	boom := errors.New("ensure boom")

	cases := []struct {
		name string
		meta *embModel.EchoEmbedding
		ok   bool
	}{
		{
			"hash differs",
			&embModel.EchoEmbedding{EchoID: "e1", ContentHash: "stale", Model: testModel, Dim: testDim},
			true,
		},
		{
			"model differs",
			&embModel.EchoEmbedding{EchoID: "e1", ContentHash: contentHash(t, "v2 content"), Model: "other-model", Dim: testDim},
			true,
		},
		{
			"dim differs",
			&embModel.EchoEmbedding{EchoID: "e1", ContentHash: contentHash(t, "v2 content"), Model: testModel, Dim: 64},
			true,
		},
		{"no existing meta", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, kv := newSvc(t)
			kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
			repo.EXPECT().GetMeta(ctx, "e1").Return(tc.meta, tc.ok, nil).Once()
			kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).
				Return(mustJSONState(t, testModel, testDim), nil).Once()
			repo.EXPECT().EnsureVecTable(ctx, testDim).Return(boom).Once()

			err := svc.IndexEcho(ctx, echoModel.Echo{ID: "e1", Content: "v2 content"})
			require.ErrorIs(t, err, boom)
		})
	}
}

func mustJSONState(t *testing.T, model string, dim int) string {
	t.Helper()
	b, err := json.Marshal(embModel.IndexState{Model: model, Dim: dim})
	require.NoError(t, err)
	return string(b)
}

func TestRemoveEcho(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to repository", func(t *testing.T) {
		svc, repo, _ := newSvc(t)
		repo.EXPECT().Delete(ctx, "e1").Return(nil).Once()
		require.NoError(t, svc.RemoveEcho(ctx, "e1"))
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc, repo, _ := newSvc(t)
		boom := errors.New("delete boom")
		repo.EXPECT().Delete(ctx, "e1").Return(boom).Once()
		require.ErrorIs(t, svc.RemoveEcho(ctx, "e1"), boom)
	})
}

func TestBackfill_PreSeamGates(t *testing.T) {
	t.Run("setting backend error surfaces", func(t *testing.T) {
		ctx := context.Background()
		svc, _, kv := newSvc(t)
		boom := errors.New("kv backend down")
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return("", boom).Once()
		res, err := svc.Backfill(ctx, nil)
		require.ErrorIs(t, err, boom)
		assert.Equal(t, embeddingService.BackfillResult{}, res)
	})

	t.Run("not enabled returns ErrNotEnabled", func(t *testing.T) {
		ctx := context.Background()
		svc, _, kv := newSvc(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return("", kvstore.ErrNotFound).Once()
		_, err := svc.Backfill(ctx, nil)
		require.ErrorIs(t, err, embedding.ErrNotEnabled)
	})

	t.Run("ensureReady error aborts before paging", func(t *testing.T) {
		ctx := context.Background()
		svc, repo, kv := newSvc(t)
		boom := errors.New("drop boom")
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).Return("", kvstore.ErrNotFound).Once()
		repo.EXPECT().DropVecTable(ctx).Return(boom).Once()
		_, err := svc.Backfill(ctx, nil)
		require.ErrorIs(t, err, boom)
	})

	t.Run("cancelled context stops at loop guard", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		svc, repo, kv := newSvc(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
		kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).
			Return(mustJSONState(t, testModel, testDim), nil).Once()
		repo.EXPECT().EnsureVecTable(ctx, testDim).Return(nil).Once()
		_, err := svc.Backfill(ctx, nil)
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestSearch_PreSeamGates(t *testing.T) {
	ctx := context.Background()

	t.Run("not enabled returns ErrNotEnabled", func(t *testing.T) {
		svc, _, kv := newSvc(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return("", kvstore.ErrNotFound).Once()
		res, err := svc.Search(ctx, "q", 5, "")
		require.ErrorIs(t, err, embedding.ErrNotEnabled)
		assert.Nil(t, res)
	})

	t.Run("disabled setting returns ErrNotEnabled", func(t *testing.T) {
		svc, _, kv := newSvc(t)
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).
			Return(mustSettingJSON(t, settingModel.EmbeddingSetting{Enable: false, Model: testModel, Dim: testDim}), nil).Once()
		res, err := svc.Search(ctx, "q", 5, "")
		require.ErrorIs(t, err, embedding.ErrNotEnabled)
		assert.Nil(t, res)
	})

	t.Run("setting backend error surfaces (not ErrNotEnabled)", func(t *testing.T) {
		svc, _, kv := newSvc(t)
		boom := errors.New("kv backend down")
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return("", boom).Once()
		res, err := svc.Search(ctx, "q", 5, "")
		require.ErrorIs(t, err, boom)
		require.NotErrorIs(t, err, embedding.ErrNotEnabled)
		assert.Nil(t, res)
	})
}
