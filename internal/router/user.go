package router

import "github.com/lin-snow/ech0/internal/di"

func setupUserRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public
	appRouterGroup.PublicRouterGroup.POST("/login", h.UserHandler.Login)

	// Auth
}
