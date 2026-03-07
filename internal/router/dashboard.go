package router

import "github.com/lin-snow/ech0/internal/handler"

func setupDashboardRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Auth
	appRouterGroup.AuthRouterGroup.GET("/dashboard/metrics", h.DashboardHandler.GetMetrics())
	appRouterGroup.WSRouterGroup.GET("/dashboard/metrics", h.DashboardHandler.WSSubsribeMetrics())
}
