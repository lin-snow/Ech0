// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"mime"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var inlineableMIMEPrefixes = []string{
	"image/",
	"audio/",
	"video/",
}

func StaticFileSecurity() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")

		ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
		if isInlineableExt(ext) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			basename := filepath.Base(c.Request.URL.Path)
			c.Header("Content-Disposition", "attachment; filename=\""+basename+"\"")
		}

		c.Next()
	}
}

func isInlineableExt(ext string) bool {
	ct := mime.TypeByExtension(ext)
	return ct != "" && isInlineableMIME(ct)
}

func isInlineableMIME(ct string) bool {
	lower := strings.ToLower(ct)
	for _, prefix := range inlineableMIMEPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
