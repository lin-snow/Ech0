package router

import "github.com/lin-snow/ech0/internal/handler"

// setupUserRoutes 设置用户路由
func setupUserRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// OAuth2
	appRouterGroup.ResourceGroup.GET("/oauth/github/login", h.UserHandler.GitHubLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/github/callback", h.UserHandler.GitHubCallback())
	appRouterGroup.ResourceGroup.GET("/oauth/google/login", h.UserHandler.GoogleLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/google/callback", h.UserHandler.GoogleCallback())
	appRouterGroup.ResourceGroup.GET("/oauth/qq/login", h.UserHandler.QQLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/qq/callback", h.UserHandler.QQCallback())
	appRouterGroup.ResourceGroup.GET("/oauth/custom/login", h.UserHandler.CustomOAuthLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/custom/callback", h.UserHandler.CustomOAuthCallback())

	// Public
	appRouterGroup.PublicRouterGroup.POST("/login", h.UserHandler.Login())
	appRouterGroup.PublicRouterGroup.POST("/register", h.UserHandler.Register())
	appRouterGroup.PublicRouterGroup.GET("/allusers", h.UserHandler.GetAllUsers())
	appRouterGroup.PublicRouterGroup.POST("/passkey/login/begin", h.UserHandler.PasskeyLoginBegin())
	appRouterGroup.PublicRouterGroup.POST(
		"/passkey/login/finish",
		h.UserHandler.PasskeyLoginFinish(),
	)

	// Auth
	appRouterGroup.AuthRouterGroup.GET("/user", h.UserHandler.GetUserInfo())
	appRouterGroup.AuthRouterGroup.PUT("/user", h.UserHandler.UpdateUser())
	appRouterGroup.AuthRouterGroup.DELETE("/user/:id", h.UserHandler.DeleteUser())
	appRouterGroup.AuthRouterGroup.PUT("/user/admin/:id", h.UserHandler.UpdateUserAdmin())
	appRouterGroup.AuthRouterGroup.POST("/oauth/github/bind", h.UserHandler.BindGitHub())
	appRouterGroup.AuthRouterGroup.POST("/oauth/google/bind", h.UserHandler.BindGoogle())
	appRouterGroup.AuthRouterGroup.POST("/oauth/qq/bind", h.UserHandler.BindQQ())
	appRouterGroup.AuthRouterGroup.POST("/oauth/custom/bind", h.UserHandler.BindCustomOAuth())
	appRouterGroup.AuthRouterGroup.GET("/oauth/info", h.UserHandler.GetOAuthInfo())
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/begin",
		h.UserHandler.PasskeyRegisterBegin(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/finish",
		h.UserHandler.PasskeyRegisterFinish(),
	)
	appRouterGroup.AuthRouterGroup.GET("/passkeys", h.UserHandler.ListPasskeys())
	appRouterGroup.AuthRouterGroup.DELETE("/passkeys/:id", h.UserHandler.DeletePasskey())
	appRouterGroup.AuthRouterGroup.PUT("/passkeys/:id", h.UserHandler.UpdatePasskeyDeviceName())
}
