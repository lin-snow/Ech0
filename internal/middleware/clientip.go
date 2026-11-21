// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/clientip"
)

// ClientIP resolves the request address once and stores it on the request
// context so middlewares and services share one answer.
func ClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientip.Resolve(c.Request.RemoteAddr, c.Request.Header)
		if ip != "" {
			c.Request = c.Request.WithContext(clientip.WithClientIP(c.Request.Context(), ip))
		}
		c.Next()
	}
}
