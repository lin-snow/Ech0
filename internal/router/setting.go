package router

import "github.com/lin-snow/ech0/internal/handler"

// setupSettingRoutes 设置设置路由
func setupSettingRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/settings", h.SettingHandler.GetSettings())
	appRouterGroup.PublicRouterGroup.GET("/comment/settings", h.SettingHandler.GetCommentSettings())
	appRouterGroup.PublicRouterGroup.GET("/oauth2/status", h.SettingHandler.GetOAuth2Status())
	appRouterGroup.PublicRouterGroup.GET("/agent/info", h.SettingHandler.GetAgentInfo())

	// Auth
	appRouterGroup.AuthRouterGroup.PUT("/settings", h.SettingHandler.UpdateSettings())

	appRouterGroup.AuthRouterGroup.PUT(
		"/comment/settings",
		h.SettingHandler.UpdateCommentSettings(),
	)

	appRouterGroup.AuthRouterGroup.GET("/s3/settings", h.SettingHandler.GetS3Settings())
	appRouterGroup.AuthRouterGroup.PUT("/s3/settings", h.SettingHandler.UpdateS3Settings())

	appRouterGroup.AuthRouterGroup.GET("/oauth2/settings", h.SettingHandler.GetOAuth2Settings())
	appRouterGroup.AuthRouterGroup.PUT("/oauth2/settings", h.SettingHandler.UpdateOAuth2Settings())

	appRouterGroup.AuthRouterGroup.GET("/webhook", h.SettingHandler.GetWebhook())
	appRouterGroup.AuthRouterGroup.POST("/webhook", h.SettingHandler.CreateWebhook())
	appRouterGroup.AuthRouterGroup.PUT("/webhook", h.SettingHandler.UpdateWebhook())
	appRouterGroup.AuthRouterGroup.DELETE("/webhook/:id", h.SettingHandler.DeleteWebhook())

	appRouterGroup.AuthRouterGroup.GET("/access-tokens", h.SettingHandler.ListAccessTokens())
	appRouterGroup.AuthRouterGroup.POST("/access-tokens", h.SettingHandler.CreateAccessToken())
	appRouterGroup.AuthRouterGroup.DELETE(
		"/access-tokens/:id",
		h.SettingHandler.DeleteAccessToken(),
	)

	appRouterGroup.AuthRouterGroup.GET(
		"/backup/schedule",
		h.SettingHandler.GetBackupScheduleSetting(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/backup/schedule",
		h.SettingHandler.UpdateBackupScheduleSetting(),
	)

	appRouterGroup.AuthRouterGroup.GET("/agent/settings", h.SettingHandler.GetAgentSettings())
	appRouterGroup.AuthRouterGroup.PUT("/agent/settings", h.SettingHandler.UpdateAgentSettings())
}
