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
	appRouterGroup.AuthRouterGroup.GET(
		"/echo/page",
		middleware.RequireScopes(authModel.ScopeEchoRead),
		h.EchoHandler.GetEchosByPage(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/echo/page",
		middleware.RequireScopes(authModel.ScopeEchoRead),
		h.EchoHandler.GetEchosByPage(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/echo/today",
		middleware.RequireScopes(authModel.ScopeEchoRead),
		h.EchoHandler.GetTodayEchos(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/echo/:id",
		middleware.RequireScopes(authModel.ScopeEchoRead),
		h.EchoHandler.GetEchoById(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/echo/tag/:tagid",
		middleware.RequireScopes(authModel.ScopeEchoRead),
		h.EchoHandler.GetEchosByTagId(),
	)
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
