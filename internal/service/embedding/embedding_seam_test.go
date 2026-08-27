// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	embeddingmock "github.com/lin-snow/ech0/internal/test/mocks/embeddingmock"
	kvmock "github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func enabledSetting() settingModel.EmbeddingSetting {
	return settingModel.EmbeddingSetting{Enable: true, Model: testModel, Dim: testDim}
}

func newSeamSvc(t *testing.T) (
	*embeddingService.EmbeddingService,
	*embeddingmock.MockRepository,
	*kvmock.MockStore,
	*embeddingmock.MockEchoReader,
	*embeddingmock.MockEmbedder,
) {
	t.Helper()
	repo := embeddingmock.NewMockRepository(t)
	kv := kvmock.NewMockStore(t)
	reader := embeddingmock.NewMockEchoReader(t)
	emb := embeddingmock.NewMockEmbedder(t)
	svc := embeddingService.NewEmbeddingService(repo, kv, reader).WithEmbedder(emb)
	return svc, repo, kv, reader, emb
}

func expectEnsureReadyFastPath(t *testing.T, repo *embeddingmock.MockRepository, kv *kvmock.MockStore, ctx context.Context) {
	t.Helper()
	kv.EXPECT().Get(ctx, commonModel.EmbeddingIndexStateKey).
		Return(mustJSONState(t, testModel, testDim), nil).Once()
	repo.EXPECT().EnsureVecTable(ctx, testDim).Return(nil).Once()
}

func TestIndexEcho_UpsertsEmbedding(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, _, emb := newSeamSvc(t)

	echo := echoModel.Echo{
		ID:        "e1",
		Content:   "hello world",
		Username:  "alice",
		CreatedAt: 12345,
		Tags:      []echoModel.Tag{{Name: "go"}, {Name: "vue"}},
	}
	const wantText = "hello world go, vue"
	wantVec := []float32{0.1, 0.2, 0.3}

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	repo.EXPECT().GetMeta(ctx, "e1").Return(nil, false, nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	emb.EXPECT().EmbedOne(ctx, enabledSetting(), wantText).Return(wantVec, nil).Once()

	var gotMeta *embModel.EchoEmbedding
	var gotVec []float32
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).
		Run(func(_ context.Context, meta *embModel.EchoEmbedding, vector []float32) {
			gotMeta = meta
			gotVec = vector
		}).Return(nil).Once()

	require.NoError(t, svc.IndexEcho(ctx, echo))

	require.NotNil(t, gotMeta)
	assert.Equal(t, "e1", gotMeta.EchoID)
	assert.Equal(t, contentHash(t, wantText), gotMeta.ContentHash, "hash is over buildText, not raw content")
	assert.Equal(t, testModel, gotMeta.Model)
	assert.Equal(t, testDim, gotMeta.Dim)
	assert.Equal(t, "hello world", gotMeta.Content, "stored content is the raw echo content, not buildText")
	assert.Equal(t, "alice", gotMeta.Username)
	assert.Equal(t, int64(12345), gotMeta.EchoCreated)
	assert.Equal(t, wantVec, gotVec, "the embedder's vector is forwarded verbatim")
}

func TestIndexEcho_EmbedError_Propagates(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, _, emb := newSeamSvc(t)
	boom := errors.New("embed boom")

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	repo.EXPECT().GetMeta(ctx, "e1").Return(nil, false, nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	emb.EXPECT().EmbedOne(ctx, enabledSetting(), "hi").Return(nil, boom).Once()

	require.ErrorIs(t, svc.IndexEcho(ctx, echoModel.Echo{ID: "e1", Content: "hi"}), boom)
}

func TestIndexEcho_UpsertError_Propagates(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, _, emb := newSeamSvc(t)
	boom := errors.New("upsert boom")

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	repo.EXPECT().GetMeta(ctx, "e1").Return(nil, false, nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	emb.EXPECT().EmbedOne(ctx, enabledSetting(), "hi").Return([]float32{1}, nil).Once()
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).Return(boom).Once()

	require.ErrorIs(t, svc.IndexEcho(ctx, echoModel.Echo{ID: "e1", Content: "hi"}), boom)
}

func TestSearch_KNormalization(t *testing.T) {
	ctx := context.Background()
	wantVec := []float32{1, 2}

	cases := []struct {
		name  string
		inK   int
		wantK int
	}{
		{"zero falls back to defaultTopK", 0, 6},
		{"negative falls back to defaultTopK", -3, 6},
		{"positive is preserved", 4, 4},
		{"one is preserved", 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, kv, _, emb := newSeamSvc(t)
			kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
			emb.EXPECT().EmbedOne(ctx, enabledSetting(), "q").Return(wantVec, nil).Once()
			repo.EXPECT().Search(ctx, wantVec, tc.wantK, "").
				Return([]embModel.SearchResult{{EchoID: "e1"}}, nil).Once()

			got, err := svc.Search(ctx, "q", tc.inK, "")
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, "e1", got[0].EchoID)
		})
	}
}

