// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	openapiCatalog "github.com/lin-snow/ech0/internal/openapi/catalog"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

// catalogState holds the operation catalog built from the Huma document, so the
// CLI can discover commands instead of hard-coding them per server version.
type catalogState struct {
	mu          sync.RWMutex
	body        []byte
	fingerprint string
}

func newCatalogState() *catalogState {
	return &catalogState{}
}

func (state *catalogState) build(api huma.API) {
	_, body, fingerprint, err := openapiCatalog.Build(api.OpenAPI())
	if err != nil {
		logUtil.Warn("operation catalog unavailable", slog.String("module", "router"), slog.String("error", err.Error()))
		return
	}
	state.mu.Lock()
	state.body = body
	state.fingerprint = fingerprint
	state.mu.Unlock()
}

func (state *catalogState) current() ([]byte, string) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.body, state.fingerprint
}

// middleware stamps the catalog fingerprint on every /api response and rejects
// requests built against another catalog with 412, before the handler runs.
func (state *catalogState) middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		_, fingerprint := state.current()
		if fingerprint == "" || !strings.HasPrefix(path, "/api/") {
			ctx.Next()
			return
		}
		ctx.Header(openapiCatalog.HeaderName, fingerprint)

		if path == "/api"+openapiCatalog.RoutePath {
			ctx.Next()
			return
		}
		if sent := ctx.GetHeader(openapiCatalog.HeaderName); sent != "" && sent != fingerprint {
			result := commonModel.Fail[any]("catalog stale")
			result.ErrorCode = openapiCatalog.StaleErrorCode
			ctx.AbortWithStatusJSON(http.StatusPreconditionFailed, result)
			return
		}
		ctx.Next()
	}
}

// handle serves the canonical catalog bytes for CLI discovery.
func (state *catalogState) handle(ctx *gin.Context) {
	body, fingerprint := state.current()
	if len(body) == 0 {
		result := commonModel.Fail[any]("operation catalog unavailable")
		ctx.JSON(http.StatusServiceUnavailable, result)
		return
	}
	ctx.Header(openapiCatalog.HeaderName, fingerprint)
	ctx.Header("ETag", `"`+fingerprint+`"`)
	ctx.Header("Cache-Control", "no-cache")
	ctx.Data(http.StatusOK, "application/json; charset=utf-8", body)
}
