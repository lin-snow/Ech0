package router

import "github.com/lin-snow/ech0/internal/handler"

// setupSettingRoutes 设置设置路由
func setupSettingRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/settings", h.SettingHandler.GetSettings())
	appRouterGroup.PublicRouterGroup.GET("/oauth2/status", h.SettingHandler.GetOAuth2Status())
	appRouterGroup.PublicRouterGroup.GET("/passkey/status", h.SettingHandler.GetPasskeyStatus())
	appRouterGroup.PublicRouterGroup.GET("/agent/info", h.SettingHandler.GetAgentInfo())

	// Auth
	appRouterGroup.FullAuthRouterGroup.PUT("/settings", h.SettingHandler.UpdateSettings())

	appRouterGroup.FullAuthRouterGroup.GET("/s3/settings", h.SettingHandler.GetS3Settings())
	appRouterGroup.FullAuthRouterGroup.PUT("/s3/settings", h.SettingHandler.UpdateS3Settings())

	appRouterGroup.FullAuthRouterGroup.GET("/oauth2/settings", h.SettingHandler.GetOAuth2Settings())
	appRouterGroup.FullAuthRouterGroup.PUT("/oauth2/settings", h.SettingHandler.UpdateOAuth2Settings())
	appRouterGroup.FullAuthRouterGroup.GET("/passkey/settings", h.SettingHandler.GetPasskeySettings())
	appRouterGroup.FullAuthRouterGroup.PUT("/passkey/settings", h.SettingHandler.UpdatePasskeySettings())

	appRouterGroup.FullAuthRouterGroup.GET("/webhook", h.SettingHandler.GetWebhook())
	appRouterGroup.FullAuthRouterGroup.POST("/webhook", h.SettingHandler.CreateWebhook())
	appRouterGroup.FullAuthRouterGroup.PUT("/webhook/:id", h.SettingHandler.UpdateWebhook())
	appRouterGroup.FullAuthRouterGroup.DELETE("/webhook/:id", h.SettingHandler.DeleteWebhook())
	appRouterGroup.FullAuthRouterGroup.POST("/webhook/:id/test", h.SettingHandler.TestWebhook())

	appRouterGroup.FullAuthRouterGroup.GET("/access-tokens", h.SettingHandler.ListAccessTokens())
	appRouterGroup.FullAuthRouterGroup.POST("/access-tokens", h.SettingHandler.CreateAccessToken())
	appRouterGroup.FullAuthRouterGroup.DELETE(
		"/access-tokens/:id",
		h.SettingHandler.DeleteAccessToken(),
	)

	appRouterGroup.FullAuthRouterGroup.GET(
		"/backup/schedule",
		h.SettingHandler.GetBackupScheduleSetting(),
	)
	appRouterGroup.FullAuthRouterGroup.POST(
		"/backup/schedule",
		h.SettingHandler.UpdateBackupScheduleSetting(),
	)

	appRouterGroup.FullAuthRouterGroup.GET("/agent/settings", h.SettingHandler.GetAgentSettings())
	appRouterGroup.FullAuthRouterGroup.PUT("/agent/settings", h.SettingHandler.UpdateAgentSettings())
}
