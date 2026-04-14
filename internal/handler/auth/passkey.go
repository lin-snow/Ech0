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

func (h *AuthHandler) ListPasskeys() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		devs, err := h.authService.ListPasskeys(ctx.Request.Context())
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: devs}
	})
}

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
