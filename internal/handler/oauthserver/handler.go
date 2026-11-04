// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package oauthserver serves the OAuth 2.1 authorization server used by MCP
// clients. Protocol endpoints keep RFC wire formats; consent endpoints use
// Ech0's result envelope because they are consumed by the SPA.
package oauthserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	oauthserverService "github.com/lin-snow/ech0/internal/service/oauthserver"
	"github.com/lin-snow/ech0/internal/util/reqinfo"
)

type Handler struct {
	service oauthserverService.Service
}

func NewOAuthServerHandler(service oauthserverService.Service) *Handler {
	return &Handler{service: service}
}

func writeCacheableJSON(ctx *gin.Context, payload any) {
	// The document embeds the request origin, so a shared cache must key on
	// the headers that origin derives from.
	ctx.Header("Vary", "Host, X-Forwarded-Proto, X-Forwarded-Host")
	ctx.Header("Cache-Control", "public, max-age=3600")
	ctx.JSON(http.StatusOK, payload)
}

// requestOrigin derives the externally visible origin. Both well-known
// documents and OAuth responses use it when no canonical server_url is set.
func requestOrigin(ctx *gin.Context) string {
	return reqinfo.Origin(ctx.Request)
}
