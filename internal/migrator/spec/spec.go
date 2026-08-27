// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package spec

import "context"

type ImportRequest struct {
	SourcePayload  map[string]any
	UpdateProgress func(progress ImportProgress)
}

type ImportProgress struct {
	CurrentPhase string
	Processed    int64
	Total        int64
	SuccessCount int64
	FailCount    int64
	ErrorSummary string
}

type FailedItem struct {
	SourceID string `json:"source_id"`
	Reason   string `json:"reason"`
}

type ImportResult struct {
	Processed    int64
	Total        int64
	SuccessCount int64
	FailCount    int64
	ErrorSummary string
	JobID        string
	Report       map[string]any
}

type Importer interface {
	Import(ctx context.Context, req ImportRequest) (ImportResult, error)
}

type ExportRequest struct {
	UpdateProgress func(progress ExportProgress)
}

type ExportProgress struct {
	CurrentPhase string
}

type ExportResult struct {
	ArtifactPath string
	FileName     string
	Size         int64
}

type Exporter interface {
	Export(ctx context.Context, req ExportRequest) (ExportResult, error)
}
