package router

import "github.com/lin-snow/ech0/internal/handler"

func setupAgentRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/agent/recent", h.AgentHandler.GetRecent())

	// Auth
}
