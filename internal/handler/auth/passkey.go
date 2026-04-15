package handler

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lin-snow/ech0/internal/config"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
)

func getPasskeyOriginAndRPID(ctx *gin.Context) (origin string, rpID string) {
	cfg := config.Config().Auth.WebAuthn
	if strings.TrimSpace(cfg.RPID) != "" && len(cfg.Origins) > 0 {
		return strings.TrimSpace(cfg.Origins[0]), strings.TrimSpace(cfg.RPID)
	}
	origin = strings.TrimSpace(ctx.GetHeader("Origin"))
	if origin == "" {
		return "", ""
	}
	if u, err := url.Parse(origin); err == nil && u.Hostname() != "" {
		return origin, u.Hostname()
	}
	return "", ""
}

func getOriginAndRPID(ctx *gin.Context) (origin string, rpID string) {
	if configured := strings.TrimSpace(config.Config().Setting.Serverurl); configured != "" {
		if u, err := url.Parse(configured); err == nil && u.Scheme != "" && u.Hostname() != "" {
			return u.Scheme + "://" + u.Host, u.Hostname()
		}
	}

	origin = strings.TrimSpace(ctx.GetHeader("Origin"))
	if origin == "" {
		ref := strings.TrimSpace(ctx.GetHeader("Referer"))
		if ref != "" {
			if u, err := url.Parse(ref); err == nil && u.Scheme != "" && u.Host != "" {
				origin = u.Scheme + "://" + u.Host
			}
		}
		if origin == "" {
			scheme := "http"
			if ctx.Request.TLS != nil {
				scheme = "https"
			}
			origin = scheme + "://" + ctx.Request.Host
		}
	}

	if u, err := url.Parse(origin); err == nil {
		if h := u.Hostname(); h != "" {
			rpID = h
		}
	}
	if rpID == "" {
		host := ctx.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		rpID = host
	}
	return origin, rpID
}

// PasskeyLoginBeginV2 发起 Passkey 登录挑战
//
//	@Summary		Passkey 登录 - 开始
//	@Description	生成 WebAuthn 登录挑战（discoverable credential），返回 publicKey 选项供前端调用 navigator.credentials.get
//	@Tags			认证 - Passkey
//	@Produce		json
//	@Success		200	{object}	handler.Response	"挑战生成成功，包含 nonce 和 publicKey 选项"
//	@Failure		200	{object}	handler.Response	"参数无效"
//	@Router			/passkey/login/begin [post]
func (h *AuthHandler) PasskeyLoginBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		data, err := h.authService.PasskeyLoginBegin(rpID, origin)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

// PasskeyLoginFinishV2 完成 Passkey 登录验证
//
//	@Summary		Passkey 登录 - 完成
//	@Description	验证 WebAuthn 登录断言，成功后返回 access_token 并通过 Set-Cookie 下发 refresh_token
//	@Tags			认证 - Passkey
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.PasskeyFinishReq					true	"包含 nonce 和 credential 的请求体"
//	@Success		200		{object}	handler.Response{data=model.TokenPair}	"登录成功，返回 access_token"
//	@Failure		200		{object}	handler.Response						"验证失败"
//	@Router			/passkey/login/finish [post]
func (h *AuthHandler) PasskeyLoginFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		tokenPair, err := h.authService.PasskeyLoginFinish(rpID, origin, req.Nonce, req.Credential)
		if err != nil {
			return res.Response{Err: err}
		}
		cookieUtil.SetRefreshTokenCookie(ctx, tokenPair.RefreshToken, config.Config().Auth.Jwt.RefreshExpires)
		return res.Response{Data: tokenPair}
	})
}

// PasskeyRegisterBeginV2 发起 Passkey 注册挑战
//
//	@Summary		Passkey 注册 - 开始
//	@Description	为当前已认证用户生成 WebAuthn 注册挑战，返回 publicKey 选项供前端调用 navigator.credentials.create
//	@Tags			认证 - Passkey
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.PasskeyRegisterBeginReq	true	"可选的设备名称"
//	@Success		200		{object}	handler.Response				"挑战生成成功，包含 nonce 和 publicKey 选项"
//	@Failure		200		{object}	handler.Response				"参数无效"
//	@Router			/passkey/register/begin [post]
func (h *AuthHandler) PasskeyRegisterBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyRegisterBeginReq
		_ = ctx.ShouldBindJSON(&req)
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		data, err := h.authService.PasskeyRegisterBegin(ctx.Request.Context(), rpID, origin, req.DeviceName)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

