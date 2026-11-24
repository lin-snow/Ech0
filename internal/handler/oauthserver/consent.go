// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
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
	return res.Execute(func(ctx *gin.Context) res.Response {
		userID, err := requireSession(ctx)
		if err != nil {
			return res.Failed(err)
		}
		consent, oauthErr := h.service.Consent(
			ctx.Request.Context(), requestOrigin(ctx), ctx.Request.URL.Query(), userID)
		if oauthErr != nil {
			return writeConsentError(ctx, oauthErr)
		}
		ctx.Header("Cache-Control", "no-store")
		return res.Payload(consent)
	})
}

func (h *Handler) Approve() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userID, err := requireSession(ctx)
		if err != nil {
			return res.Failed(err)
		}
		var request approveRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			return writeConsentError(ctx, oauthserverModel.ErrInvalidRequest("request body must be JSON"))
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
			return writeConsentError(ctx, oauthserverModel.ErrInvalidRequest("decision must be approve or deny"))
		}
		if oauthErr != nil {
			return writeConsentError(ctx, oauthErr)
		}
		ctx.Header("Cache-Control", "no-store")
		return res.Payload(gin.H{"redirect_url": redirect})
	})
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

func requireSession(ctx *gin.Context) (string, error) {
	current := viewer.MustFromContext(ctx.Request.Context())
	if current.UserID() == "" {
		return "", commonModel.Fail(http.StatusUnauthorized, "authentication required")
	}
	// Consent must never be grantable with an API token: that would let a
	// token mint another, broader token.
	if current.Kind() != viewer.KindSession {
		return "", commonModel.Fail(http.StatusForbidden, "a browser session is required")
	}
	return current.UserID(), nil
}

// writeConsentError renders the OAuth protocol error as an RFC 9457 problem
// document. The status comes from the protocol error; its description is the
// developer-facing detail.
func writeConsentError(ctx *gin.Context, oauthErr *oauthserverModel.Error) res.Response {
	if oauthErr == nil {
		oauthErr = oauthserverModel.ErrInvalidRequest("request failed")
	}
	ctx.Header("Cache-Control", "no-store")
	return res.Failed(commonModel.Fail(oauthErr.Status, oauthErr.Description))
}
