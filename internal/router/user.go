package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

// setupUserRoutes 设置用户路由
func setupUserRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// OAuth2/OIDC (统一 provider 路由)
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/login", middleware.NoCache(), h.UserHandler.OAuthLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/callback", middleware.NoCache(), h.UserHandler.OAuthCallback())

	// Public
	appRouterGroup.PublicRouterGroup.POST("/login", middleware.NoCache(), h.UserHandler.Login())
	appRouterGroup.PublicRouterGroup.POST("/register", middleware.NoCache(), h.UserHandler.Register())
	appRouterGroup.PublicRouterGroup.GET("/allusers", h.UserHandler.GetAllUsers())
	appRouterGroup.PublicRouterGroup.POST(
		"/passkey/login/begin",
		middleware.NoCache(),
		h.UserHandler.PasskeyLoginBeginV2(),
	)
	appRouterGroup.PublicRouterGroup.POST(
		"/passkey/login/finish",
		middleware.NoCache(),
		h.UserHandler.PasskeyLoginFinishV2(),
	)

	// Auth
	appRouterGroup.AuthRouterGroup.GET("/user", h.UserHandler.GetUserInfo())
	appRouterGroup.FullAuthRouterGroup.PUT("/user", h.UserHandler.UpdateUser())
	appRouterGroup.FullAuthRouterGroup.DELETE("/user/:id", h.UserHandler.DeleteUser())
	appRouterGroup.FullAuthRouterGroup.PUT("/user/admin/:id", h.UserHandler.UpdateUserAdmin())
	appRouterGroup.FullAuthRouterGroup.POST("/oauth/:provider/bind", h.UserHandler.OAuthBind())
	appRouterGroup.FullAuthRouterGroup.GET("/oauth/info", h.UserHandler.GetOAuthInfo())
	appRouterGroup.FullAuthRouterGroup.POST(
		"/passkey/register/begin",
		h.UserHandler.PasskeyRegisterBeginV2(),
	)
	appRouterGroup.FullAuthRouterGroup.POST(
		"/passkey/register/finish",
		h.UserHandler.PasskeyRegisterFinishV2(),
	)
	appRouterGroup.FullAuthRouterGroup.GET("/passkeys", h.UserHandler.ListPasskeys())
	appRouterGroup.FullAuthRouterGroup.DELETE("/passkeys/:id", h.UserHandler.DeletePasskey())
	appRouterGroup.FullAuthRouterGroup.PUT("/passkeys/:id", h.UserHandler.UpdatePasskeyDeviceName())
}
