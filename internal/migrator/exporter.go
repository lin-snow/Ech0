// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"context"
	"strings"

	"github.com/lin-snow/ech0/internal/migrator/spec"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
)

// ExportOutcome 是一次导出的产物描述。ArtifactPath 为本地归档路径(供同步下载流式下发,不暴露给
// 前端),其余字段可序列化进作业终态 Payload 供 UI 展示。
//
// Format 与 ExportPayload.Format 同名同值,这不是巧合:作业行的 Payload 列先存输入 payload、
// 成功后被本结构覆盖,同名让 DTO 转换用一次解析覆盖「运行中」与「已完成」两种形态。
type ExportOutcome struct {
	ArtifactPath string `json:"-"`
	FileName     string `json:"file_name"`
	Size         int64  `json:"size,omitempty"`
	Format       string `json:"format,omitempty"`
}

// ExportEngine 跑导出编排:按目的地选 Exporter 适配器(配了对象存储用 s3,否则 fs)→ 运行 →
// 返回产物。与 ImportEngine 对称,不感知作业状态机(只接受裸 report 回调)。S3 上传逻辑在 s3
// 适配器内,故本编排体只需 storageManager 来判定目的地并构造适配器。
type ExportEngine struct {
	storageManager StorageManager
}

func NewExportEngine(storageManager StorageManager) *ExportEngine {
	return &ExportEngine{storageManager: storageManager}
}

// Export 选目的地适配器(fs / s3)→ 产出 Snapshot。两种目的地都会落本地产物(s3 额外上传),
// 故下载(取回最新本地产物)始终可用。
func (ex *ExportEngine) Export(
	ctx context.Context,
	report func(phase string, snapshot any),
) (ExportOutcome, error) {
	dest := migratorModel.ExportDestFS
	if ex.storageManager != nil {
		if sel := ex.storageManager.GetSelector(); sel != nil && sel.ObjectEnabled() {
			dest = migratorModel.ExportDestS3
		}
	}

	exporter, err := BuildExporter(dest, ex.storageManager)
	if err != nil {
		return ExportOutcome{}, err
	}

	result, err := exporter.Export(ctx, spec.ExportRequest{
		UpdateProgress: func(progress spec.ExportProgress) {
			if phase := strings.TrimSpace(progress.CurrentPhase); phase != "" {
				report(phase, nil)
			}
		},
	})
	if err != nil {
		return ExportOutcome{}, err
	}

	return ExportOutcome{
		ArtifactPath: result.ArtifactPath,
		FileName:     result.FileName,
		Size:         result.Size,
		Format:       migratorModel.ExportFormatSnapshot,
	}, nil
}
