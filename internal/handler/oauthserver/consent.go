// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	"github.com/lin-snow/ech0/pkg/viewer"
)

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
			writeConsentError(ctx, oauthserverModel.ErrInvalidRequest("request body must be JSON"))
			return
		}
		query := approveQuery(request)

		var (
			redirect string
			oauthErr *oauthserverModel.Error
		)
		switch request.Decision {
		case "approve":
			redirect, oauthErr = h.service.Approve(ctx.Request.Context(), requestOrigin(ctx), query, userID)
		case "deny":
			redirect, oauthErr = h.service.Deny(ctx.Request.Context(), requestOrigin(ctx), query)
		default:
			writeConsentError(ctx, oauthserverModel.ErrInvalidRequest("decision must be approve or deny"))
			return
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

// writeConsentError keeps the SPA on Ech0's result envelope while preserving
// the OAuth error code for diagnostics.
func writeConsentError(ctx *gin.Context, oauthErr *oauthserverModel.Error) {
	if oauthErr == nil {
		oauthErr = oauthserverModel.ErrInvalidRequest("request failed")
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(oauthErr.Status, commonModel.FailWithLocalized[any](
		oauthErr.Description, oauthErr.Code, "", nil))
}
