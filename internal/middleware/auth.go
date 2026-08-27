// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type rejection struct {
	status  int
	errCode string
	msgKey  string
	msg     string
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
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenMissing, commonModel.MsgKeyAuthTokenMissing, commonModel.TOKEN_NOT_FOUND}, false
	}
	if parts[0] != "Bearer" {
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenInvalid, commonModel.MsgKeyAuthTokenInvalid, commonModel.TOKEN_NOT_VALID}, false
	}

	mc, err := jwtUtil.ParseToken(parts[1])
	if err != nil {
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenParse, commonModel.MsgKeyAuthTokenParse, commonModel.TOKEN_PARSE_ERROR}, false
	}

	if tokenBlacklist != nil && mc.ID != "" && tokenBlacklist.IsTokenRevoked(mc.ID) {
		return &rejection{http.StatusUnauthorized, commonModel.ErrCodeTokenRevoked, commonModel.MsgKeyAuthTokenRevoked, commonModel.TOKEN_REVOKED}, false
	}

	if tokenFromQuery && authModel.HasAdminScope(mc.Scopes) {
		return &rejection{http.StatusForbidden, commonModel.ErrCodeTokenTransportForbidden, commonModel.MsgKeyAuthTokenTransportForbidden, commonModel.NO_PERMISSION_DENIED}, true
	}

	viewer.AttachToRequest(
		&ctx.Request,
		viewer.NewUserViewerWithToken(mc.Userid, mc.Type, mc.Scopes, []string(mc.Audience), mc.ID),
	)
	i18nUtil.ApplyUserLocaleFromUserID(ctx, mc.Userid)
	return nil, false
}

func writeRejection(ctx *gin.Context, rej *rejection) {
	msg := i18nUtil.Localize(
		i18nUtil.LocalizerFromGin(ctx),
		rej.msgKey,
		errUtil.HandleError(&commonModel.ServerError{Msg: rej.msg, Err: nil}),
		nil,
	)
	ctx.JSON(rej.status, commonModel.FailWithLocalized[any](msg, rej.errCode, rej.msgKey, nil))
	ctx.Abort()
}

func RequireAuth(tokenBlacklist authService.TokenRevoker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if rej, _ := resolveViewer(ctx, tokenBlacklist); rej != nil {
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
			viewer.AttachToRequest(&ctx.Request, viewer.NewNoopViewer())
		}
		ctx.Next()
	}
}
