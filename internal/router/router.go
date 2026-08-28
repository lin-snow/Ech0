// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/mcp"
	"github.com/lin-snow/ech0/internal/middleware"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

type AppRouterGroup struct {
	ResourceGroup           *gin.RouterGroup
	PublicRouterGroup       *gin.RouterGroup
	AuthRouterGroup         *gin.RouterGroup
	OptionalAuthRouterGroup *gin.RouterGroup
	WSRouterGroup           *gin.RouterGroup
	MCPRouterGroup          *gin.RouterGroup
}

func SetupRouter(r *gin.Engine, h *handler.Bundle, mwDeps *middleware.Deps) {
	setupTemplateRoutes(r, h)
	setupStaticFiles(r)
	setupMiddleware(r)
	groups := setupRouterGroup(r, mwDeps)
	api := setupHumaAPI(r)

	revoker := revokerOf(mwDeps)
	setupResourceRoutes(groups, h)
	setupAuthRoutes(groups, h)
	setupCommentRoutes(groups, h)
	setupFileRoutes(groups, h)
	setupDashboardRoutes(groups, h)
	setupCopilotRoutes(groups, h)
	registerOperations(api, h, revoker)
	setupMigrationRoutes(groups, h)
	setupMCPRoutes(groups, h)
}

func setupStaticFiles(r *gin.Engine) {
	root := config.Config().Storage.DataRoot
	if root == "" {
		root = "data/files"
	}
	r.Group("api/files", middleware.StaticFileSecurity()).StaticFS("/", http.Dir(root))
}

func revokerOf(mwDeps *middleware.Deps) authService.TokenRevoker {
	if mwDeps != nil {
		return mwDeps.TokenRevoker
	}
	return nil
}

func setupRouterGroup(r *gin.Engine, mwDeps *middleware.Deps) *AppRouterGroup {
	revoker := revokerOf(mwDeps)

	resource := r.Group("/")
	public := r.Group("/api")
	auth := r.Group("/api")
	auth.Use(middleware.NoCache(), middleware.RequireAuth(revoker))
	optionalAuth := r.Group("/api")
	optionalAuth.Use(middleware.NoCache(), middleware.OptionalAuth(revoker))
	ws := r.Group("/ws")
	mcpGroup := r.Group(mcp.MCPEndpointPath)
	mcpGroup.Use(middleware.NoCache(), middleware.RequireAuth(revoker))
	return &AppRouterGroup{
		ResourceGroup:           resource,
		PublicRouterGroup:       public,
		AuthRouterGroup:         auth,
		OptionalAuthRouterGroup: optionalAuth,
		WSRouterGroup:           ws,
		MCPRouterGroup:          mcpGroup,
	}
}
