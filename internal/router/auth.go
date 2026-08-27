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

func setupAuthRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/login", middleware.NoCache(), h.AuthHandler.OAuthLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/callback", middleware.NoCache(), h.AuthHandler.OAuthCallback())

	appRouterGroup.PublicRouterGroup.POST("/login", middleware.NoCache(), h.AuthHandler.Login())
	appRouterGroup.PublicRouterGroup.POST("/passkey/login/begin", middleware.NoCache(), h.AuthHandler.PasskeyLoginBeginV2())
	appRouterGroup.PublicRouterGroup.POST("/passkey/login/finish", middleware.NoCache(), h.AuthHandler.PasskeyLoginFinishV2())
	appRouterGroup.PublicRouterGroup.POST("/auth/refresh", middleware.NoCache(), h.AuthHandler.Refresh())
	appRouterGroup.PublicRouterGroup.POST("/auth/logout", middleware.NoCache(), h.AuthHandler.Logout())
	appRouterGroup.PublicRouterGroup.POST("/auth/exchange", middleware.NoCache(), h.AuthHandler.Exchange())

	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/begin",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.PasskeyRegisterBeginV2(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/finish",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.PasskeyRegisterFinishV2(),
	)
}

func registerAuth(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, secured(revoker, authModel.ScopeProfileWrite), huma.Operation{
		OperationID: "oauth-bind",
		Method:      http.MethodPost,
		Path:        "/oauth/{provider}/bind",
		Summary:     "绑定 OAuth2 账号到当前用户",
		Tags:        []string{"Auth"},
	}, h.AuthHandler.OAuthBind)

	route(api, secured(revoker, authModel.ScopeProfileRead), huma.Operation{
		OperationID: "oauth-info",
		Method:      http.MethodGet,
		Path:        "/oauth/info",
		Summary:     "获取当前用户的 OAuth2 绑定信息",
		Tags:        []string{"Auth"},
	}, h.AuthHandler.GetOAuthInfo)

	route(api, secured(revoker, authModel.ScopeProfileRead), huma.Operation{
		OperationID: "passkey-list",
		Method:      http.MethodGet,
		Path:        "/passkeys",
		Summary:     "列出当前用户的 Passkey 设备",
		Tags:        []string{"Auth"},
	}, h.AuthHandler.ListPasskeys)

	route(api, secured(revoker, authModel.ScopeProfileWrite), huma.Operation{
		OperationID: "passkey-delete",
		Method:      http.MethodDelete,
		Path:        "/passkeys/{id}",
		Summary:     "删除 Passkey 设备",
		Tags:        []string{"Auth"},
	}, h.AuthHandler.DeletePasskey)

	route(api, secured(revoker, authModel.ScopeProfileWrite), huma.Operation{
		OperationID: "passkey-update-name",
		Method:      http.MethodPut,
		Path:        "/passkeys/{id}",
		Summary:     "更新 Passkey 设备名称",
		Tags:        []string{"Auth"},
	}, h.AuthHandler.UpdatePasskeyDeviceName)
}
