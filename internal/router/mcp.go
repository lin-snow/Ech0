// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/mcp"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func setupMCPRoutes(groups *AppRouterGroup, h *handler.Bundle) {
	g := groups.MCPRouterGroup
	g.Use(
		middleware.RateLimit(20, 40),
		middleware.OriginGuard(config.Config().Web.CORS.AllowedOrigins),
		middleware.RequireAudience(authModel.AudienceMCPRemote),
	)
	g.POST("", h.MCPHandler.ServeEndpoint())
	g.GET("", h.MCPHandler.ServeEndpoint())
	g.DELETE("", h.MCPHandler.ServeEndpoint())
}

func registerMCP(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, secured(revoker, authModel.ScopeAdminToken), huma.Operation{
		OperationID: "mcp-manifest",
		Method:      http.MethodGet,
		Path:        mcp.MCPEndpointPath + "/manifest",
		Summary:     "获取 MCP 端点信息与能力清单",
		Tags:        []string{"MCP"},
	}, h.MCPHandler.GetManifest)
}
