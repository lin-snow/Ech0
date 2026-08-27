// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package memos

import (
	"context"

	"github.com/lin-snow/ech0/internal/migrator/spec"
)

type Importer struct{}

func New() *Importer {
	return &Importer{}
}

func (e *Importer) Import(_ context.Context, req spec.ImportRequest) (spec.ImportResult, error) {
	if req.UpdateProgress != nil {
		req.UpdateProgress(spec.ImportProgress{CurrentPhase: "loading"})
	}
	return spec.ImportResult{}, nil
}
