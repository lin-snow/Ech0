// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	UploadSourceZip(ctx context.Context, sourceType string, file *multipart.FileHeader) (migratorModel.UploadMigrationSourceZipResponse, error)
	StartGlobalMigration(ctx context.Context, req migratorModel.StartGlobalMigrationRequest) (migratorModel.GlobalMigrationStateDTO, error)
	GetGlobalMigrationStatus(ctx context.Context) (migratorModel.GlobalMigrationStateDTO, error)
	CancelGlobalMigration(ctx context.Context) (migratorModel.GlobalMigrationStateDTO, error)
	CleanupGlobalMigration(ctx context.Context) error

	StartExport(ctx context.Context, req migratorModel.StartExportRequest) (migratorModel.ExportStateDTO, error)
	GetExportStatus(ctx context.Context) (migratorModel.ExportStateDTO, error)
	CancelExport(ctx context.Context) (migratorModel.ExportStateDTO, error)
	DownloadExport(ctx *gin.Context, reqCtx context.Context, format string) error
}

type CommonService = commonService.Service
