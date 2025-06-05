package router

import "github.com/lin-snow/ech0/internal/di"

func setupEchoRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/echo", h.EchoHandler.PostEcho)
}
