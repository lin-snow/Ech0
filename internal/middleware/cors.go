// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" && isValidOrigin(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization, Accept-Language, Range, X-Timezone, X-Locale, X-Direct-URL",
		)
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
		c.Header(
			"Access-Control-Expose-Headers",
			"Content-Length, Content-Range, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type",
		)
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isValidOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}

	u, err := url.Parse(origin)
	return err == nil && u.Scheme != "" && u.Hostname() != ""
}
