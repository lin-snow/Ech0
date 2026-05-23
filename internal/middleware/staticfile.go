// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"mime"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// inlineableMIMEPrefixes lists Content-Type prefixes that are safe to display
// inline in the browser (images and audio only).
var inlineableMIMEPrefixes = []string{
	"image/",
	"audio/",
}

// StaticFileSecurity returns a middleware that hardens responses served from the
// public file endpoint:
//   - X-Content-Type-Options: nosniff (prevents MIME-sniffing attacks).
//   - Content-Disposition: attachment for non-image/audio files (forces download
//     instead of in-browser execution).
//   - Cache-Control: long-lived immutable cache for inlineable assets (image/audio).
//     Stored filenames are content-hashed (see storage layer), so reusing a key
//     implies identical bytes.
//
// All response headers must be set before c.Next(): http.FileServer's first body
// Write triggers an implicit WriteHeader(200) that flushes the header map to the
// socket; later c.Header(...) calls mutate the map but do not re-send headers
// (the Range-request path through ServeContent flushes especially early, which
// is why browsers were missing Cache-Control even when curl saw it).
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

// isInlineableExt resolves the URL extension via the same MIME table that
// http.ServeContent uses, so our decision matches the Content-Type that will be
// written downstream.
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
