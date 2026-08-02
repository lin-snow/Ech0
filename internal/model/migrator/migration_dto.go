// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// MigrationPayload 是迁移作业的领域 payload，序列化进通用 Job 的 Payload 列：既是
// 提交时的输入，也是 MigrationRunner 干活时认得的结构。
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
	Version    int    `json:"version"`
	SourceType string `json:"source_type"`
	Status     string `json:"status"`
	// Phase 是细粒度阶段（extracting/loading/...），迁入 job 子系统后的净增字段；
	// 前端忽略未知字段，故向后兼容。
	Phase         string         `json:"phase,omitempty"`
	ErrorMessage  string         `json:"error_message"`
	SourcePayload map[string]any `json:"source_payload,omitempty"`
	StartedAt     *int64         `json:"started_at,omitempty"`
	UpdatedAt     *int64         `json:"updated_at,omitempty"`
	FinishedAt    *int64         `json:"finished_at,omitempty"`
}

// ExportPayload 是导出作业的领域 payload（序列化进通用 Job 的 Payload 列）。
// Format 空值即 snapshot——既是缺省语义，也让改版前落库的作业行仍可解读。
type ExportPayload struct {
	Format string `json:"format,omitempty"`
	// IncludePrivate 仅对胶囊有意义：快照本就整库带走，无所谓包含与否。
	IncludePrivate bool `json:"include_private,omitempty"`
}

// StartExportRequest 是 POST /migration/export 的请求体。两个字段皆可省：省略即
// 「导出快照、不含私密」，与加入格式选择之前的行为完全一致。
type StartExportRequest struct {
	Format         string `json:"format,omitempty"`
	IncludePrivate bool   `json:"include_private,omitempty"`
}

// ExportStateDTO 是导出作业对前端的状态契约，与 GlobalMigrationStateDTO 对称（status 复用
// MigrationStatus* 常量，含 idle 哨兵）。FileName/Size 在终态成功时由作业 Payload 补出。
type ExportStateDTO struct {
	Version      int    `json:"version"`
	Status       string `json:"status"`
	Phase        string `json:"phase,omitempty"`
	ErrorMessage string `json:"error_message"`
	FileName     string `json:"file_name,omitempty"`
	Size         int64  `json:"size,omitempty"`
	// Format 让前端知道当前产物是哪种格式：下载链接与文件名都据此决定。缺省 snapshot。
	Format     string `json:"format,omitempty"`
	StartedAt  *int64 `json:"started_at,omitempty"`
	UpdatedAt  *int64 `json:"updated_at,omitempty"`
	FinishedAt *int64 `json:"finished_at,omitempty"`
}
