// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func registerUser(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, public(), huma.Operation{
		OperationID: "user-register",
		Method:      http.MethodPost,
		Path:        "/register",
		Summary:     "用户注册",
		Tags:        []string{"User"},
		Middlewares: noCache(),
	}, h.UserHandler.Register)

	route(api, secured(revoker, authModel.ScopeAdminUser), huma.Operation{
		OperationID: "user-list",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "获取所有用户",
		Tags:        []string{"User"},
	}, h.UserHandler.GetAllUsers)

	route(api, secured(revoker, authModel.ScopeProfileRead), huma.Operation{
		OperationID: "user-info",
		Method:      http.MethodGet,
		Path:        "/user",
		Summary:     "获取当前用户信息",
		Tags:        []string{"User"},
	}, h.UserHandler.GetUserInfo)

	route(api, secured(revoker, authModel.ScopeProfileWrite), huma.Operation{
		OperationID: "user-update",
		Method:      http.MethodPut,
		Path:        "/user",
		Summary:     "更新当前用户信息",
		Tags:        []string{"User"},
	}, h.UserHandler.UpdateUser)

	route(api, secured(revoker, authModel.ScopeAdminUser), huma.Operation{
		OperationID: "user-delete",
		Method:      http.MethodDelete,
		Path:        "/user/{id}",
		Summary:     "删除用户",
		Tags:        []string{"User"},
	}, h.UserHandler.DeleteUser)

	route(api, secured(revoker, authModel.ScopeAdminUser), huma.Operation{
		OperationID: "user-set-admin",
		Method:      http.MethodPut,
		Path:        "/user/admin/{id}",
		Summary:     "切换用户管理员权限",
		Tags:        []string{"User"},
	}, h.UserHandler.UpdateUserAdmin)
}
