// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func setupDashboardRoutes(
	appRouterGroup *AppRouterGroup,
	h *handler.Bundle,
	revoker authService.TokenRevoker,
) {
	// Log streams authenticate with a query token, and high-privilege query
	// tokens are rejected at the transport layer. Keep both streams bound to
	// an interactive session; programmatic clients read the JSON log endpoint
	// with the admin:ops scope instead.
	appRouterGroup.AuthRouterGroup.GET(
		"/system/logs/stream",
		middleware.RequireSession(),
		middleware.RequireAdminSession(),
		h.DashboardHandler.SSESubscribeSystemLogs(),
	)
	appRouterGroup.WSRouterGroup.GET(
		"/system/logs",
		middleware.RequireAuth(revoker),
		middleware.RequireSession(),
		middleware.RequireAdminSession(),
		h.DashboardHandler.WSSubscribeSystemLogs(),
	)
}

func registerDashboard(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, secured(revoker, authz.ScopeAdminOps), huma.Operation{
		OperationID: "dashboard-check-update",
		Method:      http.MethodGet,
		Path:        "/system/check-update",
		Summary:     "检查 Ech0 版本更新",
		Tags:        []string{"Dashboard"},
	}, h.DashboardHandler.CheckUpdate)

	route(api, secured(revoker, authz.ScopeAdminOps), huma.Operation{
		OperationID: "dashboard-system-logs",
		Method:      http.MethodGet,
		Path:        "/system/logs",
		Summary:     "获取系统历史日志",
		Tags:        []string{"Dashboard"},
	}, h.DashboardHandler.GetSystemLogs)

	route(api, secured(revoker, authz.ScopeAdminOps), huma.Operation{
		OperationID: "dashboard-visitor-stats",
		Method:      http.MethodGet,
		Path:        "/system/visitor-stats",
		Summary:     "获取近七天访客统计",
		Tags:        []string{"Dashboard"},
	}, h.DashboardHandler.GetVisitorStats)
}
