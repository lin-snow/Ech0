// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	return coreSetting.Get(ctx, s.durableKV, coreSetting.Embedding)
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

func (s *EmbeddingService) Search(ctx context.Context, query string, k int, authorUsername string) ([]model.SearchResult, error) {
	setting, err := s.getSetting(ctx)
	if err != nil {
		return nil, err
	}
	if !setting.Enable || setting.Model == "" || setting.Dim <= 0 {
		return nil, embedding.ErrNotEnabled
	}
	if k <= 0 {
		k = defaultTopK
	}
	vec, err := s.embedder.EmbedOne(ctx, setting, query)
	if err != nil {
		return nil, err
	}
	return s.repo.Search(ctx, vec, k, authorUsername)
}

func (s *EmbeddingService) Backfill(ctx context.Context, onProgress func(BackfillResult)) (BackfillResult, error) {
	var result BackfillResult

	setting, err := s.getSetting(ctx)
	if err != nil {
		return result, err
	}
	if !setting.Enable || setting.Model == "" || setting.Dim <= 0 {
		return result, embedding.ErrNotEnabled
	}
	if err := s.ensureReady(ctx, setting); err != nil {
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
			vecs, embErr := s.embedder.Embed(ctx, setting, texts)
			if embErr != nil {
				logUtil.GetLogger().Error("backfill embed failed", logUtil.Err(embErr))
				result.Failed += len(texts)
				lastErr = embErr
			} else {
				for i, e := range picked {
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
			}
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
