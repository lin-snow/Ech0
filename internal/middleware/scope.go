// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/authz"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// RequireSession rejects access tokens for operations that are inherently
// bound to an interactive browser session (account security, token
// administration). Session tokens carry no scopes, so this guard is the only
// authorization boundary those routes need.
func RequireSession() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if v.TokenType() == authModel.TokenTypeSession {
			ctx.Next()
			return
		}
		abortAuth(ctx, http.StatusForbidden, commonModel.ErrCodeScopeForbidden, commonModel.MsgKeyAuthScopeForbidden, commonModel.NO_PERMISSION_DENIED)
	}
}

// abortAuth writes the shared localized auth error and stops the chain.
func abortAuth(ctx *gin.Context, status int, code, messageKey, message string) {
	ctx.JSON(
		status,
		commonModel.FailWithLocalized[any](
			i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), messageKey, errUtil.HandleError(&commonModel.ServerError{
				Msg: message,
				Err: nil,
			}), nil),
			code,
			messageKey,
			nil,
		),
	)
	ctx.Abort()
}

func RequireScopes(scopes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if v.TokenType() == authModel.TokenTypeSession {
			ctx.Next()
			return
		}
		if v.TokenType() != authModel.TokenTypeAccess {
			abortAuth(ctx, http.StatusUnauthorized, commonModel.ErrCodeTokenInvalid, commonModel.MsgKeyAuthTokenInvalid, commonModel.TOKEN_NOT_VALID)
			return
		}
		if !containsValidAudience(v.Audience()) {
			abortAuth(ctx, http.StatusForbidden, commonModel.ErrCodeAudienceForbidden, commonModel.MsgKeyAuthAudienceForbidden, commonModel.NO_PERMISSION_DENIED)
			return
		}
		if !authz.Allows(v.Scopes(), scopes) {
			abortAuth(ctx, http.StatusForbidden, commonModel.ErrCodeScopeForbidden, commonModel.MsgKeyAuthScopeForbidden, commonModel.NO_PERMISSION_DENIED)
			return
		}
		ctx.Next()
	}
}

func RequireAudience(allowed ...string) gin.HandlerFunc {
	set := make(map[string]struct{}, len(allowed))
	for _, a := range allowed {
		set[a] = struct{}{}
	}
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if v.TokenType() == authModel.TokenTypeSession {
			ctx.Next()
			return
		}
		for _, aud := range v.Audience() {
			if _, ok := set[aud]; ok {
				ctx.Next()
				return
			}
		}
		abortAuth(ctx, http.StatusForbidden, commonModel.ErrCodeAudienceForbidden, commonModel.MsgKeyAuthAudienceForbidden, commonModel.NO_PERMISSION_DENIED)
	}
}

func containsValidAudience(audiences []string) bool {
	if len(audiences) == 0 {
		return false
	}
	return slices.ContainsFunc(audiences, authz.IsValidAudience)
}
