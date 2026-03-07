package router

import "github.com/lin-snow/ech0/internal/handler"

// setupEchoRoutes 设置Echo路由
func setupEchoRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.PUT("/echo/like/:id", h.EchoHandler.LikeEcho())
	appRouterGroup.PublicRouterGroup.GET("/tags", h.EchoHandler.GetAllTags())

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/echo", h.EchoHandler.PostEcho())
	appRouterGroup.AuthRouterGroup.GET("/echo/page", h.EchoHandler.GetEchosByPage())
	appRouterGroup.AuthRouterGroup.POST("/echo/page", h.EchoHandler.GetEchosByPage())
	appRouterGroup.AuthRouterGroup.DELETE("/echo/:id", h.EchoHandler.DeleteEcho())
	appRouterGroup.AuthRouterGroup.GET("/echo/today", h.EchoHandler.GetTodayEchos())
	appRouterGroup.AuthRouterGroup.PUT("/echo", h.EchoHandler.UpdateEcho())
	appRouterGroup.AuthRouterGroup.GET("/echo/:id", h.EchoHandler.GetEchoById())
	appRouterGroup.AuthRouterGroup.GET("/echo/tag/:tagid", h.EchoHandler.GetEchosByTagId())
	appRouterGroup.AuthRouterGroup.DELETE("/tag/:id", h.EchoHandler.DeleteTag())
	// appRouterGroup.AuthRouterGroup.PUT("/tag", h.EchoHandler.UpdateTag())
}
