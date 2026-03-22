package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// FullAccessOnly blocks integration-scoped tokens from accessing full-privilege routes.
func FullAccessOnly() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if v.TokenScope() == authModel.TokenScopeIntegration {
			ctx.AbortWithStatusJSON(
				http.StatusForbidden,
				commonModel.Fail[any](errUtil.HandleError(&commonModel.ServerError{
					Msg: commonModel.NO_PERMISSION_DENIED,
					Err: nil,
				})),
			)
			return
		}

		ctx.Next()
	}
}
