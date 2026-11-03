// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// ResourceMetadataChallenge builds a Bearer challenge pointing at the OAuth
// 2.0 Protected Resource Metadata document (RFC 9728). The origin is derived
// from the request so the value stays correct behind a reverse proxy that
// forwards X-Forwarded-Proto.
func ResourceMetadataChallenge(path string) func(*gin.Context) string {
	return func(ctx *gin.Context) string {
		return fmt.Sprintf(`Bearer resource_metadata="%s://%s%s"`,
			requestScheme(ctx.Request), ctx.Request.Host, path)
	}
}
