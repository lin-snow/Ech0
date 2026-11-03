// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authzModel "github.com/lin-snow/ech0/internal/model/authz"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authzService "github.com/lin-snow/ech0/internal/service/authz"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type Handler struct {
	service authzService.Service
}

func NewAuthzHandler(service authzService.Service) *Handler {
	return &Handler{service: service}
}

type approveRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	ResponseType        string `json:"response_type"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	Resource            string `json:"resource"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	Decision            string `json:"decision"`
}

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
		var metadata authzModel.ClientMetadata
		if err := ctx.ShouldBindJSON(&metadata); err != nil {
			writeOAuthError(ctx, authzModel.ErrInvalidClientMetadata("request body must be JSON"))
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
			writeOAuthError(ctx, authzModel.ErrInvalidRequest("malformed form body"))
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

func (h *Handler) Consent() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := requireSession(ctx)
		if !ok {
			return
		}
		consent, oauthErr := h.service.Consent(
			ctx.Request.Context(), requestOrigin(ctx), ctx.Request.URL.Query(), userID)
		if oauthErr != nil {
			writeConsentError(ctx, oauthErr)
			return
		}
		ctx.Header("Cache-Control", "no-store")
		ctx.JSON(http.StatusOK, commonModel.OK(consent))
	}
}

func (h *Handler) Approve() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := requireSession(ctx)
		if !ok {
			return
		}
		var request approveRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			writeOAuthError(ctx, authzModel.ErrInvalidRequest("request body must be JSON"))
			return
		}
		query := approveQuery(request)

		var (
			redirect string
			oauthErr *authzModel.Error
		)
		if request.Decision == "deny" {
			redirect, oauthErr = h.service.Deny(ctx.Request.Context(), requestOrigin(ctx), query)
		} else {
			redirect, oauthErr = h.service.Approve(ctx.Request.Context(), requestOrigin(ctx), query, userID)
		}
		if oauthErr != nil {
			writeConsentError(ctx, oauthErr)
			return
		}
		ctx.Header("Cache-Control", "no-store")
		ctx.JSON(http.StatusOK, commonModel.OK(gin.H{"redirect_url": redirect}))
	}
}

func approveQuery(request approveRequest) url.Values {
	return url.Values{
		"client_id":             {request.ClientID},
		"redirect_uri":          {request.RedirectURI},
		"response_type":         {request.ResponseType},
		"scope":                 {request.Scope},
		"state":                 {request.State},
		"resource":              {request.Resource},
		"code_challenge":        {request.CodeChallenge},
		"code_challenge_method": {request.CodeChallengeMethod},
	}
}

func requireSession(ctx *gin.Context) (string, bool) {
	current := viewer.MustFromContext(ctx.Request.Context())
	if current.UserID() == "" {
		ctx.JSON(http.StatusUnauthorized, commonModel.Fail[any]("authentication required"))
		ctx.Abort()
		return "", false
	}
	// Consent must never be grantable with an API token: that would let a
	// token mint another, broader token.
	if current.TokenType() != authModel.TokenTypeSession {
		ctx.JSON(http.StatusForbidden, commonModel.Fail[any]("a browser session is required"))
		ctx.Abort()
		return "", false
	}
	return current.UserID(), true
}

func writeCacheableJSON(ctx *gin.Context, payload any) {
	ctx.Header("Cache-Control", "public, max-age=3600")
	ctx.JSON(http.StatusOK, payload)
}

// writeConsentError keeps the SPA on Ech0's result envelope while preserving
// the OAuth error code for diagnostics.
func writeConsentError(ctx *gin.Context, oauthErr *authzModel.Error) {
	if oauthErr == nil {
		oauthErr = authzModel.ErrInvalidRequest("request failed")
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(oauthErr.Status, commonModel.FailWithLocalized[any](
		oauthErr.Description, oauthErr.Code, "", nil))
}

func writeOAuthError(ctx *gin.Context, oauthErr *authzModel.Error) {
	if oauthErr == nil {
		oauthErr = authzModel.ErrInvalidRequest("request failed")
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(oauthErr.Status, gin.H{
		"error":             oauthErr.Code,
		"error_description": oauthErr.Description,
	})
}

// requestOrigin derives the externally visible origin. Both well-known
// documents and OAuth responses use it when no canonical server_url is set.
func requestOrigin(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil || ctx.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + ctx.Request.Host
}
