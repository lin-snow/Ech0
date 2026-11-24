// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
)

type (
	ListPasskeysInput struct{}
	PasskeyIDInput    struct {
		ID string `path:"id" format:"uuid" doc:"Passkey 设备 ID（UUID）"`
	}
	UpdatePasskeyNameInput struct {
		ID   string `path:"id" format:"uuid" doc:"Passkey 设备 ID（UUID）"`
		Body authModel.PasskeyUpdateDeviceNameReq
	}
)

type (
	PasskeyListOutput = struct{ Body []authModel.PasskeyDeviceDto }
	EmptyOutput       = struct{}
)

func (h *AuthHandler) getPasskeyOriginAndRPID(ctx *gin.Context) (origin string, rpID string) {
	rpID, origins := h.authService.PasskeyBoundary(ctx.Request.Context())
	if rpID != "" && len(origins) > 0 {
		return strings.TrimSpace(origins[0]), rpID
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

func (h *AuthHandler) PasskeyLoginBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Failed(errInvalidParams)
		}
		data, err := h.authService.PasskeyLoginBegin(rpID, origin)
		if err != nil {
			return res.Failed(err)
		}
		return res.Payload(data)
	})
}

func (h *AuthHandler) PasskeyLoginFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Failed(errInvalidBody)
		}
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Failed(errInvalidParams)
		}
		tokenPair, err := h.authService.PasskeyLoginFinish(rpID, origin, req.Nonce, req.Credential)
		if err != nil {
			return res.Failed(err)
		}
		cookieUtil.SetRefreshTokenCookie(ctx, tokenPair.RefreshToken, config.Config().Auth.Jwt.RefreshExpires)
		return res.Payload(tokenPair)
	})
}

func (h *AuthHandler) PasskeyRegisterBeginV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyRegisterBeginReq
		_ = ctx.ShouldBindJSON(&req)
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Failed(errInvalidParams)
		}
		data, err := h.authService.PasskeyRegisterBegin(ctx.Request.Context(), rpID, origin, req.DeviceName)
		if err != nil {
			return res.Failed(err)
		}
		return res.Payload(data)
	})
}

func (h *AuthHandler) PasskeyRegisterFinishV2() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Failed(errInvalidBody)
		}
		origin, rpID := h.getPasskeyOriginAndRPID(ctx)
		if origin == "" || rpID == "" {
			return res.Failed(errInvalidParams)
		}
		if err := h.authService.PasskeyRegisterFinish(ctx.Request.Context(), rpID, origin, req.Nonce, req.Credential); err != nil {
			return res.Failed(err)
		}
		return res.NoContent(http.StatusOK)
	})
}

func (h *AuthHandler) ListPasskeys(ctx context.Context, _ *ListPasskeysInput) (*PasskeyListOutput, error) {
	devs, err := h.authService.ListPasskeys(ctx)
	if err != nil {
		return nil, err
	}
	return &PasskeyListOutput{Body: devs}, nil
}

func (h *AuthHandler) DeletePasskey(ctx context.Context, in *PasskeyIDInput) (*EmptyOutput, error) {
	if err := h.authService.DeletePasskey(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *AuthHandler) UpdatePasskeyDeviceName(ctx context.Context, in *UpdatePasskeyNameInput) (*EmptyOutput, error) {
	if err := h.authService.UpdatePasskeyDeviceName(ctx, in.ID, in.Body.DeviceName); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}
