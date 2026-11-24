// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/ratelimit"
)

func registerUser(api *apiRegistry, h *handler.Bundle) {
	api.route(public("registration is open until the instance reaches its member limit"), huma.Operation{
		OperationID:   "user-register",
		Method:        http.MethodPost,
		Path:          "/register",
		Summary:       "用户注册",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"User"},
		Middlewares:   append(noCache(), api.scopeLimit(ratelimit.ScopeSignUpIP, 1)...),
	}, h.UserHandler.Register)

	api.route(scoped(authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-list",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "获取所有用户",
		Tags:        []string{"User"},
	}, h.UserHandler.GetAllUsers)

	api.route(scoped(authz.ScopeProfileRead), huma.Operation{
		OperationID: "user-info",
		Method:      http.MethodGet,
		Path:        "/user",
		Summary:     "获取当前用户信息",
		Tags:        []string{"User"},
	}, h.UserHandler.GetUserInfo)

	api.route(scoped(authz.ScopeProfileWrite), huma.Operation{
		OperationID: "user-update",
		Method:      http.MethodPut,
		Path:        "/user",
		Summary:     "更新当前用户信息",
		Tags:        []string{"User"},
	}, h.UserHandler.UpdateUser)

	api.route(scoped(authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-delete",
		Method:      http.MethodDelete,
		Path:        "/user/{id}",
		Summary:     "删除用户",
		Tags:        []string{"User"},
	}, h.UserHandler.DeleteUser)

	api.route(scoped(authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-set-status",
		Method:      http.MethodPut,
		Path:        "/user/status/{id}",
		Summary:     "启用或禁用用户",
		Tags:        []string{"User"},
	}, h.UserHandler.UpdateUserStatus)

	api.route(scoped(authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-set-role",
		Method:      http.MethodPut,
		Path:        "/user/role/{id}",
		Summary:     "设置用户角色（admin 或 member）",
		Tags:        []string{"User"},
	}, h.UserHandler.UpdateUserRole)

	api.route(scoped(authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-transfer-owner",
		Method:      http.MethodPut,
		Path:        "/user/owner/{id}",
		Summary:     "转让空间所有权",
		Tags:        []string{"User"},
	}, h.UserHandler.TransferOwnership)
}
