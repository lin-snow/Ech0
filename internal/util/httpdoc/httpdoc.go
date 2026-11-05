// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package httpdoc writes standalone HTTP documents (discovery metadata,
// well-known files) together with their cache validators.
package httpdoc

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Write serves body as a complete document.
//
// The body is derived from the request origin, so it is revalidated rather than
// cached fresh: a spoofed X-Forwarded-Host cannot pin a wrong origin into a
// shared cache, and a 304 still saves the round trip.
func Write(ctx *gin.Context, contentType string, body []byte) {
	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`

	ctx.Header("ETag", etag)
	ctx.Header("Cache-Control", "public, max-age=0, must-revalidate")

	if matchesETag(ctx.GetHeader("If-None-Match"), etag) {
		ctx.Status(http.StatusNotModified)
		return
	}

	ctx.Data(http.StatusOK, contentType, body)
}

func matchesETag(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimPrefix(strings.TrimSpace(candidate), "W/")
		if candidate == "*" || candidate == etag {
			return true
		}
	}

	return false
}
