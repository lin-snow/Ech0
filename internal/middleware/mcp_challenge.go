// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/util/reqinfo"
)

// ResourceMetadataChallenge builds a Bearer challenge pointing at the OAuth
// 2.0 Protected Resource Metadata document (RFC 9728). The origin is derived
// from the request so the value stays correct behind a reverse proxy that
// forwards X-Forwarded-Proto / X-Forwarded-Host.
func ResourceMetadataChallenge(path string) func(*gin.Context) string {
	return func(ctx *gin.Context) string {
		return fmt.Sprintf(`Bearer resource_metadata="%s"`, reqinfo.Origin(ctx.Request)+path)
	}
}

// InvalidTokenChallenge is ResourceMetadataChallenge for a token that was
// validly signed but is not acceptable for this resource (for example the wrong
// audience). RFC 6750 clients retry the authorization flow on 401 with
// error="invalid_token" instead of treating the call as permanently forbidden.
func InvalidTokenChallenge(path string) func(*gin.Context) string {
	return func(ctx *gin.Context) string {
		return fmt.Sprintf(`Bearer error="invalid_token", resource_metadata="%s"`,
			reqinfo.Origin(ctx.Request)+path)
	}
}
