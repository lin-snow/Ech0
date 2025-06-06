package router

import "github.com/lin-snow/ech0/internal/di"

func setupUserRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public
	appRouterGroup.PublicRouterGroup.POST("/login", h.UserHandler.Login)
	appRouterGroup.PublicRouterGroup.POST("/register", h.UserHandler.Register)
	appRouterGroup.PublicRouterGroup.GET("/allusers", h.UserHandler.GetAllUsers)

	// Auth
	appRouterGroup.AuthRouterGroup.GET("/user", h.UserHandler.GetUserInfo)
	appRouterGroup.AuthRouterGroup.PUT("/user", h.UserHandler.UpdateUser)
	appRouterGroup.AuthRouterGroup.DELETE("/user/:id", h.UserHandler.DeleteUser)
	appRouterGroup.AuthRouterGroup.GET("/user/admin/:id", h.UserHandler.UpdateUserAdmin)
}