// PasskeyRegisterFinishV2 完成 Passkey 注册
//
//	@Summary		Passkey 注册 - 完成
//	@Description	验证 WebAuthn 注册凭据并保存，完成 Passkey 绑定
//	@Tags			认证 - Passkey
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.PasskeyFinishReq	true	"包含 nonce 和 credential 的请求体"
//	@Success		200		{object}	handler.Response		"注册成功"
//	@Failure		200		{object}	handler.Response		"验证失败"
//	@Router			/passkey/register/finish [post]
func (h *AuthHandler) PasskeyRegisterFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		if err := h.authService.PasskeyRegisterFinish(ctx.Request.Context(), rpID, origin, req.Nonce, req.Credential); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}

func (h *AuthHandler) PasskeyLoginBegin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		origin, rpID := getOriginAndRPID(ctx)
		data, err := h.authService.PasskeyLoginBegin(rpID, origin)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (h *AuthHandler) PasskeyLoginFinish() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := getOriginAndRPID(ctx)
		token, err := h.authService.PasskeyLoginFinish(rpID, origin, req.Nonce, req.Credential)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: token}
	})
}

func (h *AuthHandler) PasskeyRegisterBegin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyRegisterBeginReq
		_ = ctx.ShouldBindJSON(&req)

		origin, rpID := getOriginAndRPID(ctx)
		data, err := h.authService.PasskeyRegisterBegin(
			ctx.Request.Context(),
			rpID,
			origin,
			req.DeviceName,
		)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

func (h *AuthHandler) PasskeyRegisterFinish() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		origin, rpID := getOriginAndRPID(ctx)
		if err := h.authService.PasskeyRegisterFinish(ctx.Request.Context(), rpID, origin, req.Nonce, req.Credential); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}

// ListPasskeys 列出当前用户的所有 Passkey 设备
//
//	@Summary		列出 Passkey 设备
//	@Description	获取当前已认证用户绑定的所有 Passkey 设备列表
//	@Tags			认证 - Passkey
//	@Produce		json
//	@Success		200	{object}	handler.Response{data=[]model.PasskeyDeviceDto}	"获取成功"
//	@Failure		200	{object}	handler.Response								"获取失败"
//	@Router			/passkeys [get]
func (h *AuthHandler) ListPasskeys() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		devs, err := h.authService.ListPasskeys(ctx.Request.Context())
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: devs}
	})
}

// DeletePasskey 删除指定的 Passkey 设备
//
//	@Summary		删除 Passkey 设备
//	@Description	根据 ID 删除当前用户绑定的指定 Passkey 设备
//	@Tags			认证 - Passkey
//	@Produce		json
//	@Param			id	path		string				true	"Passkey 设备 ID (UUID)"
//	@Success		200	{object}	handler.Response	"删除成功"
//	@Failure		200	{object}	handler.Response	"删除失败"
//	@Router			/passkeys/{id} [delete]
func (h *AuthHandler) DeletePasskey() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		idStr := ctx.Param("id")
		if _, err := uuid.Parse(idStr); err != nil {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}

		if err := h.authService.DeletePasskey(ctx.Request.Context(), idStr); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}

// UpdatePasskeyDeviceName 更新 Passkey 设备名称
//
//	@Summary		更新 Passkey 设备名称
//	@Description	根据 ID 更新当前用户绑定的指定 Passkey 设备的显示名称
//	@Tags			认证 - Passkey
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Passkey 设备 ID (UUID)"
//	@Param			body	body		model.PasskeyUpdateDeviceNameReq	true	"新的设备名称"
//	@Success		200		{object}	handler.Response					"更新成功"
//	@Failure		200		{object}	handler.Response					"更新失败"
//	@Router			/passkeys/{id} [put]
func (h *AuthHandler) UpdatePasskeyDeviceName() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		idStr := ctx.Param("id")
		if _, err := uuid.Parse(idStr); err != nil {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}

		var req authModel.PasskeyUpdateDeviceNameReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		if err := h.authService.UpdatePasskeyDeviceName(ctx.Request.Context(), idStr, req.DeviceName); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}
