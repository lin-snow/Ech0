// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	userService "github.com/lin-snow/ech0/internal/service/user"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

type AuthHandler struct {
	authService authService.Service
	userService userService.Service
}

func NewAuthHandler(authService authService.Service, userService userService.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

func (h *AuthHandler) Refresh() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		localizer := i18nUtil.LocalizerFromGin(ctx)

		refreshTokenStr, err := cookieUtil.GetRefreshTokenFromCookie(ctx)
		if err != nil || refreshTokenStr == "" {
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithLocalized[any](
				i18nUtil.Localize(localizer, commonModel.MsgKeyAuthRefreshTokenInvalid, errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.REFRESH_TOKEN_INVALID, Err: err,
				}), nil),
				commonModel.ErrCodeRefreshTokenInvalid,
				commonModel.MsgKeyAuthRefreshTokenInvalid,
				nil,
			))
			return
		}

		claims, err := jwtUtil.ParseRefreshToken(refreshTokenStr)
		if err != nil {
			cookieUtil.ClearRefreshTokenCookie(ctx)
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithLocalized[any](
				i18nUtil.Localize(localizer, commonModel.MsgKeyAuthRefreshTokenInvalid, errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.REFRESH_TOKEN_INVALID, Err: err,
				}), nil),
				commonModel.ErrCodeRefreshTokenInvalid,
				commonModel.MsgKeyAuthRefreshTokenInvalid,
				nil,
			))
			return
		}

		if h.authService.IsTokenRevoked(claims.ID) {
			cookieUtil.ClearRefreshTokenCookie(ctx)
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithLocalized[any](
				i18nUtil.Localize(localizer, commonModel.MsgKeyAuthTokenRevoked, errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.TOKEN_REVOKED, Err: nil,
				}), nil),
				commonModel.ErrCodeTokenRevoked,
				commonModel.MsgKeyAuthTokenRevoked,
				nil,
			))
			return
		}

		user, err := h.userService.GetUserByID(claims.Userid)
		if err != nil {
			cookieUtil.ClearRefreshTokenCookie(ctx)
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithLocalized[any](
				i18nUtil.Localize(localizer, commonModel.MsgKeyAuthRefreshTokenInvalid, errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.REFRESH_TOKEN_INVALID, Err: err,
				}), nil),
				commonModel.ErrCodeRefreshTokenInvalid,
				commonModel.MsgKeyAuthRefreshTokenInvalid,
				nil,
			))
			return
		}

		accessClaims := jwtUtil.CreateClaims(user)
		accessToken, err := jwtUtil.GenerateToken(accessClaims)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, commonModel.FailWithLocalized[any](
				i18nUtil.Localize(localizer, commonModel.MsgKeyAuthTokenGenerateFailed, errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.TOKEN_GENERATE_FAILED, Err: err,
				}), nil),
				commonModel.ErrCodeTokenGenerateFailed,
				commonModel.MsgKeyAuthTokenGenerateFailed,
				nil,
			))
			return
		}

		ctx.JSON(http.StatusOK, commonModel.OK(authModel.TokenPair{
			AccessToken: accessToken,
			ExpiresIn:   config.Config().Auth.Jwt.Expires,
		}))
	}
}

func (h *AuthHandler) Logout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshTokenStr, _ := cookieUtil.GetRefreshTokenFromCookie(ctx)
		if refreshTokenStr != "" {
			if claims, err := jwtUtil.ParseRefreshToken(refreshTokenStr); err == nil && claims.ID != "" {
				h.authService.RevokeToken(claims.ID, remainingTTLFromClaims(claims.ExpiresAt))
			}
		}

		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			if claims, err := jwtUtil.ParseToken(authHeader[7:]); err == nil && claims.ID != "" {
				h.authService.RevokeToken(claims.ID, remainingTTLFromClaims(claims.ExpiresAt))
			}
		}

		cookieUtil.ClearRefreshTokenCookie(ctx)
		ctx.JSON(http.StatusOK, commonModel.OK[any](nil))
	}
}

func remainingTTLFromClaims(expiresAt *jwt.NumericDate) time.Duration {
	const legacyNeverFallback = 100 * 365 * 24 * time.Hour
	if expiresAt == nil {
		return legacyNeverFallback
	}
	return time.Until(expiresAt.Time)
}

func (h *AuthHandler) Exchange() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		localizer := i18nUtil.LocalizerFromGin(ctx)

		var req authModel.ExchangeCodeReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, commonModel.FailWithLocalized[any](
				i18nUtil.Localize(localizer, commonModel.MsgKeyCommonRequestFailed, errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.INVALID_REQUEST_BODY, Err: err,
				}), nil),
				commonModel.ErrCodeInvalidRequest,
				commonModel.MsgKeyCommonRequestFailed,
				nil,
			))
			return
		}

		tokenPair, err := h.authService.ExchangeOAuthCode(req.Code)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, commonModel.FailWithLocalized[any](
				i18nUtil.Localize(localizer, commonModel.MsgKeyAuthExchangeCodeInvalid, errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.EXCHANGE_CODE_INVALID, Err: err,
				}), nil),
				commonModel.ErrCodeExchangeCodeInvalid,
				commonModel.MsgKeyAuthExchangeCodeInvalid,
				nil,
			))
			return
		}

		cookieUtil.SetRefreshTokenCookie(ctx, tokenPair.RefreshToken, config.Config().Auth.Jwt.RefreshExpires)

		ctx.JSON(http.StatusOK, commonModel.OK(authModel.TokenPair{
			AccessToken: tokenPair.AccessToken,
			ExpiresIn:   tokenPair.ExpiresIn,
		}))
	}
}
