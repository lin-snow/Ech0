// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusForbidden, gin.H{"error": "origin not allowed"})
		c.Abort()
	}
}

func isSameOrigin(origin string, r *http.Request) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	if !strings.EqualFold(parsed.Host, r.Host) {
		return false
	}
	return strings.EqualFold(parsed.Scheme, requestScheme(r))
}

func requestScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return strings.ToLower(strings.TrimSpace(strings.Split(proto, ",")[0]))
	}
	return "http"
}
