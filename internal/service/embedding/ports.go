// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	model "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

type Embedder interface {
	Embed(ctx context.Context, setting settingModel.EmbeddingSetting, inputs []string) ([][]float32, error)
	EmbedOne(ctx context.Context, setting settingModel.EmbeddingSetting, input string) ([]float32, error)
}

type Service interface {
	IndexEcho(ctx context.Context, echo echoModel.Echo) error
	RemoveEcho(ctx context.Context, echoID string) error
	Backfill(ctx context.Context, onProgress func(BackfillResult)) (BackfillResult, error)
	Search(ctx context.Context, query string, k int, authorUsername string) ([]model.SearchResult, error)
	Enabled(ctx context.Context) bool
}

type Indexer interface {
	IndexEcho(ctx context.Context, echo echoModel.Echo) error
	RemoveEcho(ctx context.Context, echoID string) error
}

type Repository interface {
	EnsureVecTable(ctx context.Context, dim int) error
	DropVecTable(ctx context.Context) error
	Upsert(ctx context.Context, meta *model.EchoEmbedding, vector []float32) error
	Delete(ctx context.Context, echoID string) error
	GetMeta(ctx context.Context, echoID string) (*model.EchoEmbedding, bool, error)
	Search(ctx context.Context, vector []float32, k int, authorUsername string) ([]model.SearchResult, error)
	ClearAll(ctx context.Context) error
	Count(ctx context.Context) (int64, error)
}

type EchoReader interface {
	GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]echoModel.Echo, int64)
}

type BackfillResult struct {
	Total   int `json:"total"`
	Indexed int `json:"indexed"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}
