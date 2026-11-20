// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/lin-snow/ech0/internal/embedding"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	model "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

const defaultTopK = 6

type EmbeddingService struct {
	repo       Repository
	durableKV  kvstore.Store
	echoReader EchoReader
	embedder   Embedder
}

var (
	_ Service = (*EmbeddingService)(nil)
	_ Indexer = (*EmbeddingService)(nil)
)

func NewEmbeddingService(repo Repository, durableKV kvstore.Store, echoReader EchoReader) *EmbeddingService {
	return &EmbeddingService{repo: repo, durableKV: durableKV, echoReader: echoReader, embedder: embedding.Client{}}
}

func (s *EmbeddingService) WithEmbedder(e Embedder) *EmbeddingService {
	s.embedder = e
	return s
}

func (s *EmbeddingService) getSetting(ctx context.Context) (settingModel.EmbeddingSetting, error) {
	return coreSetting.Embedding.Get(ctx, s.durableKV)
}

func (s *EmbeddingService) Enabled(ctx context.Context) bool {
	setting, err := s.getSetting(ctx)
	if err != nil {
		return false
	}
	return setting.Enable && setting.Model != "" && setting.Dim > 0
}

func (s *EmbeddingService) ensureReady(ctx context.Context, setting settingModel.EmbeddingSetting) error {
	var state model.IndexState
	if stateRaw, err := s.durableKV.Get(ctx, commonModel.EmbeddingIndexStateKey); err == nil {
		_ = json.Unmarshal([]byte(stateRaw), &state)
	}

	if state.Dim == setting.Dim && state.Model == setting.Model {
		return s.repo.EnsureVecTable(ctx, setting.Dim)
	}

	if err := s.repo.DropVecTable(ctx); err != nil {
		return err
	}
	if err := s.repo.ClearAll(ctx); err != nil {
		return err
	}
	if err := s.repo.EnsureVecTable(ctx, setting.Dim); err != nil {
		return err
	}
	newState, _ := json.Marshal(model.IndexState{Model: setting.Model, Dim: setting.Dim})
	return s.durableKV.Set(ctx, commonModel.EmbeddingIndexStateKey, string(newState))
}

func buildText(echo echoModel.Echo) string {
	var b strings.Builder
	b.WriteString(echo.Content)
	if len(echo.Tags) > 0 {
		b.WriteString(" ")
		for i, t := range echo.Tags {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t.Name)
		}
	}
	return strings.TrimSpace(b.String())
}

func hashContent(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func (s *EmbeddingService) IndexEcho(ctx context.Context, echo echoModel.Echo) error {
	setting, err := s.getSetting(ctx)
	if err != nil {
		return err
	}
	if !setting.Enable || setting.Model == "" || setting.Dim <= 0 {
		return nil
	}

	text := buildText(echo)
	if text == "" {
		return s.repo.Delete(ctx, echo.ID)
	}

	hash := hashContent(text)
	if meta, ok, _ := s.repo.GetMeta(ctx, echo.ID); ok &&
		meta.ContentHash == hash && meta.Model == setting.Model && meta.Dim == setting.Dim {
		return nil
	}

	if err := s.ensureReady(ctx, setting); err != nil {
		return err
	}

	vec, err := s.embedder.EmbedOne(ctx, setting, text)
	if err != nil {
		return err
	}

	return s.repo.Upsert(ctx, &model.EchoEmbedding{
		EchoID:      echo.ID,
		ContentHash: hash,
		Model:       setting.Model,
		Dim:         setting.Dim,
		Content:     echo.Content,
		Username:    echo.Username,
		EchoCreated: echo.CreatedAt,
	}, vec)
}

func (s *EmbeddingService) RemoveEcho(ctx context.Context, echoID string) error {
	return s.repo.Delete(ctx, echoID)
}

// Search returns the Echos semantically nearest to query, optionally only those
// by authorID. It refuses with embedding.ErrIndexNotReady when no index was
// built for the configured model and dimension — before the first rebuild, or
// after either was changed — instead of querying a missing table or comparing
// vectors of the wrong size; callers fall back to keyword search on any error.
func (s *EmbeddingService) Search(ctx context.Context, query string, k int, authorID string) ([]model.SearchResult, error) {
	setting, err := s.getSetting(ctx)
	if err != nil {
		return nil, err
	}
	if !setting.Enable || setting.Model == "" || setting.Dim <= 0 {
		return nil, embedding.ErrNotEnabled
	}
	if !s.indexMatches(ctx, setting) {
		return nil, embedding.ErrIndexNotReady
	}
	if k <= 0 {
		k = defaultTopK
	}
	vec, err := s.embedder.EmbedOne(ctx, setting, query)
	if err != nil {
		return nil, err
	}
	return s.repo.Search(ctx, vec, k, authorID)
}

func (s *EmbeddingService) indexMatches(ctx context.Context, setting settingModel.EmbeddingSetting) bool {
	raw, err := s.durableKV.Get(ctx, commonModel.EmbeddingIndexStateKey)
	if err != nil {
		return false
	}
	var state model.IndexState
	if json.Unmarshal([]byte(raw), &state) != nil {
		return false
	}
	return state.Model == setting.Model && state.Dim == setting.Dim
}

func (s *EmbeddingService) Backfill(ctx context.Context, onProgress func(BackfillResult)) (BackfillResult, error) {
	var result BackfillResult

	setting, err := s.PrepareBackfill(ctx)
	if err != nil {
		return result, err
	}

	const pageSize = 100
	page := 1
	var lastErr error
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}

		items, total := s.echoReader.GetEchosByPage(page, pageSize, "", true)
		result.Total = int(total)
		if len(items) == 0 {
			break
		}

		batch, err := s.IndexBatch(ctx, setting, items)
		result.Indexed += batch.Indexed
		result.Failed += batch.Failed
		result.Skipped += batch.Skipped
		if err != nil {
			lastErr = err
		}

		if onProgress != nil {
			onProgress(result)
		}

		if page*pageSize >= int(total) {
			break
		}
		page++
	}

	if result.Indexed == 0 && result.Failed > 0 && lastErr != nil {
		return result, lastErr
	}

	return result, nil
}

