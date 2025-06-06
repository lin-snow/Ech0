package router

import "github.com/lin-snow/ech0/internal/di"

func setupSettingRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/settings", h.SettingHandler.GetSettings)

	// Auth
	appRouterGroup.AuthRouterGroup.PUT("/settings", h.SettingHandler.UpdateSettings)
}
