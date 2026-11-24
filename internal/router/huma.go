// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	"github.com/lin-snow/ech0/internal/ratelimit"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

const (
	humaAPITitle   = "Ech0 API 文档"
	humaAPIVersion = "1.0"
	humaAPIBase    = "/api"
)

// apiRegistry is the router's Huma API, the token revoker shared by every
// admission chain, the fallback request budget, and the declarations recorded
// while registering. The declarations exist so the router can verify itself at
// boot instead of trusting every call site to remember the same invariants.
type apiRegistry struct {
	huma.API
	limits    *ratelimit.MemoryLimiter
	budget    huma.Middlewares
	endpoints []declaredEndpoint
	revoker   authService.TokenRevoker
}

type declaredEndpoint struct {
	method   string
	path     string
	endpoint endpoint
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

// route registers one operation under its admission declaration. The
// declaration supplies the OpenAPI security requirement and the gate chain;
// whatever the operation adds (rate limits, request-specific middleware) runs
// inside that chain, and the fallback budget runs last.
func (a *apiRegistry) route[I, T any](e endpoint, op huma.Operation, h func(context.Context, *I) (*T, error)) {
	if e.access == AccessPublic && op.Description == "" {
		op.Description = e.reason
	}
	op.Security = e.security()

	chain := a.chain(e)
	mws := make(huma.Middlewares, 0, len(chain)+len(op.Middlewares)+len(a.budget))
	mws = append(mws, chain...)
	mws = append(mws, op.Middlewares...)
	mws = append(mws, a.budget...)
	op.Middlewares = mws

	a.endpoints = append(a.endpoints, declaredEndpoint{method: op.Method, path: op.Path, endpoint: e})
	huma.Register(a.API, op, h)
}

func registerOperations(api *apiRegistry, h *handler.Bundle, revoker authService.TokenRevoker) {
	api.revoker = revoker
	registerInit(api, h)
	registerAuth(api, h)
	registerCommon(api, h)
	registerEcho(api, h)
	registerConnect(api, h)
	registerUser(api, h)
	registerSetting(api, h)
	registerFile(api, h)
	registerDashboard(api, h)
	registerCopilot(api, h)
	registerComment(api, h)
	registerMigration(api, h)
	registerEmbedding(api, h)
	registerMCP(api, h)
}

// GenerateOpenAPISpec builds the spec from the same declarations the server
// registers, so the committed spec cannot describe a different gate.
func GenerateOpenAPIYAML() ([]byte, error) {
	gin.SetMode(gin.TestMode)
	api := setupHumaAPI(gin.New(), nil)
	registerOperations(api, &handler.Bundle{}, nil)
	if err := api.verify(); err != nil {
		return nil, err
	}
	return api.OpenAPI().YAML()
}