func TestSearch_AuthorPassthrough(t *testing.T) {
	ctx := context.Background()
	wantVec := []float32{9}

	cases := []struct {
		name   string
		author string
	}{
		{"scoped to an author", "alice"},
		{"unscoped empty author", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, kv, _, emb := newSeamSvc(t)
			kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
			emb.EXPECT().EmbedOne(ctx, enabledSetting(), "q").Return(wantVec, nil).Once()
			repo.EXPECT().Search(ctx, wantVec, 5, tc.author).
				Return([]embModel.SearchResult{}, nil).Once()

			_, err := svc.Search(ctx, "q", 5, tc.author)
			require.NoError(t, err)
		})
	}
}

func TestSearch_PostSeamErrors(t *testing.T) {
	ctx := context.Background()

	t.Run("embed error propagates and skips repo.Search", func(t *testing.T) {
		svc, _, kv, _, emb := newSeamSvc(t)
		boom := errors.New("embed boom")
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
		emb.EXPECT().EmbedOne(ctx, enabledSetting(), "q").Return(nil, boom).Once()

		got, err := svc.Search(ctx, "q", 5, "")
		require.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})

	t.Run("repo.Search error propagates", func(t *testing.T) {
		svc, repo, kv, _, emb := newSeamSvc(t)
		boom := errors.New("search boom")
		vec := []float32{1}
		kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
		emb.EXPECT().EmbedOne(ctx, enabledSetting(), "q").Return(vec, nil).Once()
		repo.EXPECT().Search(ctx, vec, 5, "").Return(nil, boom).Once()

		got, err := svc.Search(ctx, "q", 5, "")
		require.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})
}

func newBackfillEcho(id, content, username string, created int64) echoModel.Echo {
	return echoModel.Echo{ID: id, Content: content, Username: username, CreatedAt: created}
}

func TestBackfill_SinglePageIndexesAll(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, emb := newSeamSvc(t)

	e1 := newBackfillEcho("e1", "a", "alice", 11)
	e2 := newBackfillEcho("e2", "b", "bob", 22)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{e1, e2}, int64(2)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"a", "b"}).
		Return([][]float32{{1, 1}, {2, 2}}, nil).Once()

	var metas []*embModel.EchoEmbedding
	var vecs [][]float32
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).
		Run(func(_ context.Context, m *embModel.EchoEmbedding, v []float32) {
			metas = append(metas, m)
			vecs = append(vecs, v)
		}).Return(nil).Times(2)

	res, err := svc.Backfill(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, embeddingService.BackfillResult{Total: 2, Indexed: 2}, res)

	require.Len(t, metas, 2)
	assert.Equal(t, "e1", metas[0].EchoID)
	assert.Equal(t, "a", metas[0].Content)
	assert.Equal(t, contentHash(t, "a"), metas[0].ContentHash)
	assert.Equal(t, "alice", metas[0].Username)
	assert.Equal(t, int64(11), metas[0].EchoCreated)
	assert.Equal(t, []float32{1, 1}, vecs[0], "vector picked by index aligns with its echo")
	assert.Equal(t, "e2", metas[1].EchoID)
	assert.Equal(t, []float32{2, 2}, vecs[1])
}

func TestBackfill_SkipsEmptyText(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, emb := newSeamSvc(t)

	e1 := newBackfillEcho("e1", "x", "u", 1)
	eEmpty := newBackfillEcho("e-empty", "   ", "u", 2)
	e2 := newBackfillEcho("e2", "y", "u", 3)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{e1, eEmpty, e2}, int64(3)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"x", "y"}).
		Return([][]float32{{1}, {2}}, nil).Once()
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).Return(nil).Times(2)

	res, err := svc.Backfill(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, embeddingService.BackfillResult{Total: 3, Indexed: 2, Skipped: 1}, res)
}

