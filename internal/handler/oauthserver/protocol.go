// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
)

// ProtectedResource serves RFC 9728 metadata at both the root and the
// /mcp-suffixed well-known URI so clients using either discovery order work.
func (h *Handler) ProtectedResource() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metadata := h.service.ProtectedResourceMetadata(ctx.Request.Context(), requestOrigin(ctx))
		writeCacheableJSON(ctx, metadata)
	}
}

func (h *Handler) AuthorizationServerMetadata() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		metadata := h.service.AuthorizationServerMetadata(ctx.Request.Context(), requestOrigin(ctx))
		writeCacheableJSON(ctx, metadata)
	}
}

func (h *Handler) Register() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var metadata oauthserverModel.ClientMetadata
		if err := ctx.ShouldBindJSON(&metadata); err != nil {
			writeOAuthError(ctx, oauthserverModel.ErrInvalidClientMetadata("request body must be JSON"))
			return
		}
		registered, oauthErr := h.service.RegisterClient(ctx.Request.Context(), metadata)
		if oauthErr != nil {
			writeOAuthError(ctx, oauthErr)
			return
		}
		ctx.Header("Cache-Control", "no-store")
		ctx.JSON(http.StatusCreated, registered)
	}
}

func (h *Handler) Token() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := ctx.Request.ParseForm(); err != nil {
			writeOAuthError(ctx, oauthserverModel.ErrInvalidRequest("malformed form body"))
			return
		}
		token, oauthErr := h.service.Exchange(
			ctx.Request.Context(), requestOrigin(ctx), ctx.Request.PostForm)
		if oauthErr != nil {
			writeOAuthError(ctx, oauthErr)
			return
		}
		ctx.Header("Cache-Control", "no-store")
		ctx.Header("Pragma", "no-cache")
		ctx.JSON(http.StatusOK, token)
	}
}

func writeOAuthError(ctx *gin.Context, oauthErr *oauthserverModel.Error) {
	if oauthErr == nil {
		oauthErr = oauthserverModel.ErrInvalidRequest("request failed")
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(oauthErr.Status, gin.H{
		"error":             oauthErr.Code,
		"error_description": oauthErr.Description,
	})
}
