package router

import "github.com/lin-snow/ech0/internal/handler"

// setupEchoRoutes 设置Echo路由
func setupEchoRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.PUT("/echo/like/:id", h.EchoHandler.LikeEcho())
	appRouterGroup.PublicRouterGroup.GET("/tags", h.EchoHandler.GetAllTags())

	// Auth
	appRouterGroup.FullAuthRouterGroup.POST("/echo", h.EchoHandler.PostEcho())
	appRouterGroup.FullAuthRouterGroup.GET("/echo/page", h.EchoHandler.GetEchosByPage())
	appRouterGroup.FullAuthRouterGroup.POST("/echo/page", h.EchoHandler.GetEchosByPage())
	appRouterGroup.FullAuthRouterGroup.DELETE("/echo/:id", h.EchoHandler.DeleteEcho())
	appRouterGroup.FullAuthRouterGroup.GET("/echo/today", h.EchoHandler.GetTodayEchos())
	appRouterGroup.FullAuthRouterGroup.PUT("/echo", h.EchoHandler.UpdateEcho())
	appRouterGroup.FullAuthRouterGroup.GET("/echo/:id", h.EchoHandler.GetEchoById())
	appRouterGroup.FullAuthRouterGroup.GET("/echo/tag/:tagid", h.EchoHandler.GetEchosByTagId())
	appRouterGroup.FullAuthRouterGroup.DELETE("/tag/:id", h.EchoHandler.DeleteTag())
	// appRouterGroup.FullAuthRouterGroup.PUT("/tag", h.EchoHandler.UpdateTag())
}
