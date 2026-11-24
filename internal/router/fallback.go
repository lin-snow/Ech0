// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/httperr"
)

// setupFallbacks answers the two routing misses a REST client has to tell
// apart: a path that exists under another method (405, with the Allow header
// gin fills in) and a path that does not exist at all (the SPA shell, or a
// problem document for /api).
func setupFallbacks(r *gin.Engine, notFound gin.HandlerFunc) {
	r.HandleMethodNotAllowed = true
	r.NoMethod(func(ctx *gin.Context) {
		httperr.Abort(ctx, httperr.Edge(
			http.StatusMethodNotAllowed,
			ctx.Request.Method+" is not allowed on "+ctx.Request.URL.Path+": see the Allow header",
		))
	})
	r.NoRoute(notFound)
}
