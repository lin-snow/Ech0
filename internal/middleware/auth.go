// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/httperr"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type rejection struct {
	status int
	err    error
}

func reject(err error) *rejection {
	return &rejection{status: httperr.Code(err), err: err}
}

func resolveViewer(ctx *gin.Context, tokenBlacklist authService.TokenRevoker) (rej *rejection, hard bool) {
	auth := strings.TrimSpace(ctx.Request.Header.Get("Authorization"))
	tokenFromQuery := false
	if auth == "" {
		queryToken := strings.TrimSpace(ctx.Query("token"))
		queryToken = strings.Trim(queryToken, `"`)
		if queryToken != "" && queryToken != "null" && queryToken != "undefined" {
			auth = "Bearer " + queryToken
			tokenFromQuery = true
		}
	}

	parts := strings.SplitN(auth, " ", 2)

	if auth == "" || len(parts) != 2 || len(parts[1]) == 0 || parts[1] == "null" || parts[1] == "undefined" {
		return reject(apperr.ErrTokenMissing), false
	}
	if parts[0] != "Bearer" {
		return reject(apperr.ErrTokenInvalid), false
	}

	// The bearer is either a session or an access token; each parser pins its
	// own type, issuer, and (for sessions) instance audience, so a token can
	// never be accepted as the wrong kind.
	mc, err := jwtUtil.ParseAccessToken(parts[1])
	if err != nil {
		mc, err = jwtUtil.ParseSessionToken(parts[1])
	}
	if err != nil {
		return reject(apperr.ErrTokenUnparsable), false
	}

	if tokenBlacklist != nil && mc.ID != "" && tokenBlacklist.IsTokenRevoked(mc.ID) {
		return reject(apperr.ErrTokenRevoked), false
	}

	if tokenFromQuery && authz.HasAdminScope(mc.Scopes) {
		return reject(apperr.ErrTokenQueryForbidden), true
	}

	// Role and status live in the database, not in the signed token: a token
	// minted before a demotion, an ownership transfer or a disable must stop
	// working on the next request.
	state := userModel.AccountState{Found: true}
	if tokenBlacklist != nil {
		var err error
		state, err = tokenBlacklist.AccountState(ctx.Request.Context(), mc.Subject)
		if err != nil {
			return reject(apperr.ErrAccountState), true
		}
		if !state.Active() {
			return reject(apperr.ErrAccountDisabled), true
		}
	}

	var principal viewer.Principal
	switch mc.Type {
	case authModel.TokenTypeSession:
		// A session's capabilities are its role's capabilities, derived on every
		// request. The token carries no scopes, so a promotion or demotion takes
		// effect immediately without a re-login.
		principal = viewer.NewSession(mc.Subject, state.Role, authz.ScopesForRole(state.Role))
	case authModel.TokenTypeAccess:
		// An access token was minted with an explicit, frozen capability set.
		// The role is still live: it bounds rank-based decisions.
		principal = viewer.NewAccess(mc.Subject, mc.ID, state.Role, mc.Scopes, []string(mc.Audience))
	default:
		return reject(apperr.ErrTokenUnparsable), false
	}

	viewer.AttachToRequest(&ctx.Request, principal)
	i18nUtil.ApplyUserLocaleFromUserID(ctx, mc.Subject)
	return nil, false
}

func writeRejection(ctx *gin.Context, rej *rejection) {
	httperr.Abort(ctx, rej.err)
}

func RequireAuth(tokenBlacklist authService.TokenRevoker) gin.HandlerFunc {
	return requireAuth(tokenBlacklist, nil)
}

// RequireAuthWithChallenge behaves like RequireAuth but decorates 401
// responses with an RFC 6750 bearertoken challenge. MCP uses it to advertise
// the protected resource metadata URL so clients can start OAuth discovery.
func RequireAuthWithChallenge(
	tokenBlacklist authService.TokenRevoker,
	challenge func(*gin.Context) string,
) gin.HandlerFunc {
	return requireAuth(tokenBlacklist, challenge)
}

func requireAuth(
	tokenBlacklist authService.TokenRevoker,
	challenge func(*gin.Context) string,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rej, _ := resolveViewer(ctx, tokenBlacklist)
		if rej != nil {
			if challenge != nil && rej.status == http.StatusUnauthorized {
				ctx.Header("WWW-Authenticate", challenge(ctx))
			}
			writeRejection(ctx, rej)
			return
		}
		ctx.Next()
	}
}

func OptionalAuth(tokenBlacklist authService.TokenRevoker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rej, hard := resolveViewer(ctx, tokenBlacklist)
		if rej != nil {
			if hard {
				writeRejection(ctx, rej)
				return
			}
			viewer.AttachToRequest(&ctx.Request, viewer.Anonymous)
		}
		ctx.Next()
	}
}
