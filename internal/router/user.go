package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupUserRoutes 设置用户路由
func setupUserRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.PublicRouterGroup.POST("/register", middleware.NoCache(), h.UserHandler.Register())

	// Auth
	appRouterGroup.AuthRouterGroup.GET(
		"/users",
		middleware.RequireScopes(authModel.ScopeAdminUser),
		h.UserHandler.GetAllUsers(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/user",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.GetUserInfo(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/user",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.UserHandler.UpdateUser(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/user/:id",
		middleware.RequireScopes(authModel.ScopeAdminUser),
		h.UserHandler.DeleteUser(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/user/admin/:id",
		middleware.RequireScopes(authModel.ScopeAdminUser),
		h.UserHandler.UpdateUserAdmin(),
	)
}
