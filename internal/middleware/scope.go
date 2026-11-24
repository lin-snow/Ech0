// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/httperr"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// RequireSession rejects access tokens for operations that are inherently
// bound to an interactive browser session (account security, token
// administration, OAuth consent, device approval).
func RequireSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if viewer.IsSession(v) {
			ctx.Next()
			return
		}
		abortAuth(ctx, http.StatusForbidden, commonModel.SCOPE_FORBIDDEN_MSG)
	}
}

// RequireMinRank is the admission gate for administrative surfaces. It asks one
// question — does the principal's live role reach min? — so a session is
// admitted by its current role and a token is admitted by the role read from
// the database, never by a scope it happens to carry. Resource-level decisions
// still run in the service.
func RequireMinRank(min role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if !v.Role().AtLeast(min) {
			abortAuth(ctx, http.StatusForbidden, commonModel.SCOPE_FORBIDDEN_MSG)
			return
		}
		ctx.Next()
	}
}

// abortAuth writes the shared auth problem document and stops the chain.
func abortAuth(ctx *gin.Context, status int, message string) {
	httperr.Abort(ctx, commonModel.Fail(status, message))
}

// RequireScopes checks the principal's capability set. Sessions derive their
// scopes from the live role at authentication time, so both credential kinds go
// through the same predicate: there is no session bypass to keep in sync with
// token scopes.
func RequireScopes(scopes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if viewer.IsAnonymous(v) {
			abortAuth(ctx, http.StatusUnauthorized, commonModel.TOKEN_NOT_VALID)
			return
		}
		if v.Kind() == viewer.KindAccess && !containsValidAudience(v.Audience()) {
			abortAuth(ctx, http.StatusForbidden, commonModel.AUDIENCE_FORBIDDEN_MSG)
			return
		}
		if !authz.Allows(v.Scopes(), scopes) {
			abortAuth(ctx, http.StatusForbidden, commonModel.SCOPE_FORBIDDEN_MSG)
			return
		}
		ctx.Next()
	}
}

func RequireAudience(allowed ...string) gin.HandlerFunc {
	return requireAudience(nil, allowed...)
}

// RequireAudienceWithChallenge is RequireAudience for resource servers that
// advertise OAuth discovery. A token signed for another audience is not
// acceptable here, so the request is rejected as 401 with an RFC 6750
// challenge instead of a bare 403: MCP clients only start (or restart) the
// authorization flow when they see a challenge on 401.
func RequireAudienceWithChallenge(
	challenge func(*gin.Context) string,
	allowed ...string,
) gin.HandlerFunc {
	return requireAudience(challenge, allowed...)
}

func requireAudience(challenge func(*gin.Context) string, allowed ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if viewer.IsSession(v) {
			ctx.Next()
			return
		}
		for _, aud := range v.Audience() {
			if _, ok := set[aud]; ok {
				ctx.Next()
				return
			}
		}
		if challenge != nil {
			ctx.Header("WWW-Authenticate", challenge(ctx))
			abortAuth(ctx, http.StatusUnauthorized, commonModel.TOKEN_NOT_VALID)
			return
		}
		abortAuth(ctx, http.StatusForbidden, commonModel.AUDIENCE_FORBIDDEN_MSG)
	}
}

func containsValidAudience(audiences []string) bool {
	if len(audiences) == 0 {
		return false
	}
	return slices.ContainsFunc(audiences, authz.IsValidAudience)
}
