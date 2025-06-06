package middleware

import (
	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"net/http"
	"strings"
)

// JWTAuthMiddleware JWT 拦截器中间件
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.Request.Header.Get("Authorization")

		parts := strings.SplitN(auth, " ", 2)

		if auth == "" || len(parts) != 2 || len(parts[1]) == 0 || parts[1] == "null" || parts[1] == "undefined" {
			// 如果只是分页获取首页留言，则不需要鉴权
			if strings.HasPrefix(ctx.Request.URL.Path, "/api/echo/page") {
				ctx.Set("userid", authModel.NO_USER_LOGINED)
				ctx.Next()
				return
			}

			// 查看留言详情也不需要鉴权
			//if strings.HasPrefix(ctx.Request.URL.Path, "/api/messages/") {
			//	ctx.Set("userid", uint(0))
			//	ctx.Next()
			//	return
			//}

			ctx.JSON(http.StatusOK, commonModel.Fail[any](errUtil.HandleError(&commonModel.ServerError{
				Msg: commonModel.TOKEN_NOT_FOUND,
				Err: nil,
			})))
			ctx.Abort()
			return
		}

		if !(len(parts) == 2 && parts[0] == "Bearer") {
			ctx.JSON(http.StatusOK, commonModel.Fail[any](errUtil.HandleError(&commonModel.ServerError{
				Msg: commonModel.TOKEN_NOT_VALID,
				Err: nil,
			})))
			ctx.Abort()
			return
		}

		mc, err := jwtUtil.ParseToken(parts[1])
		if err != nil {
			ctx.JSON(http.StatusOK, errUtil.HandleError(&commonModel.ServerError{
				Msg: commonModel.TOKEN_PARSE_ERROR,
				Err: err,
			}))
			ctx.Abort()
			return
		}

		ctx.Set("userid", mc.Userid)
		ctx.Next()
	}
}
