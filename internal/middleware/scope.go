package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	errUtil "github.com/lin-snow/ech0/internal/util/err"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func RequireScopes(scopes ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		v := viewer.MustFromContext(ctx.Request.Context())
		if v.TokenType() == authModel.TokenTypeSession {
			ctx.Next()
			return
		}
		if v.TokenType() != authModel.TokenTypeAccess {
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
		if !containsValidAudience(v.Audience()) {
			ctx.JSON(
				http.StatusForbidden,
				commonModel.FailWithLocalized[any](
					i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), commonModel.MsgKeyAuthAudienceForbidden, errUtil.HandleError(&commonModel.ServerError{
						Msg: commonModel.NO_PERMISSION_DENIED,
						Err: nil,
					}), nil),
					commonModel.ErrCodeAudienceForbidden,
					commonModel.MsgKeyAuthAudienceForbidden,
					nil,
				),
			)
			ctx.Abort()
			return
		}
		if !containsAllScopes(v.Scopes(), scopes) {
			ctx.JSON(
				http.StatusForbidden,
				commonModel.FailWithLocalized[any](
					i18nUtil.Localize(i18nUtil.LocalizerFromGin(ctx), commonModel.MsgKeyAuthScopeForbidden, errUtil.HandleError(&commonModel.ServerError{
						Msg: commonModel.NO_PERMISSION_DENIED,
						Err: nil,
					}), nil),
					commonModel.ErrCodeScopeForbidden,
					commonModel.MsgKeyAuthScopeForbidden,
					nil,
				),
			)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func containsValidAudience(audiences []string) bool {
	if len(audiences) == 0 {
		return false
	}
	for _, audience := range audiences {
		if authModel.IsValidAudience(audience) {
			return true
		}
	}
	return false
}

func containsAllScopes(actual, required []string) bool {
	if len(required) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(actual))
	for _, scope := range actual {
		set[scope] = struct{}{}
	}
	for _, requiredScope := range required {
		if _, ok := set[requiredScope]; !ok {
			return false
		}
	}
	return true
}
