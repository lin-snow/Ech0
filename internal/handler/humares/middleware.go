// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

type ctxKey int

const localizerKey ctxKey = iota

func injectLocalizer(ctx huma.Context, next func(huma.Context)) {
	gctx := humagin.Unwrap(ctx)
	provider := func() *goi18n.Localizer { return i18nUtil.LocalizerFromGin(gctx) }
	gctx.Request = gctx.Request.WithContext(context.WithValue(gctx.Request.Context(), localizerKey, provider))
	next(ctx)
}

func localizerFrom(ctx context.Context) *goi18n.Localizer {
	if provider, ok := ctx.Value(localizerKey).(func() *goi18n.Localizer); ok {
		return provider()
	}
	return nil
}

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
