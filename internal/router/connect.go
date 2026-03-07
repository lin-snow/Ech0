package router

import "github.com/lin-snow/ech0/internal/handler"

// setupConnectRoutes 设置连接路由
func setupConnectRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/connect", h.ConnectHandler.GetConnect())
	appRouterGroup.PublicRouterGroup.GET("/connect/list", h.ConnectHandler.GetConnects())
	appRouterGroup.PublicRouterGroup.GET("/connects/info", h.ConnectHandler.GetConnectsInfo())

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/addConnect", h.ConnectHandler.AddConnect())
	appRouterGroup.AuthRouterGroup.DELETE("/delConnect/:id", h.ConnectHandler.DeleteConnect())
}
