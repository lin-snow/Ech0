// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

// setupAuthzRoutes mounts the MCP authorization server. Protocol endpoints live
// at their spec-mandated root paths and bypass Ech0's result envelope; the
// consent endpoints stay under /api/authz so the SPA can authenticate them with
// the normal browser session.
func setupAuthzRoutes(r *gin.Engine, groups *AppRouterGroup, h *handler.Bundle) {
	// Discovery documents are safe to cache; the handler sets
	// `Cache-Control: public, max-age=3600` itself.
	r.GET("/.well-known/oauth-protected-resource", h.AuthzHandler.ProtectedResource())
	r.GET("/.well-known/oauth-protected-resource/mcp", h.AuthzHandler.ProtectedResource())
	r.GET("/.well-known/oauth-authorization-server", h.AuthzHandler.AuthorizationServerMetadata())

	// Rate limiting registration keeps an open DCR endpoint from being used to
	// fill the client table.
	r.POST("/oauth/register", middleware.NoCache(), middleware.RateLimit(5, 10), h.AuthzHandler.Register())
	r.POST("/oauth/token", middleware.NoCache(), h.AuthzHandler.Token())

	groups.AuthRouterGroup.GET("/authz/consent", middleware.NoCache(), h.AuthzHandler.Consent())
	groups.AuthRouterGroup.POST("/authz/authorize", middleware.NoCache(), h.AuthzHandler.Approve())
}
