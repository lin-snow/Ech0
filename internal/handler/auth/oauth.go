// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
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
	OAuthBindOutput = commonModel.Result[string]
	OAuthInfoOutput = commonModel.Result[userModel.OAuthInfoDto]
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
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}

		redirectURI := ctx.Query("redirect_uri")
		redirectURL, err := h.authService.GetOAuthLoginURL(provider, redirectURI)
		if err != nil {
			return res.Response{Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) OAuthCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider, ok := normalizeOAuthProvider(ctx.Param("provider"))
		if !ok {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}

		redirectURL, err := h.authService.HandleOAuthCallback(provider, code, state)
		if err != nil || redirectURL == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) OAuthBind(ctx context.Context, in *OAuthBindInput) (OAuthBindOutput, error) {
	provider, ok := normalizeOAuthProvider(in.Provider)
	if !ok {
		return OAuthBindOutput{}, commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, commonModel.INVALID_PARAMS)
	}
	bindURL, err := h.authService.BindOAuth(ctx, provider, in.Body.RedirectURI)
	if err != nil {
		return OAuthBindOutput{}, err
	}
	return commonModel.OK(bindURL, commonModel.GET_OAUTH_BINGURL_SUCCESS), nil
}

func (h *AuthHandler) GetOAuthInfo(ctx context.Context, in *GetOAuthInfoInput) (OAuthInfoOutput, error) {
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
	return commonModel.OK(oauthInfo, commonModel.GET_OAUTH_INFO_SUCCESS), nil
}
