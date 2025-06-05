package middleware

import (
	"github.com/gin-gonic/gin"
)

// JWT Auth Middleware
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// auth := ctx.Request.Header.Get("Authorization")

		// parts := strings.SplitN(auth, " ", 2)

		// if auth == "" || len(parts) != 2 || len(parts[1]) == 0 || parts[1] == "null" || parts[1] == "undefined" {
		// 	// 如果只是分页获取首页留言，则不需要鉴权
		// 	if strings.HasPrefix(ctx.Request.URL.Path, "/api/messages/page") {
		// 		ctx.Set("userid", uint(0))
		// 		ctx.Next()
		// 		return
		// 	}

		// 	// 查看留言详情也不需要鉴权
		// 	if strings.HasPrefix(ctx.Request.URL.Path, "/api/messages/") {
		// 		ctx.Set("userid", uint(0))
		// 		ctx.Next()
		// 		return
		// 	}

		// 	ctx.JSON(http.StatusOK, dto.Fail[any](models.TokenNotFoundMessage))
		// 	ctx.Abort()
		// 	return
		// }

		// if !(len(parts) == 2 && parts[0] == "Bearer") {
		// 	ctx.JSON(http.StatusOK, dto.Fail[any](models.TokenInvalidMessage))
		// 	ctx.Abort()
		// 	return
		// }

		// mc, err := pkg.ParseToken(parts[1])
		// if err != nil {
		// 	ctx.JSON(http.StatusOK, dto.Fail[any](models.TokenInvalidMessage))
		// 	ctx.Abort()
		// 	return
		// }

		// ctx.Set("userid", mc.Userid)
		// ctx.Next()
	}
}
