// Package cookie 提供 refresh token 的 HttpOnly Cookie 读写工具。
//
// Cookie 安全策略：
//   - HttpOnly: JS 无法读取，防 XSS（核心防线）
//   - SameSite=Lax: 阻止 CSRF POST，同时兼容 OAuth 重定向链路
//   - Path=/api/auth: 仅在 refresh / logout / exchange 请求时携带，减少攻击面
//   - Secure: 根据请求来源自动判断——HTTPS 时开启，HTTP 时关闭，
//     兼顾自托管反代 TLS 终止场景和生产 HTTPS 部署的安全性。
package cookie

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const RefreshTokenCookieName = "ech0_refresh_token"

// cookiePath 限制 Cookie 只在 /api/auth/* 路径下携带，
// 避免每次普通 API 请求都带上 refresh_token，减少攻击面。
const cookiePath = "/api/auth"

// isHTTPS 通过多种信号判断当前请求是否经由 HTTPS 到达。
// 检查顺序：TLS 直连 → X-Forwarded-Proto（反代） → Origin / Referer 头。
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

// SetRefreshTokenCookie 将 refresh token 写入 HttpOnly Cookie。
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

// ClearRefreshTokenCookie 清除 refresh token Cookie。
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

// GetRefreshTokenFromCookie 从 Cookie 中读取 refresh token。
func GetRefreshTokenFromCookie(c *gin.Context) (string, error) {
	return c.Cookie(RefreshTokenCookieName)
}
