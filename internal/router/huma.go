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
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

const (
	humaAPITitle   = "Ech0 API 文档"
	humaAPIVersion = "1.0"
	humaAPIBase    = "/api"
)

func setupHumaAPI(r *gin.Engine) huma.API {
	humaGroup := r.Group(humaAPIBase)
	docs := humares.ParseDocsRenderer(config.Config().OpenAPI.DocsRenderer)
	return humares.NewAPI(r, humaGroup, humaAPITitle, humaAPIVersion, humaAPIBase, docs)
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

func secured(revoker authService.TokenRevoker, scopes ...string) posture {
	mws := huma.Middlewares{
		humares.Bridge(middleware.NoCache()),
		humares.Bridge(middleware.RequireAuth(revoker)),
	}
	if len(scopes) > 0 {
		mws = append(mws, humares.Bridge(middleware.RequireScopes(scopes...)))
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

func route[I, T any](api huma.API, p posture, op huma.Operation, h func(context.Context, *I) (commonModel.Result[T], error)) {
	op.Security = p.security
	if len(p.middlewares) > 0 || len(op.Middlewares) > 0 {
		mws := make(huma.Middlewares, 0, len(p.middlewares)+len(op.Middlewares))
		mws = append(mws, p.middlewares...)
		mws = append(mws, op.Middlewares...)
		op.Middlewares = mws
	}
	huma.Register(api, op, humares.Wrap(h))
}

func registerOperations(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
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
}

func GenerateOpenAPIYAML() ([]byte, error) {
	gin.SetMode(gin.TestMode)
	api := setupHumaAPI(gin.New())
	registerOperations(api, &handler.Bundle{}, nil)
	return api.OpenAPI().YAML()
}
