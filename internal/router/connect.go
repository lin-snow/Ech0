package router

import "github.com/lin-snow/ech0/internal/di"

func setupConnectRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/connect/list", h.ConnectHandler.GetConnects)
	appRouterGroup.PublicRouterGroup.GET("/connects/info", h.ConnectHandler.GetConnectsInfo)

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/addConnect", h.ConnectHandler.AddConnect)
	appRouterGroup.AuthRouterGroup.DELETE("/delConnect/:id", h.ConnectHandler.DeleteConnect)
}