func TestBackfill_EmbedError_ReturnsLastErr(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, emb := newSeamSvc(t)
	boom := errors.New("embed http 401")

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{newBackfillEcho("e1", "a", "u", 1), newBackfillEcho("e2", "b", "u", 2)}, int64(2)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"a", "b"}).Return(nil, boom).Once()

	res, err := svc.Backfill(ctx, nil)
	require.ErrorIs(t, err, boom)
	assert.Equal(t, 0, res.Indexed)
	assert.Equal(t, 2, res.Failed)
	assert.Equal(t, 2, res.Total)
}

func TestBackfill_PartialUpsertFailure(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, emb := newSeamSvc(t)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{newBackfillEcho("e1", "a", "u", 1), newBackfillEcho("e2", "b", "u", 2)}, int64(2)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"a", "b"}).
		Return([][]float32{{1}, {2}}, nil).Once()
	repo.EXPECT().Upsert(ctx, mock.MatchedBy(func(m *embModel.EchoEmbedding) bool { return m.EchoID == "e1" }), mock.Anything).
		Return(nil).Once()
	repo.EXPECT().Upsert(ctx, mock.MatchedBy(func(m *embModel.EchoEmbedding) bool { return m.EchoID == "e2" }), mock.Anything).
		Return(errors.New("upsert boom")).Once()

	res, err := svc.Backfill(ctx, nil)
	require.NoError(t, err, "Upsert failures alone never surface an error")
	assert.Equal(t, embeddingService.BackfillResult{Total: 2, Indexed: 1, Failed: 1}, res)
}

func TestBackfill_AllUpsertFail_NoError(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, emb := newSeamSvc(t)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{newBackfillEcho("e1", "a", "u", 1), newBackfillEcho("e2", "b", "u", 2)}, int64(2)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"a", "b"}).
		Return([][]float32{{1}, {2}}, nil).Once()
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).
		Return(errors.New("upsert boom")).Times(2)

	res, err := svc.Backfill(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, embeddingService.BackfillResult{Total: 2, Indexed: 0, Failed: 2}, res)
}

func TestBackfill_MultiPage(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, emb := newSeamSvc(t)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{newBackfillEcho("e1", "a", "u", 1)}, int64(150)).Once()
	reader.EXPECT().GetEchosByPage(2, 100, "", true).
		Return([]echoModel.Echo{newBackfillEcho("e2", "b", "u", 2)}, int64(150)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"a"}).Return([][]float32{{1}}, nil).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"b"}).Return([][]float32{{2}}, nil).Once()
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).Return(nil).Times(2)

	res, err := svc.Backfill(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, 150, res.Total)
	assert.Equal(t, 2, res.Indexed)
}

func TestBackfill_EmptyFirstPage(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, _ := newSeamSvc(t)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).Return([]echoModel.Echo{}, int64(0)).Once()

	res, err := svc.Backfill(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, embeddingService.BackfillResult{}, res)
}

func TestBackfill_OnProgress(t *testing.T) {
	ctx := context.Background()
	svc, repo, kv, reader, emb := newSeamSvc(t)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{newBackfillEcho("e1", "a", "u", 1)}, int64(1)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"a"}).Return([][]float32{{1}}, nil).Once()
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).Return(nil).Once()

	var progress []embeddingService.BackfillResult
	res, err := svc.Backfill(ctx, func(r embeddingService.BackfillResult) {
		progress = append(progress, r)
	})
	require.NoError(t, err)
	require.Len(t, progress, 1)
	assert.Equal(t, embeddingService.BackfillResult{Total: 1, Indexed: 1}, progress[0])
	assert.Equal(t, progress[0], res)
}

func TestBackfill_ContextCancelledMidLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	svc, repo, kv, reader, emb := newSeamSvc(t)

	kv.EXPECT().Get(ctx, commonModel.EmbeddingSettingKey).Return(enabledSettingJSON(t), nil).Once()
	expectEnsureReadyFastPath(t, repo, kv, ctx)
	reader.EXPECT().GetEchosByPage(1, 100, "", true).
		Return([]echoModel.Echo{newBackfillEcho("e1", "a", "u", 1)}, int64(150)).Once()
	emb.EXPECT().Embed(ctx, enabledSetting(), []string{"a"}).Return([][]float32{{1}}, nil).Once()
	repo.EXPECT().Upsert(ctx, mock.Anything, mock.Anything).Return(nil).Once()

	res, err := svc.Backfill(ctx, func(embeddingService.BackfillResult) { cancel() })
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, res.Indexed, "page 1 work is preserved in the returned partial result")
}
