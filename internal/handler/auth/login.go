package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
)

// Login 用户登录
//
//	@Summary		用户登录
//	@Description	使用用户名和密码登录，返回 access_token 并通过 Set-Cookie 下发 refresh_token
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.LoginDto							true	"登录凭证"
//	@Success		200		{object}	handler.Response{data=model.TokenPair}	"登录成功"
//	@Failure		200		{object}	handler.Response						"登录失败"
//	@Router			/login [post]
func (h *AuthHandler) Login() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var loginDto authModel.LoginDto
		if err := ctx.ShouldBindJSON(&loginDto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		tokenPair, err := h.authService.Login(&loginDto)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		cookieUtil.SetRefreshTokenCookie(ctx, tokenPair.RefreshToken, config.Config().Auth.Jwt.RefreshExpires)
		return res.Response{
			Data: tokenPair,
			Msg:  commonModel.LOGIN_SUCCESS,
		}
	})
}
