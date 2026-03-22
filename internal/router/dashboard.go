package router

import "github.com/lin-snow/ech0/internal/handler"

func setupDashboardRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Auth
	appRouterGroup.FullAuthRouterGroup.GET("/system/logs", h.DashboardHandler.GetSystemLogs())
	appRouterGroup.FullAuthRouterGroup.GET("/system/logs/stream", h.DashboardHandler.SSESubscribeSystemLogs())
	appRouterGroup.WSRouterGroup.GET("/system/logs", h.DashboardHandler.WSSubscribeSystemLogs())
}
