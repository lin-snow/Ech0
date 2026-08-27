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

func setupFileRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.GET(
		"/file/stream",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.StreamFileByPath,
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/file/:id/stream",
		middleware.RequireScopes(authModel.ScopeFileRead),
		h.FileHandler.StreamFileByID,
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/files/upload",
		middleware.RequireScopes(authModel.ScopeFileWrite),
		h.FileHandler.UploadFile(),
	)
}

func registerFile(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, secured(revoker, authModel.ScopeFileRead), huma.Operation{
		OperationID: "file-list",
		Method:      http.MethodGet,
		Path:        "/files",
		Summary:     "分页获取文件列表",
		Tags:        []string{"File"},
	}, h.FileHandler.ListFiles)

	route(api, secured(revoker, authModel.ScopeFileRead), huma.Operation{
		OperationID: "file-tree",
		Method:      http.MethodGet,
		Path:        "/file/tree",
		Summary:     "获取文件树",
		Tags:        []string{"File"},
	}, h.FileHandler.ListFileTree)

	route(api, secured(revoker, authModel.ScopeFileRead), huma.Operation{
		OperationID: "file-get",
		Method:      http.MethodGet,
		Path:        "/file/{id}",
		Summary:     "获取文件元信息",
		Tags:        []string{"File"},
	}, h.FileHandler.GetFileByID)

	route(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-update-meta",
		Method:      http.MethodPut,
		Path:        "/file/{id}/meta",
		Summary:     "更新对象存储文件元信息",
		Tags:        []string{"File"},
	}, h.FileHandler.UpdateFileMeta)

	route(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-external",
		Method:      http.MethodPost,
		Path:        "/files/external",
		Summary:     "登记外链文件",
		Tags:        []string{"File"},
	}, h.FileHandler.CreateExternalFile)

	route(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-delete",
		Method:      http.MethodDelete,
		Path:        "/file/{id}",
		Summary:     "删除文件",
		Tags:        []string{"File"},
	}, h.FileHandler.DeleteFile)

	route(api, secured(revoker, authModel.ScopeFileWrite), huma.Operation{
		OperationID: "file-presign",
		Method:      http.MethodPut,
		Path:        "/files/presign",
		Summary:     "获取对象存储直传预签名 URL",
		Tags:        []string{"File"},
	}, h.FileHandler.GetFilePresignURL)
}
