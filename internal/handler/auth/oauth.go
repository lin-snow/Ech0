// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type (
	OAuthBindBody struct {
		RedirectURI string `json:"redirect_uri" doc:"OAuth 回调地址"`
	}
	OAuthBindInput struct {
		Provider string `path:"provider" doc:"OAuth2 提供商 (github/google/qq/custom)"`
		Body     OAuthBindBody
	}
	GetOAuthInfoInput struct {
		Provider string `query:"provider" doc:"OAuth2 提供商，默认 github"`
	}
)

type (
	OAuthBindOutput = struct{ Body string }
	OAuthInfoOutput = struct{ Body userModel.OAuthInfoDto }
)

func normalizeOAuthProvider(provider string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case string(commonModel.OAuth2GITHUB):
		return string(commonModel.OAuth2GITHUB), true
	case string(commonModel.OAuth2GOOGLE):
		return string(commonModel.OAuth2GOOGLE), true
	case string(commonModel.OAuth2QQ):
		return string(commonModel.OAuth2QQ), true
	case string(commonModel.OAuth2CUSTOM):
		return string(commonModel.OAuth2CUSTOM), true
	default:
		return "", false
	}
}

func (h *AuthHandler) OAuthLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider, ok := normalizeOAuthProvider(ctx.Param("provider"))
		if !ok {
			return res.Failed(errInvalidParams)
		}

		redirectURI := ctx.Query("redirect_uri")
		redirectURL, err := h.authService.GetOAuthLoginURL(provider, redirectURI)
		if err != nil {
			return res.Failed(err)
		}
		ctx.Redirect(http.StatusFound, redirectURL)
		return res.NoContent(http.StatusFound)
	})
}

func (h *AuthHandler) OAuthCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider, ok := normalizeOAuthProvider(ctx.Param("provider"))
		if !ok {
			return res.Failed(errInvalidParams)
		}
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Failed(errInvalidParams)
		}

		redirectURL, err := h.authService.HandleOAuthCallback(provider, code, state)
		if err != nil || redirectURL == "" {
			if err == nil {
				err = errInvalidParams
			}
			return res.Failed(err)
		}
		ctx.Redirect(http.StatusFound, redirectURL)
		return res.NoContent(http.StatusFound)
	})
}

func (h *AuthHandler) OAuthBind(ctx context.Context, in *OAuthBindInput) (*OAuthBindOutput, error) {
	provider, ok := normalizeOAuthProvider(in.Provider)
	if !ok {
		return nil, errInvalidParams
	}
	bindURL, err := h.authService.BindOAuth(ctx, provider, in.Body.RedirectURI)
	if err != nil {
		return nil, err
	}
	return &OAuthBindOutput{Body: bindURL}, nil
}

func (h *AuthHandler) GetOAuthInfo(ctx context.Context, in *GetOAuthInfoInput) (*OAuthInfoOutput, error) {
	provider := in.Provider
	switch provider {
	case string(commonModel.OAuth2GITHUB),
		string(commonModel.OAuth2GOOGLE),
		string(commonModel.OAuth2QQ),
		string(commonModel.OAuth2CUSTOM):
	default:
		provider = string(commonModel.OAuth2GITHUB)
	}

	oauthInfo, _ := h.authService.GetOAuthInfo(ctx, provider)
	return &OAuthInfoOutput{Body: oauthInfo}, nil
}
