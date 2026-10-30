// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"errors"
	"net/url"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// deviceApprovalPath is the web route where the user approves the code.
const deviceApprovalPath = "/device"

func deviceErrorCode(err error) string {
	switch {
	case errors.Is(err, authService.ErrDevicePending):
		return commonModel.ErrCodeDevicePending
	case errors.Is(err, authService.ErrDeviceSlowDown):
		return commonModel.ErrCodeDeviceSlowDown
	case errors.Is(err, authService.ErrDeviceExpired):
		return commonModel.ErrCodeDeviceExpired
	case errors.Is(err, authService.ErrDeviceDenied):
		return commonModel.ErrCodeDeviceDenied
	case errors.Is(err, authService.ErrDeviceUnknown):
		return commonModel.ErrCodeDeviceUnknown
	default:
		return ""
	}
}

func (h *AuthHandler) DeviceCode() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.DeviceCodeRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		code, err := h.authService.StartDeviceAuthorization(req.ClientName, req.Scopes)
		if err != nil {
			return res.Response{Err: err, ErrorCode: deviceErrorCode(err)}
		}
		code.VerificationURI = deviceApprovalPath
		code.VerificationURIComplete = deviceApprovalPath + "?code=" + url.QueryEscape(code.UserCode)
		return res.Response{Data: code}
	})
}

func (h *AuthHandler) DeviceToken() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.DeviceTokenRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		token, err := h.authService.PollDeviceAuthorization(req.DeviceCode)
		if err != nil {
			return res.Response{Err: err, ErrorCode: deviceErrorCode(err)}
		}
		return res.Response{Data: token}
	})
}

func (h *AuthHandler) DevicePending() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		info, err := h.authService.DeviceAuthorizationInfo(ctx.Query("user_code"))
		if err != nil {
			return res.Response{Err: err, ErrorCode: deviceErrorCode(err)}
		}
		return res.Response{Data: info}
	})
}

func (h *AuthHandler) DeviceApprove() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.DeviceApproveRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		if err := h.authService.ApproveDeviceAuthorization(ctx.Request.Context(), req.UserCode); err != nil {
			return res.Response{Err: err, ErrorCode: deviceErrorCode(err)}
		}
		return res.Response{}
	})
}
