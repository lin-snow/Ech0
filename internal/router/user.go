package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// setupUserRoutes 设置用户路由
func setupUserRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// OAuth2/OIDC (统一 provider 路由)
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/login", middleware.NoCache(), h.UserHandler.OAuthLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/callback", middleware.NoCache(), h.UserHandler.OAuthCallback())

	// Public
	appRouterGroup.PublicRouterGroup.POST("/login", middleware.NoCache(), h.UserHandler.Login())
	appRouterGroup.PublicRouterGroup.POST("/register", middleware.NoCache(), h.UserHandler.Register())
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
	appRouterGroup.AuthRouterGroup.GET(
		"/allusers",
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
		middleware.RequireScopes(authModel.ScopeProfileRead),
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
	appRouterGroup.AuthRouterGroup.POST(
		"/oauth/:provider/bind",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.OAuthBind(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/oauth/info",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.GetOAuthInfo(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/begin",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.PasskeyRegisterBeginV2(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/finish",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.PasskeyRegisterFinishV2(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/passkeys",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.ListPasskeys(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/passkeys/:id",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.DeletePasskey(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/passkeys/:id",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.UserHandler.UpdatePasskeyDeviceName(),
	)
}
