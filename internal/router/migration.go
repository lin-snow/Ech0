// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func setupMigrationRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST(
		"/migration/upload",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.UploadSourceZip(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/migration/export/download",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.MigrationHandler.DownloadExport(),
	)
}

func registerMigration(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	admin := secured(revoker, authModel.ScopeAdminSettings)

	route(api, admin, huma.Operation{
		OperationID: "migration-start",
		Method:      http.MethodPost,
		Path:        "/migration/start",
		Summary:     "启动全局数据迁移",
		Tags:        []string{"Migration"},
	}, h.MigrationHandler.StartMigration)

	route(api, admin, huma.Operation{
		OperationID: "migration-status",
		Method:      http.MethodGet,
		Path:        "/migration/status",
		Summary:     "查询全局迁移状态",
		Tags:        []string{"Migration"},
	}, h.MigrationHandler.GetMigrationStatus)

	route(api, admin, huma.Operation{
		OperationID: "migration-cancel",
		Method:      http.MethodPost,
		Path:        "/migration/cancel",
		Summary:     "取消进行中的全局迁移",
		Tags:        []string{"Migration"},
	}, h.MigrationHandler.CancelMigration)

	route(api, admin, huma.Operation{
		OperationID: "migration-cleanup",
		Method:      http.MethodPost,
		Path:        "/migration/cleanup",
		Summary:     "清理迁移中间产物",
		Tags:        []string{"Migration"},
	}, h.MigrationHandler.CleanupMigration)

	route(api, admin, huma.Operation{
		OperationID: "migration-export",
		Method:      http.MethodPost,
		Path:        "/migration/export",
		Summary:     "提交导出作业",
		Tags:        []string{"Migration"},
	}, h.MigrationHandler.StartExport)

	route(api, admin, huma.Operation{
		OperationID: "migration-export-status",
		Method:      http.MethodGet,
		Path:        "/migration/export/status",
		Summary:     "查询导出作业状态",
		Tags:        []string{"Migration"},
	}, h.MigrationHandler.GetExportStatus)

	route(api, admin, huma.Operation{
		OperationID: "migration-export-cancel",
		Method:      http.MethodPost,
		Path:        "/migration/export/cancel",
		Summary:     "取消导出作业",
		Tags:        []string{"Migration"},
	}, h.MigrationHandler.CancelExport)
}
