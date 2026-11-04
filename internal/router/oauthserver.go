// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
)

// setupOAuthServerRoutes mounts the MCP authorization server. Protocol
// endpoints live at their spec-mandated root paths and bypass Ech0's result
// envelope; the consent endpoints stay under /api/consent so the SPA can
// authenticate them with the normal browser session.
func setupOAuthServerRoutes(r *gin.Engine, groups *AppRouterGroup, h *handler.Bundle) {
	// Discovery documents are safe to cache; the handler sets
	// `Cache-Control: public, max-age=3600` itself.
	r.GET(oauthserverModel.ProtectedResourcePath, h.OAuthServerHandler.ProtectedResource())
	r.GET(oauthserverModel.ProtectedResourcePathForMCP, h.OAuthServerHandler.ProtectedResource())
	r.GET(oauthserverModel.AuthorizationServerMetadataPath, h.OAuthServerHandler.AuthorizationServerMetadata())

	// Rate limiting registration keeps an open DCR endpoint from being used to
	// fill the client table.
	r.POST(oauthserverModel.RegistrationPath, middleware.NoCache(), middleware.RateLimit(5, 10), h.OAuthServerHandler.Register())
	// The token endpoint is unauthenticated apart from the code and PKCE
	// verifier, so cap guessing attempts per client IP as well.
	r.POST(oauthserverModel.TokenPath, middleware.NoCache(), middleware.RateLimit(10, 20), h.OAuthServerHandler.Token())

	groups.AuthRouterGroup.GET(
		"/consent",
		middleware.NoCache(),
		middleware.RequireSession(),
		h.OAuthServerHandler.Consent(),
	)
	groups.AuthRouterGroup.POST(
		"/consent/approve",
		middleware.NoCache(),
		middleware.RequireSession(),
		h.OAuthServerHandler.Approve(),
	)
}
