// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/httperr"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/util/reqinfo"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

func OriginGuard(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		if trimmed := strings.TrimSuffix(strings.TrimSpace(o), "/"); trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		origin := strings.TrimSuffix(c.GetHeader("Origin"), "/")
		if origin == "" {
			c.Next()
			return
		}
		if _, ok := allowed[origin]; ok {
			c.Next()
			return
		}
		if isSameOrigin(origin, c.Request) {
			c.Next()
			return
		}
		logUtil.GetLogger().Warn("blocked cross-origin request",
			slog.String("module", "middleware"),
			slog.String("origin", origin),
			slog.String("host", c.Request.Host),
			slog.String("path", c.Request.URL.Path),
			slog.String("hint", "add this origin to ECH0_WEB_CORS_ALLOWED_ORIGINS if it is a trusted browser client"),
		)
		httperr.Abort(c, commonModel.Fail(http.StatusForbidden, "origin not allowed"))
	}
}

func isSameOrigin(origin string, r *http.Request) bool {
	return reqinfo.IsSameOrigin(origin, r)
}
