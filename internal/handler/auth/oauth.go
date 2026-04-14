package handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
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

func (h *AuthHandler) OAuthBind() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider, ok := normalizeOAuthProvider(ctx.Param("provider"))
		if !ok {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		type reqBody struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req reqBody
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		bindURL, err := h.authService.BindOAuth(ctx.Request.Context(), provider, req.RedirectURI)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: bindURL, Msg: commonModel.GET_OAUTH_BINGURL_SUCCESS}
	})
}

func (h *AuthHandler) GetOAuthInfo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		provider := ctx.Query("provider")
		switch provider {
		case string(commonModel.OAuth2GITHUB),
			string(commonModel.OAuth2GOOGLE),
			string(commonModel.OAuth2QQ),
			string(commonModel.OAuth2CUSTOM):
		default:
			provider = string(commonModel.OAuth2GITHUB)
		}

		oauthInfo, _ := h.authService.GetOAuthInfo(ctx.Request.Context(), provider)
		return res.Response{
			Data: oauthInfo,
			Msg:  commonModel.GET_OAUTH_INFO_SUCCESS,
		}
	})
}

func (h *AuthHandler) BindGitHub() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		type reqBody struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req reqBody
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		bindURL, err := h.authService.BindOAuth(ctx.Request.Context(), string(commonModel.OAuth2GITHUB), req.RedirectURI)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: bindURL, Msg: commonModel.GET_OAUTH_BINGURL_SUCCESS}
	})
}

func (h *AuthHandler) GitHubLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		redirectURI := ctx.Query("redirect_uri")
		redirectURL, err := h.authService.GetOAuthLoginURL(string(commonModel.OAuth2GITHUB), redirectURI)
		if err != nil {
			return res.Response{Msg: commonModel.FAILED_TO_GET_GITHUB_LOGIN_URL, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) GitHubCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		redirectURL, err := h.authService.HandleOAuthCallback(string(commonModel.OAuth2GITHUB), code, state)
		if err != nil || redirectURL == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) BindGoogle() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		type reqBody struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req reqBody
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		bindURL, err := h.authService.BindOAuth(ctx.Request.Context(), string(commonModel.OAuth2GOOGLE), req.RedirectURI)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: bindURL, Msg: commonModel.GET_OAUTH_BINGURL_SUCCESS}
	})
}

func (h *AuthHandler) GoogleLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		redirectURI := ctx.Query("redirect_uri")
		redirectURL, err := h.authService.GetOAuthLoginURL(string(commonModel.OAuth2GOOGLE), redirectURI)
		if err != nil {
			return res.Response{Msg: commonModel.FAILED_TO_GET_GOOGLE_LOGIN_URL, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) GoogleCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		redirectURL, err := h.authService.HandleOAuthCallback(string(commonModel.OAuth2GOOGLE), code, state)
		if err != nil || redirectURL == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) QQLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		redirectURI := ctx.Query("redirect_uri")
		redirectURL, err := h.authService.GetOAuthLoginURL(string(commonModel.OAuth2QQ), redirectURI)
		if err != nil {
			return res.Response{Msg: commonModel.FAILED_TO_GET_QQ_LOGIN_URL, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) QQCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		redirectURL, err := h.authService.HandleOAuthCallback(string(commonModel.OAuth2QQ), code, state)
		if err != nil || redirectURL == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) BindQQ() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		type reqBody struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req reqBody
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		bindURL, err := h.authService.BindOAuth(ctx.Request.Context(), string(commonModel.OAuth2QQ), req.RedirectURI)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: bindURL, Msg: commonModel.GET_OAUTH_BINGURL_SUCCESS}
	})
}

func (h *AuthHandler) CustomOAuthLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		redirectURI := ctx.Query("redirect_uri")
		redirectURL, err := h.authService.GetOAuthLoginURL(string(commonModel.OAuth2CUSTOM), redirectURI)
		if err != nil {
			return res.Response{Msg: commonModel.FAILED_TO_GET_CUSTOM_LOGIN_URL, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) CustomOAuthCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS}
		}
		redirectURL, err := h.authService.HandleOAuthCallback(string(commonModel.OAuth2CUSTOM), code, state)
		if err != nil || redirectURL == "" {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

func (h *AuthHandler) BindCustomOAuth() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		type reqBody struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req reqBody
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		bindURL, err := h.authService.BindOAuth(ctx.Request.Context(), string(commonModel.OAuth2CUSTOM), req.RedirectURI)
		if err != nil {
			return res.Response{Msg: "", Err: err}
		}
		return res.Response{Data: bindURL, Msg: commonModel.GET_OAUTH_BINGURL_SUCCESS}
	})
}