// PrepareBackfill initializes the configured native or remote index and prunes
// orphan rows for both the local runner and the checkpointed Function worker.
func (s *EmbeddingService) PrepareBackfill(ctx context.Context) (settingModel.EmbeddingSetting, error) {
	setting, err := s.getSetting(ctx)
	if err != nil {
		return setting, err
	}
	if !setting.Enable || setting.Model == "" || setting.Dim <= 0 {
		return setting, embedding.ErrNotEnabled
	}
	if err := s.ensureReady(ctx, setting); err != nil {
		return setting, err
	}
	if pruned, err := s.repo.PruneOrphans(ctx); err != nil {
		logUtil.GetLogger().Warn("prune orphaned embeddings failed",
			slog.String("module", "embedding"), logUtil.Err(err))
	} else if pruned > 0 {
		logUtil.GetLogger().Info("pruned orphaned embeddings",
			slog.String("module", "embedding"), slog.Int64("count", pruned))
	}

	return setting, nil
}

func (s *EmbeddingService) IndexBatch(ctx context.Context, setting settingModel.EmbeddingSetting, items []echoModel.Echo) (BackfillResult, error) {
	result := BackfillResult{Total: len(items)}
	texts := make([]string, 0, len(items))
	picked := make([]echoModel.Echo, 0, len(items))
	for _, e := range items {
		t := buildText(e)
		if t == "" {
			result.Skipped++
			continue
		}
		texts = append(texts, t)
		picked = append(picked, e)
	}

	if len(texts) > 0 {
		vecs, embErr := s.embedBatch(ctx, setting, texts)
		for i, e := range picked {
			if vecs[i] == nil {
				result.Failed++
				continue
			}
			if upErr := s.repo.Upsert(ctx, &model.EchoEmbedding{
				EchoID:      e.ID,
				ContentHash: hashContent(texts[i]),
				Model:       setting.Model,
				Dim:         setting.Dim,
				Content:     e.Content,
				Username:    e.Username,
				EchoCreated: e.CreatedAt,
			}, vecs[i]); upErr != nil {
				result.Failed++
			} else {
				result.Indexed++
			}
		}
		if embErr != nil {
			return result, embErr
		}
	}

	return result, nil
}

// embedBatch embeds texts, returning one vector per input and nil where an
// input could not be embedded. A batch the provider rejects is retried one
// input at a time, so a single bad Echo — too long for the model, or tripping
// a content filter — costs itself rather than the whole page every rebuild.
func (s *EmbeddingService) embedBatch(ctx context.Context, setting settingModel.EmbeddingSetting, texts []string) ([][]float32, error) {
	vecs, err := s.embedder.Embed(ctx, setting, texts)
	if err == nil && len(vecs) == len(texts) {
		return vecs, nil
	}
	logUtil.GetLogger().Warn("backfill batch embed failed, retrying one by one",
		slog.String("module", "embedding"), slog.Int("size", len(texts)), logUtil.Err(err))

	vecs = make([][]float32, len(texts))
	var lastErr error
	for i, t := range texts {
		if ctx.Err() != nil {
			return vecs, ctx.Err()
		}
		v, oneErr := s.embedder.EmbedOne(ctx, setting, t)
		if oneErr != nil {
			lastErr = oneErr
			continue
		}
		vecs[i] = v
	}
	return vecs, lastErr
}
