package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupSettingRoutes 设置设置路由
func setupSettingRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/settings", h.SettingHandler.GetSettings())
	appRouterGroup.PublicRouterGroup.GET("/oauth2/status", h.SettingHandler.GetOAuth2Status())
	appRouterGroup.PublicRouterGroup.GET("/passkey/status", h.SettingHandler.GetPasskeyStatus())
	appRouterGroup.PublicRouterGroup.GET("/agent/info", h.SettingHandler.GetAgentInfo())

	// Auth
	appRouterGroup.AuthRouterGroup.PUT(
		"/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.UpdateSettings(),
	)

	appRouterGroup.AuthRouterGroup.GET(
		"/s3/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.GetS3Settings(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/s3/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.UpdateS3Settings(),
	)

	appRouterGroup.AuthRouterGroup.GET(
		"/oauth2/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.GetOAuth2Settings(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/oauth2/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.UpdateOAuth2Settings(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/passkey/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.GetPasskeySettings(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/passkey/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.UpdatePasskeySettings(),
	)

	appRouterGroup.AuthRouterGroup.GET(
		"/webhook",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.GetWebhook(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/webhook",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.CreateWebhook(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/webhook/:id",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.UpdateWebhook(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/webhook/:id",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.DeleteWebhook(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/webhook/:id/test",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.TestWebhook(),
	)

	appRouterGroup.AuthRouterGroup.GET(
		"/access-tokens",
		middleware.RequireScopes(authModel.ScopeAdminToken),
		h.SettingHandler.ListAccessTokens(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/access-tokens",
		middleware.RequireScopes(authModel.ScopeAdminToken),
		h.SettingHandler.CreateAccessToken(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/access-tokens/:id",
		middleware.RequireScopes(authModel.ScopeAdminToken),
		h.SettingHandler.DeleteAccessToken(),
	)

	appRouterGroup.AuthRouterGroup.GET(
		"/backup/schedule",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.GetBackupScheduleSetting(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/backup/schedule",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.UpdateBackupScheduleSetting(),
	)

	appRouterGroup.AuthRouterGroup.GET(
		"/agent/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.GetAgentSettings(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/agent/settings",
		middleware.RequireScopes(authModel.ScopeAdminSettings),
		h.SettingHandler.UpdateAgentSettings(),
	)
}
