// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"github.com/gin-gonic/gin"
)

func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header(
			"Cache-Control",
			"no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0",
		)
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Header("Surrogate-Control", "no-store")

		c.Next()
	}
}
