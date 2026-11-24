// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

// deviceApprovalPath is the web route where the user approves the code.
const deviceApprovalPath = "/device"

// deviceFailure tags a device-flow failure with the problem type a polling
// client branches on, so it can tell "keep waiting" from a terminal error.
// Every one of them is a 400: the device grant has no other status to express.
func deviceFailure(err error) res.Response {
	switch {
	case errors.Is(err, authService.ErrDevicePending):
		return res.Failed(commonModel.FailWithType(
			http.StatusBadRequest, commonModel.TypeDevicePending, commonModel.DEVICE_CODE_PENDING))
	case errors.Is(err, authService.ErrDeviceSlowDown):
		return res.Failed(commonModel.FailWithType(
			http.StatusBadRequest, commonModel.TypeDeviceSlowDown, commonModel.DEVICE_CODE_SLOW_DOWN))
	case errors.Is(err, authService.ErrDeviceExpired):
		return res.Failed(commonModel.FailWithType(
			http.StatusBadRequest, commonModel.TypeDeviceExpired, commonModel.DEVICE_CODE_EXPIRED))
	case errors.Is(err, authService.ErrDeviceDenied):
		return res.Failed(commonModel.FailWithType(
			http.StatusBadRequest, commonModel.TypeDeviceDenied, commonModel.DEVICE_CODE_DENIED))
	case errors.Is(err, authService.ErrDeviceUnknown):
		return res.Failed(commonModel.Fail(http.StatusBadRequest, commonModel.DEVICE_CODE_UNKNOWN))
	default:
		return res.Failed(err)
	}
}

func (h *AuthHandler) DeviceCode() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.DeviceCodeRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Failed(errInvalidBody)
		}

		code, err := h.authService.StartDeviceAuthorization(req.ClientName, req.Scopes)
		if err != nil {
			return deviceFailure(err)
		}
		code.VerificationURI = deviceApprovalPath
		code.VerificationURIComplete = deviceApprovalPath + "?code=" + url.QueryEscape(code.UserCode)
		return res.Payload(code)
	})
}

func (h *AuthHandler) DeviceToken() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.DeviceTokenRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Failed(errInvalidBody)
		}

		token, err := h.authService.PollDeviceAuthorization(req.DeviceCode)
		if err != nil {
			return deviceFailure(err)
		}
		return res.Payload(token)
	})
}

func (h *AuthHandler) DevicePending() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		info, err := h.authService.DeviceAuthorizationInfo(ctx.Query("user_code"))
		if err != nil {
			return deviceFailure(err)
		}
		return res.Payload(info)
	})
}

func (h *AuthHandler) DeviceApprove() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.DeviceApproveRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Failed(errInvalidBody)
		}

		if err := h.authService.ApproveDeviceAuthorization(ctx.Request.Context(), req.UserCode); err != nil {
			return deviceFailure(err)
		}
		return res.NoContent(http.StatusOK)
	})
}
