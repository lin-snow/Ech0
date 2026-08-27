// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"github.com/gin-gonic/gin"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

func PoweredBy() gin.HandlerFunc {
	value := "Ech0/" + versionPkg.Version
	return func(c *gin.Context) {
		c.Header("X-Powered-By", value)
		c.Next()
	}
}
