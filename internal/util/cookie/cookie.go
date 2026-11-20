// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cookie

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
)

const RefreshTokenCookieName = "ech0_refresh_token"

const cookiePath = "/api/auth"

func isHTTPS(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
		return true
	}
	if strings.HasPrefix(c.GetHeader("Origin"), "https://") {
		return true
	}
	if strings.HasPrefix(c.GetHeader("Referer"), "https://") {
		return true
	}
	return false
}

func SetRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     cookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isHTTPS(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearRefreshTokenCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isHTTPS(c),
		SameSite: http.SameSiteLaxMode,
	})
}

func GetRefreshTokenFromCookie(c *gin.Context) (string, error) {
	return c.Cookie(RefreshTokenCookieName)
}

const MediaCookieName = "ech0_media_session"

func SetMediaCookie(ctx *gin.Context, token string, maxAge int) {
	if !config.Config().Upload.VerifiedDirect {
		return
	}
	http.SetCookie(ctx.Writer, &http.Cookie{Name: MediaCookieName, Value: token, MaxAge: maxAge, Path: "/media/", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
}
