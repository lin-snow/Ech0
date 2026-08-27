// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package runner

import (
	"context"

	"github.com/lin-snow/ech0/internal/job"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
)

type ReindexPayload struct{}

type ReindexRunner struct {
	svc embeddingService.Service
}

func NewReindexRunner(svc embeddingService.Service) *ReindexRunner {
	return &ReindexRunner{svc: svc}
}

func (r *ReindexRunner) Run(ctx context.Context, _ ReindexPayload, report job.ReportFunc) (any, error) {
	res, err := r.svc.Backfill(ctx, func(progress embeddingService.BackfillResult) {
		report("indexing", progress)
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
