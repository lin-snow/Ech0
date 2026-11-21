// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/ratelimit"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

const (
	humaAPITitle   = "Ech0 API 文档"
	humaAPIVersion = "1.0"
	humaAPIBase    = "/api"
)

// apiRegistry is the router's Huma API plus the fallback request budget that
// every operation inherits. Keeping the budget here means register functions
// stay unchanged while the budget still runs after each operation's auth
// middleware, where the caller identity is known.
type apiRegistry struct {
	huma.API
	limits *ratelimit.MemoryLimiter
	budget huma.Middlewares
}

func setupHumaAPI(r *gin.Engine, limits *ratelimit.MemoryLimiter) *apiRegistry {
	humaGroup := r.Group(humaAPIBase)
	docs := humares.ParseDocsRenderer(config.Config().OpenAPI.DocsRenderer)
	registry := &apiRegistry{API: humares.NewAPI(r, humaGroup, humaAPITitle, humaAPIVersion, humaAPIBase, docs), limits: limits}
	if limits != nil {
		registry.budget = huma.Middlewares{humares.Bridge(middleware.Budget(limits))}
	}
	return registry
}

type posture struct {
	security    []map[string][]string
	middlewares huma.Middlewares
}

func public() posture { return posture{} }

func optional(revoker authService.TokenRevoker) posture {
	return posture{middlewares: huma.Middlewares{
		humares.Bridge(middleware.NoCache()),
		humares.Bridge(middleware.OptionalAuth(revoker)),
	}}
}

// sessionOnly marks an endpoint as requiring an interactive session. It is
// security-silent about scopes on purpose: access tokens must never reach
// account-security or token-administration operations, regardless of scopes.
func sessionOnly(revoker authService.TokenRevoker) posture {
	return posture{
		security: humares.Secured(),
		middlewares: huma.Middlewares{
			humares.Bridge(middleware.NoCache()),
			humares.Bridge(middleware.RequireAuth(revoker)),
			humares.Bridge(middleware.RequireSession()),
		},
	}
}

func secured(revoker authService.TokenRevoker, scopes ...string) posture {
	mws := huma.Middlewares{
		humares.Bridge(middleware.NoCache()),
		humares.Bridge(middleware.RequireAuth(revoker)),
	}
	if len(scopes) > 0 {
		mws = append(mws, humares.Bridge(middleware.RequireScopes(scopes...)))
	}
	if authz.HasAdminScope(scopes) {
		mws = append(mws, humares.Bridge(middleware.RequireAdminSession()))
	}
	return posture{security: humares.Secured(scopes...), middlewares: mws}
}

func (p posture) audience(auds ...string) posture {
	p.middlewares = append(p.middlewares, humares.Bridge(middleware.RequireAudience(auds...)))
	return p
}

func noCache() huma.Middlewares {
	return huma.Middlewares{humares.Bridge(middleware.NoCache())}
}

// scopeLimit charges one named budget on top of the fallback budget.
func (a *apiRegistry) scopeLimit(scope ratelimit.Scope, cost int) huma.Middlewares {
	if a.limits == nil {
		return nil
	}
	return huma.Middlewares{humares.Bridge(middleware.ScopeLimit(a.limits, scope, cost))}
}

// route registers one operation with the posture's middleware chain, the
// operation's own middleware, and finally the fallback budget.
func (a *apiRegistry) route[I, T any](p posture, op huma.Operation, h func(context.Context, *I) (commonModel.Result[T], error)) {
	op.Security = p.security
	mws := make(huma.Middlewares, 0, len(p.middlewares)+len(op.Middlewares)+len(a.budget))
	mws = append(mws, p.middlewares...)
	mws = append(mws, op.Middlewares...)
	mws = append(mws, a.budget...)
	op.Middlewares = mws
	huma.Register(a.API, op, humares.Wrap(h))
}

func registerOperations(api *apiRegistry, h *handler.Bundle, revoker authService.TokenRevoker) {
	registerInit(api, h)
	registerAuth(api, h, revoker)
	registerCommon(api, h, revoker)
	registerEcho(api, h, revoker)
	registerConnect(api, h, revoker)
	registerUser(api, h, revoker)
	registerSetting(api, h, revoker)
	registerFile(api, h, revoker)
	registerDashboard(api, h, revoker)
	registerCopilot(api, h, revoker)
	registerComment(api, h, revoker)
	registerMigration(api, h, revoker)
	registerEmbedding(api, h, revoker)
	registerMCP(api, h, revoker)
}

func GenerateOpenAPIYAML() ([]byte, error) {
	gin.SetMode(gin.TestMode)
	api := setupHumaAPI(gin.New(), nil)
	registerOperations(api, &handler.Bundle{}, nil)
	return api.OpenAPI().YAML()
}
