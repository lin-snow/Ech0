// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/config"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	userService "github.com/lin-snow/ech0/internal/service/user"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

// Repeated failures of the auth surface. Sentinels are immutable, so one
// value per case is safe to share instead of rebuilding it per request.
var (
	errInvalidBody         = apperr.ErrInvalidRequestBody
	errInvalidParams       = apperr.ErrInvalidParams
	errRefreshTokenInvalid = apperr.ErrRefreshTokenInvalid
	errAccountDisabled     = apperr.ErrAccountDisabled
	errTokenGenerateFailed = apperr.ErrTokenIssueFailed
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
	return res.Execute(func(ctx *gin.Context) res.Response {
		refreshTokenStr, err := cookieUtil.GetRefreshTokenFromCookie(ctx)
		if err != nil || refreshTokenStr == "" {
			return res.Failed(errRefreshTokenInvalid)
		}

		claims, err := jwtUtil.ParseRefreshToken(refreshTokenStr)
		if err != nil {
			cookieUtil.ClearRefreshTokenCookie(ctx)
			return res.Failed(errRefreshTokenInvalid)
		}

		if h.authService.IsTokenRevoked(claims.ID) {
			cookieUtil.ClearRefreshTokenCookie(ctx)
			return res.Failed(apperr.ErrTokenRevoked)
		}

		user, err := h.userService.GetUserByID(claims.Subject)
		if err != nil {
			cookieUtil.ClearRefreshTokenCookie(ctx)
			return res.Failed(errRefreshTokenInvalid)
		}

		state, err := h.authService.AccountState(ctx.Request.Context(), user.ID)
		if err != nil {
			return res.Failed(apperr.ErrInternal)
		}
		if !state.Active() {
			cookieUtil.ClearRefreshTokenCookie(ctx)
			return res.Failed(errAccountDisabled)
		}

		accessClaims := jwtUtil.CreateClaims(user)
		accessToken, err := jwtUtil.GenerateToken(accessClaims)
		if err != nil {
			return res.Failed(errTokenGenerateFailed)
		}

		renewedRefreshToken, err := jwtUtil.GenerateToken(jwtUtil.RenewRefreshClaims(user, claims.ID))
		if err != nil {
			return res.Failed(errTokenGenerateFailed)
		}

		cookieUtil.SetRefreshTokenCookie(ctx, renewedRefreshToken, config.Config().Auth.Jwt.RefreshExpires)

		return res.Payload(authModel.TokenPair{
			AccessToken: accessToken,
			ExpiresIn:   config.Config().Auth.Jwt.Expires,
		})
	})
}

func (h *AuthHandler) Logout() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		refreshTokenStr, _ := cookieUtil.GetRefreshTokenFromCookie(ctx)
		if refreshTokenStr != "" {
			if claims, err := jwtUtil.ParseRefreshToken(refreshTokenStr); err == nil && claims.ID != "" {
				h.authService.RevokeToken(claims.ID, remainingTTLFromClaims(claims.ExpiresAt))
			}
		}

		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			if claims, err := jwtUtil.ParseSessionToken(authHeader[7:]); err == nil && claims.ID != "" {
				h.authService.RevokeToken(claims.ID, remainingTTLFromClaims(claims.ExpiresAt))
			}
		}

		cookieUtil.ClearRefreshTokenCookie(ctx)
		return res.NoContent(http.StatusOK)
	})
}

func remainingTTLFromClaims(expiresAt *jwt.NumericDate) time.Duration {
	const legacyNeverFallback = 100 * 365 * 24 * time.Hour
	if expiresAt == nil {
		return legacyNeverFallback
	}
	return time.Until(expiresAt.Time)
}

func (h *AuthHandler) Exchange() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.ExchangeCodeReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Failed(errInvalidBody)
		}

		tokenPair, err := h.authService.ExchangeOAuthCode(req.Code)
		if err != nil {
			return res.Failed(apperr.ErrExchangeCodeInvalid)
		}

		cookieUtil.SetRefreshTokenCookie(ctx, tokenPair.RefreshToken, config.Config().Auth.Jwt.RefreshExpires)

		return res.Payload(authModel.TokenPair{
			AccessToken: tokenPair.AccessToken,
			ExpiresIn:   tokenPair.ExpiresIn,
		})
	})
}
