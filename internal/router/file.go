// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

func setupFileRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Bytes and multipart stay on raw gin: one upload endpoint, one content
	// gateway, and the gateway answers GET and HEAD identically.
	appRouterGroup.AuthRouterGroup.POST(
		"/files",
		middleware.RequireScopes(authz.ScopeFileWrite),
		h.FileHandler.Upload(),
	)
	appRouterGroup.OptionalAuthRouterGroup.GET("/files/:id/content", h.FileHandler.Content)
	appRouterGroup.OptionalAuthRouterGroup.HEAD("/files/:id/content", h.FileHandler.Content)
}

func registerFile(api *apiRegistry, h *handler.Bundle) {
	api.route(scoped(authz.ScopeFileWrite), huma.Operation{
		OperationID:   "file-commit",
		Method:        http.MethodPost,
		Path:          "/files/{id}/commit",
		Summary:       "确认直传完成",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"File"},
	}, h.FileHandler.Commit)

	api.route(optional(), huma.Operation{
		OperationID: "file-get",
		Method:      http.MethodGet,
		Path:        "/files/{id}",
		Summary:     "获取文件元信息",
		Tags:        []string{"File"},
	}, h.FileHandler.Get)

	api.route(scoped(authz.ScopeFileRead), huma.Operation{
		OperationID: "file-list",
		Method:      http.MethodGet,
		Path:        "/files",
		Summary:     "分页获取文件列表",
		Tags:        []string{"File"},
	}, h.FileHandler.List)

	api.route(scoped(authz.ScopeFileWrite), huma.Operation{
		OperationID: "file-rename",
		Method:      http.MethodPatch,
		Path:        "/files/{id}",
		Summary:     "修改文件展示名称",
		Tags:        []string{"File"},
	}, h.FileHandler.Rename)

	api.route(scoped(authz.ScopeFileDelete), huma.Operation{
		OperationID: "file-delete",
		Method:      http.MethodDelete,
		Path:        "/files/{id}",
		Summary:     "删除文件",
		Tags:        []string{"File"},
	}, h.FileHandler.Delete)
}
