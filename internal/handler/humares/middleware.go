// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

// Bridge turns a gin middleware into a Huma middleware so the posture chain
// declared at registration time runs on both surfaces.
func Bridge(h gin.HandlerFunc) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		gctx := humagin.Unwrap(ctx)
		h(gctx)
		if gctx.IsAborted() {
			return
		}
		next(ctx)
	}
}
