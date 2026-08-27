// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	response "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	service "github.com/lin-snow/ech0/internal/service/migrator"
)

type MigrationHandler struct {
	migrationService service.Service
}

func NewMigrationHandler(migrationService service.Service) *MigrationHandler {
	return &MigrationHandler{
		migrationService: migrationService,
	}
}

type (
	StartMigrationInput struct {
		Body migratorModel.StartGlobalMigrationRequest
	}
	GetMigrationStatusInput struct{}
	CancelMigrationInput    struct{}
	CleanupMigrationInput   struct{}
	StartExportInput        struct {
		Body migratorModel.StartExportRequest
	}
	GetExportStatusInput struct{}
	CancelExportInput    struct{}
)

type (
	GlobalMigrationOutput = commonModel.Result[migratorModel.GlobalMigrationStateDTO]
	ExportOutput          = commonModel.Result[migratorModel.ExportStateDTO]
	EmptyOutput           = commonModel.Result[any]
)

func (h *MigrationHandler) StartMigration(ctx context.Context, in *StartMigrationInput) (GlobalMigrationOutput, error) {
	data, err := h.migrationService.StartGlobalMigration(ctx, in.Body)
	if err != nil {
		return GlobalMigrationOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *MigrationHandler) GetMigrationStatus(ctx context.Context, _ *GetMigrationStatusInput) (GlobalMigrationOutput, error) {
	data, err := h.migrationService.GetGlobalMigrationStatus(ctx)
	if err != nil {
		return GlobalMigrationOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *MigrationHandler) CancelMigration(ctx context.Context, _ *CancelMigrationInput) (GlobalMigrationOutput, error) {
	data, err := h.migrationService.CancelGlobalMigration(ctx)
	if err != nil {
		return GlobalMigrationOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *MigrationHandler) CleanupMigration(ctx context.Context, _ *CleanupMigrationInput) (EmptyOutput, error) {
	if err := h.migrationService.CleanupGlobalMigration(ctx); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil), nil
}

func (h *MigrationHandler) StartExport(ctx context.Context, in *StartExportInput) (ExportOutput, error) {
	data, err := h.migrationService.StartExport(ctx, in.Body)
	if err != nil {
		return ExportOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *MigrationHandler) GetExportStatus(ctx context.Context, _ *GetExportStatusInput) (ExportOutput, error) {
	data, err := h.migrationService.GetExportStatus(ctx)
	if err != nil {
		return ExportOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *MigrationHandler) CancelExport(ctx context.Context, _ *CancelExportInput) (ExportOutput, error) {
	data, err := h.migrationService.CancelExport(ctx)
	if err != nil {
		return ExportOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *MigrationHandler) UploadSourceZip() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		sourceType := ctx.PostForm("source_type")
		if sourceType == "" {
			return response.Response{Msg: commonModel.INVALID_REQUEST_BODY}
		}
		file, err := ctx.FormFile("file")
		if err != nil {
			return response.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		data, err := h.migrationService.UploadSourceZip(ctx.Request.Context(), sourceType, file)
		if err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.SUCCESS_MESSAGE, Data: data}
	})
}

func (h *MigrationHandler) DownloadExport() gin.HandlerFunc {
	return response.Execute(func(ctx *gin.Context) response.Response {
		if err := h.migrationService.DownloadExport(ctx, ctx.Request.Context(), ctx.Query("format")); err != nil {
			return response.Response{Msg: "", Err: err}
		}
		return response.Response{Msg: commonModel.EXPORT_SNAPSHOT_SUCCESS}
	})
}
