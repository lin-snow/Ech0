// Package handler 提供认证相关的 HTTP 端点：token 刷新、登出、OAuth code 交换。
//
// 这三个端点均为公开路由（不经过 JWTAuthMiddleware），因为：
//   - /refresh 通过 HttpOnly Cookie 中的 refresh_token 鉴权
//   - /logout 是 best-effort 操作（无 Cookie 也返回 200）
//   - /exchange 通过一次性 code 鉴权（code 由 OAuth 回调流程生成）
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	userService "github.com/lin-snow/ech0/internal/service/user"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

// AuthHandler 处理 token 生命周期管理（刷新、吊销、OAuth code 交换）。
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

// Refresh 静默刷新 access_token。
//
// 流程：Cookie 中读取 refresh_token → 验证签名与过期 → 检查黑名单
// → 查询用户是否仍然存在 → 签发新的 access_token 并返回。
//
// 前端在页面加载和 401 响应时自动调用此端点，使用 credentials:'include'
// 让浏览器自动携带 HttpOnly Cookie。
func (h *AuthHandler) Refresh() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshTokenStr, err := cookieUtil.GetRefreshTokenFromCookie(ctx)
		if err != nil || refreshTokenStr == "" {
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithErrorCode[any](
				commonModel.REFRESH_TOKEN_INVALID,
				commonModel.ErrCodeRefreshTokenInvalid,
			))
			return
		}

		claims, err := jwtUtil.ParseRefreshToken(refreshTokenStr)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithErrorCode[any](
				commonModel.REFRESH_TOKEN_INVALID,
				commonModel.ErrCodeRefreshTokenInvalid,
			))
			return
		}

		if h.authService.IsTokenRevoked(claims.ID) {
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithErrorCode[any](
				commonModel.TOKEN_REVOKED,
				commonModel.ErrCodeTokenRevoked,
			))
			return
		}

		user, err := h.userService.GetUserByID(claims.Userid)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, commonModel.FailWithErrorCode[any](
				commonModel.USER_NOTFOUND,
				commonModel.ErrCodeRefreshTokenInvalid,
			))
			return
		}

		accessClaims := jwtUtil.CreateClaims(user)
		accessToken, err := jwtUtil.GenerateToken(accessClaims)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, commonModel.Fail[any]("failed to generate access token"))
			return
		}

		ctx.JSON(http.StatusOK, commonModel.OK(authModel.TokenPair{
			AccessToken: accessToken,
			ExpiresIn:   config.Config().Auth.Jwt.Expires,
		}))
	}
}

// Logout 吊销当前会话。
//
// 分两步吊销：
//  1. 从 Cookie 中读取 refresh_token，将其 JTI 加入黑名单（TTL = 剩余有效期）
//  2. 如果请求携带了 Authorization header，也将 access_token JTI 加入黑名单
//
// 最后清除 Cookie 并返回 200。即使没有有效 Cookie 也不报错（best-effort）。
func (h *AuthHandler) Logout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 吊销 refresh_token
		refreshTokenStr, _ := cookieUtil.GetRefreshTokenFromCookie(ctx)
		if refreshTokenStr != "" {
			if claims, err := jwtUtil.ParseRefreshToken(refreshTokenStr); err == nil && claims.ID != "" {
				remaining := time.Until(claims.ExpiresAt.Time)
				h.authService.RevokeToken(claims.ID, remaining)
			}
		}

		// 吊销 access_token（可选，前端 logout 时会在 header 中携带）
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			if claims, err := jwtUtil.ParseToken(authHeader[7:]); err == nil && claims.ID != "" {
				remaining := time.Until(claims.ExpiresAt.Time)
				h.authService.RevokeToken(claims.ID, remaining)
			}
		}

		cookieUtil.ClearRefreshTokenCookie(ctx)
		ctx.JSON(http.StatusOK, commonModel.OK[any](nil))
	}
}

// Exchange 用一次性 code 换取 token pair（OAuth 回调专用）。
//
// OAuth 回调流程：IdP callback → 后端签发 TokenPair 并存入缓存（key=随机 code, TTL=60s）
// → 302 重定向到前端 /auth?code=xxx → 前端调用本端点用 code 换取 token。
//
// code 为一次性使用：取出后立即从缓存中删除，过期也会自动淘汰。
func (h *AuthHandler) Exchange() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req authModel.ExchangeCodeReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, commonModel.FailWithErrorCode[any](
				commonModel.INVALID_REQUEST_BODY,
				commonModel.ErrCodeInvalidRequest,
			))
			return
		}

		tokenPair, err := h.authService.ExchangeOAuthCode(req.Code)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, commonModel.FailWithErrorCode[any](
				commonModel.EXCHANGE_CODE_INVALID,
				commonModel.ErrCodeExchangeCodeInvalid,
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
