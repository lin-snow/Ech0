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

// tokenBlacklist 模块级黑名单引用，由 DI 层在启动时通过 SetTokenBlacklist 注入。
// 中间件在解析 JWT 成功后会检查其 JTI 是否已被吊销。
var tokenBlacklist authService.TokenRevoker

// SetTokenBlacklist 注入黑名单实例，必须在路由注册前调用（由 wire_gen.go 中的 BuildHandlers 触发）。
func SetTokenBlacklist(bl authService.TokenRevoker) {
	tokenBlacklist = bl
}

// JWTAuthMiddleware 是核心鉴权中间件，处理流程：
//  1. 从 Authorization header 或 query ?token= 中提取 JWT
//  2. ParseToken() 验证签名与过期（仅接受 typ=session / typ=access）
//  3. 检查 JTI 是否在黑名单中（已登出/已吊销的 token）
//  4. 校验通过后，将 viewer (用户身份) 写入 request context
//
// 部分公开路由（echo 查询等）允许匿名降级：即使 token 无效也不阻断请求。
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		setAnonymous := func() {
			viewer.AttachToRequest(&ctx.Request, viewer.NewNoopViewer())
		}
		allowAnonymousForCurrentRoute := func() bool {
			path := ctx.Request.URL.Path
			method := ctx.Request.Method
			// 统一查询接口（POST /api/echo/query）
			if path == "/api/echo/query" && method == http.MethodPost {
				return true
			}
			// 分页获取首页 Echo
			if strings.HasPrefix(path, "/api/echo/page") {
				return true
			}
			// 获取当日 Echo
			if strings.HasPrefix(path, "/api/echo/today") {
				return true
			}
			// 查看 Echo 详情
			if strings.HasPrefix(path, "/api/echo") && method == http.MethodGet {
				return true
			}
			// 根据 Tag ID 获取 Echo 列表
			if strings.HasPrefix(path, "/api/echo/tag/") && method == http.MethodGet {
				return true
			}
			return false
		}

		// 获取 Authorization 头部信息，若缺失则回退到 query token（用于 <audio>/<video> 直链等场景）
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

		// 将 Authorization 头部信息分割成两部分
		parts := strings.SplitN(auth, " ", 2)

		// 如果 Authorization 头部信息为空，或者格式不正确，或者 token 为空，则返回错误
		if auth == "" || len(parts) != 2 || len(parts[1]) == 0 || parts[1] == "null" ||
			parts[1] == "undefined" {
			if allowAnonymousForCurrentRoute() {
				setAnonymous()
				ctx.Next()
				return
			}

			// 如果 Authorization 头部信息为空，或者格式不正确，或者 token 为空，则返回错误
			ctx.JSON(
				http.StatusUnauthorized,
				commonModel.FailWithLocalized[any](
					i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), commonModel.MsgKeyAuthTokenMissing, errUtil.HandleError(&commonModel.ServerError{
						Msg: commonModel.TOKEN_NOT_FOUND,
						Err: nil,
					}), nil),
					commonModel.ErrCodeTokenMissing,
					commonModel.MsgKeyAuthTokenMissing,
					nil,
				),
			)
			ctx.Abort()
			return
		}

		// 如果 Authorization 头部信息格式不正确，或者 token 格式不正确，则返回错误
		if len(parts) != 2 || parts[0] != "Bearer" {
			if allowAnonymousForCurrentRoute() {
				setAnonymous()
				ctx.Next()
				return
			}
			ctx.JSON(
				http.StatusUnauthorized,
				commonModel.FailWithLocalized[any](
					i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), commonModel.MsgKeyAuthTokenInvalid, errUtil.HandleError(&commonModel.ServerError{
						Msg: commonModel.TOKEN_NOT_VALID,
						Err: nil,
					}), nil),
					commonModel.ErrCodeTokenInvalid,
					commonModel.MsgKeyAuthTokenInvalid,
					nil,
				),
			)
			ctx.Abort()
			return
		}

		// 解析 token
		mc, err := jwtUtil.ParseToken(parts[1])
		if err != nil {
			// 允许匿名访问的公开路由，即使带了无效 token 也按匿名降级，避免公开页被历史 token 卡住。
			if allowAnonymousForCurrentRoute() {
				setAnonymous()
				ctx.Next()
				return
			}
			// 如果 token 解析失败，则返回错误
			ctx.JSON(
				http.StatusUnauthorized,
				commonModel.FailWithLocalized[any](
					i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), commonModel.MsgKeyAuthTokenParse, errUtil.HandleError(&commonModel.ServerError{
						Msg: commonModel.TOKEN_PARSE_ERROR,
						Err: err,
					}), nil),
					commonModel.ErrCodeTokenParse,
					commonModel.MsgKeyAuthTokenParse,
					nil,
				),
			)
			ctx.Abort()
			return
		}

		// 黑名单检查：已登出或已吊销的 token 即使签名有效也应拒绝。
		// mc.ID 即 JWT 的 jti claim，登出时由 AuthHandler.Logout() 写入黑名单。
		if tokenBlacklist != nil && mc.ID != "" && tokenBlacklist.IsTokenRevoked(mc.ID) {
			if allowAnonymousForCurrentRoute() {
				setAnonymous()
				ctx.Next()
				return
			}
			ctx.JSON(
				http.StatusUnauthorized,
				commonModel.FailWithLocalized[any](
					i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), commonModel.MsgKeyAuthTokenRevoked, errUtil.HandleError(&commonModel.ServerError{
						Msg: commonModel.TOKEN_REVOKED,
						Err: nil,
					}), nil),
					commonModel.ErrCodeTokenRevoked,
					commonModel.MsgKeyAuthTokenRevoked,
					nil,
				),
			)
			ctx.Abort()
			return
		}

		if tokenFromQuery && authModel.HasAdminScope(mc.Scopes) {
			ctx.JSON(
				http.StatusForbidden,
				commonModel.FailWithLocalized[any](
					i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), commonModel.MsgKeyAuthTokenTransportForbidden, errUtil.HandleError(&commonModel.ServerError{
						Msg: commonModel.NO_PERMISSION_DENIED,
						Err: nil,
					}), nil),
					commonModel.ErrCodeTokenTransportForbidden,
					commonModel.MsgKeyAuthTokenTransportForbidden,
					nil,
				),
			)
			ctx.Abort()
			return
		}

		// 如果 token 解析成功，则将 viewer 写入 request context
		viewer.AttachToRequest(
			&ctx.Request,
			viewer.NewUserViewerWithToken(mc.Userid, mc.Type, mc.Scopes, []string(mc.Audience), mc.ID),
		)
		// 鉴权成功后，若请求未显式指定语言，则按用户偏好覆盖语言上下文
		i18nUtil.ApplyUserLocaleFromUserID(ctx, mc.Userid)
		ctx.Next()
	}
}
