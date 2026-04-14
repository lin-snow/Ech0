package router

import (
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func setupAuthRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// OAuth2/OIDC (统一 provider 路由)
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/login", middleware.NoCache(), h.AuthHandler.OAuthLogin())
	appRouterGroup.ResourceGroup.GET("/oauth/:provider/callback", middleware.NoCache(), h.AuthHandler.OAuthCallback())

	// Public
	appRouterGroup.PublicRouterGroup.POST("/login", middleware.NoCache(), h.AuthHandler.Login())
	appRouterGroup.PublicRouterGroup.POST(
		"/passkey/login/begin",
		middleware.NoCache(),
		h.AuthHandler.PasskeyLoginBeginV2(),
	)
	appRouterGroup.PublicRouterGroup.POST(
		"/passkey/login/finish",
		middleware.NoCache(),
		h.AuthHandler.PasskeyLoginFinishV2(),
	)

	// Token lifecycle
	appRouterGroup.PublicRouterGroup.POST("/auth/refresh", middleware.NoCache(), h.AuthHandler.Refresh())
	appRouterGroup.PublicRouterGroup.POST("/auth/logout", middleware.NoCache(), h.AuthHandler.Logout())
	appRouterGroup.PublicRouterGroup.POST("/auth/exchange", middleware.NoCache(), h.AuthHandler.Exchange())

	// Auth
	appRouterGroup.AuthRouterGroup.POST(
		"/oauth/:provider/bind",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.OAuthBind(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/oauth/info",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.AuthHandler.GetOAuthInfo(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/begin",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.PasskeyRegisterBeginV2(),
	)
	appRouterGroup.AuthRouterGroup.POST(
		"/passkey/register/finish",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.PasskeyRegisterFinishV2(),
	)
	appRouterGroup.AuthRouterGroup.GET(
		"/passkeys",
		middleware.RequireScopes(authModel.ScopeProfileRead),
		h.AuthHandler.ListPasskeys(),
	)
	appRouterGroup.AuthRouterGroup.DELETE(
		"/passkeys/:id",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.DeletePasskey(),
	)
	appRouterGroup.AuthRouterGroup.PUT(
		"/passkeys/:id",
		middleware.RequireScopes(authModel.ScopeProfileWrite),
		h.AuthHandler.UpdatePasskeyDeviceName(),
	)
}
