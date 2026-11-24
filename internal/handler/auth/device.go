// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

// deviceApprovalPath is the web route where the user approves the code.
const deviceApprovalPath = "/device"

func (h *AuthHandler) DeviceCode() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.DeviceCodeRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Failed(errInvalidBody)
		}

		code, err := h.authService.StartDeviceAuthorization(req.ClientName, req.Scopes)
		if err != nil {
			return res.Failed(err)
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
			return res.Failed(err)
		}
		return res.Payload(token)
	})
}

func (h *AuthHandler) DevicePending() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		info, err := h.authService.DeviceAuthorizationInfo(ctx.Query("user_code"))
		if err != nil {
			return res.Failed(err)
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
			return res.Failed(err)
		}
		return res.NoContent(http.StatusOK)
	})
}
