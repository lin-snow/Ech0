package router

import "github.com/lin-snow/ech0/internal/di"

func setupConnectRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/addConnect")
	appRouterGroup.AuthRouterGroup.DELETE("/delConnect/:id")
}
