package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupEchoRoutes 设置Echo路由
func setupEchoRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.PUT("/echo/like/:id", h.EchoHandler.LikeEcho())
	appRouterGroup.PublicRouterGroup.GET("/tags", h.EchoHandler.GetAllTags())

	// Auth
	// 读接口保留“可匿名降级”行为：无 token 或无效 token 时由 JWT 中间件降级为匿名用户继续访问。
	appRouterGroup.AuthRouterGroup.GET("/echo/page", h.EchoHandler.GetEchosByPage())
	appRouterGroup.AuthRouterGroup.POST("/echo/page", h.EchoHandler.GetEchosByPage())
	appRouterGroup.AuthRouterGroup.GET("/echo/today", h.EchoHandler.GetTodayEchos())
	appRouterGroup.AuthRouterGroup.GET("/echo/:id", h.EchoHandler.GetEchoById())
	appRouterGroup.AuthRouterGroup.GET("/echo/tag/:tagid", h.EchoHandler.GetEchosByTagId())
	appRouterGroup.AuthRouterGroup.POST(
		"/echo",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.PostEcho(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/echo",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.UpdateEcho(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/echo/:id",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.DeleteEcho(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/tag/:id",
		middleware.RequireScopes(authModel.ScopeEchoWrite),
		h.EchoHandler.DeleteTag(),
	)
	// appRouterGroup.AuthRouterGroup.PUT("/tag", h.EchoHandler.UpdateTag())
}
