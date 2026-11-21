// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/ratelimit"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func registerUser(api *apiRegistry, h *handler.Bundle, revoker authService.TokenRevoker) {
	api.route(public(), huma.Operation{
		OperationID: "user-register",
		Method:      http.MethodPost,
		Path:        "/register",
		Summary:     "用户注册",
		Tags:        []string{"User"},
		Middlewares: append(noCache(), api.scopeLimit(ratelimit.ScopeSignUpIP, 1)...),
	}, h.UserHandler.Register)

	api.route(secured(revoker, authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-list",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "获取所有用户",
		Tags:        []string{"User"},
	}, h.UserHandler.GetAllUsers)

	api.route(secured(revoker, authz.ScopeProfileRead), huma.Operation{
		OperationID: "user-info",
		Method:      http.MethodGet,
		Path:        "/user",
		Summary:     "获取当前用户信息",
		Tags:        []string{"User"},
	}, h.UserHandler.GetUserInfo)

	api.route(secured(revoker, authz.ScopeProfileWrite), huma.Operation{
		OperationID: "user-update",
		Method:      http.MethodPut,
		Path:        "/user",
		Summary:     "更新当前用户信息",
		Tags:        []string{"User"},
	}, h.UserHandler.UpdateUser)

	api.route(secured(revoker, authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-delete",
		Method:      http.MethodDelete,
		Path:        "/user/{id}",
		Summary:     "删除用户",
		Tags:        []string{"User"},
	}, h.UserHandler.DeleteUser)

	api.route(secured(revoker, authz.ScopeAdminUser), huma.Operation{
		OperationID: "user-set-admin",
		Method:      http.MethodPut,
		Path:        "/user/admin/{id}",
		Summary:     "切换用户管理员权限",
		Tags:        []string{"User"},
	}, h.UserHandler.UpdateUserAdmin)
}
