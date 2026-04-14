// Package cookie 提供 refresh token 的 HttpOnly Cookie 读写工具。
//
// Cookie 安全策略：
//   - HttpOnly: JS 无法读取，防 XSS（核心防线）
//   - SameSite=Lax: 阻止 CSRF POST，同时兼容 OAuth 重定向链路
//   - Path=/api/auth: 仅在 refresh / logout / exchange 请求时携带，减少攻击面
//   - Secure=false: 自托管场景下大多数部署为反代 TLS 终止，内部链路为 HTTP，
//     强制 Secure 会导致 Cookie 无法送达。HttpOnly+SameSite+Path 已覆盖核心安全面。
package cookie

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RefreshTokenCookieName = "ech0_refresh_token"

// cookiePath 限制 Cookie 只在 /api/auth/* 路径下携带，
// 避免每次普通 API 请求都带上 refresh_token，减少攻击面。
const cookiePath = "/api/auth"

// SetRefreshTokenCookie 将 refresh token 写入 HttpOnly Cookie。
func SetRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     cookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false,
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
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetRefreshTokenFromCookie 从 Cookie 中读取 refresh token。
func GetRefreshTokenFromCookie(c *gin.Context) (string, error) {
	return c.Cookie(RefreshTokenCookieName)
}
