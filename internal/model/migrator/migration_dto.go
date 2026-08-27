// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type MigrationPayload struct {
	SourceType    string         `json:"source_type"`
	SourcePayload map[string]any `json:"source_payload"`
}

type StartGlobalMigrationRequest struct {
	SourceType    string         `json:"source_type" binding:"required"`
	SourcePayload map[string]any `json:"source_payload"`
}

type UploadMigrationSourceZipResponse struct {
	SourceType    string         `json:"source_type"`
	TmpDir        string         `json:"tmp_dir"`
	SourcePayload map[string]any `json:"source_payload"`
}

type GlobalMigrationStateDTO struct {
	Version       int            `json:"version"`
	SourceType    string         `json:"source_type"`
	Status        string         `json:"status"`
	Phase         string         `json:"phase,omitempty"`
	ErrorMessage  string         `json:"error_message"`
	SourcePayload map[string]any `json:"source_payload,omitempty"`
	StartedAt     *int64         `json:"started_at,omitempty"`
	UpdatedAt     *int64         `json:"updated_at,omitempty"`
	FinishedAt    *int64         `json:"finished_at,omitempty"`
}

type ExportPayload struct {
	Format         string `json:"format,omitempty"`
	IncludePrivate bool   `json:"include_private,omitempty"`
}

type StartExportRequest struct {
	Format         string `json:"format,omitempty"`
	IncludePrivate bool   `json:"include_private,omitempty"`
}

type ExportStateDTO struct {
	Version      int    `json:"version"`
	Status       string `json:"status"`
	Phase        string `json:"phase,omitempty"`
	ErrorMessage string `json:"error_message"`
	FileName     string `json:"file_name,omitempty"`
	Size         int64  `json:"size,omitempty"`
	Format       string `json:"format,omitempty"`
	StartedAt    *int64 `json:"started_at,omitempty"`
	UpdatedAt    *int64 `json:"updated_at,omitempty"`
	FinishedAt   *int64 `json:"finished_at,omitempty"`
}
