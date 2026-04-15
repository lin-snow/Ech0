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

// OAuthLogin 发起 OAuth2 登录重定向
//
//	@Summary		OAuth2 登录
//	@Description	根据 provider 发起 OAuth2 授权重定向，支持 github / google / qq / custom
//	@Tags			认证 - OAuth2
//	@Param			provider		path	string	true	"OAuth2 提供商 (github/google/qq/custom)"
//	@Param			redirect_uri	query	string	false	"登录成功后的前端回调地址"
//	@Success		302				"重定向到 OAuth2 提供商授权页"
//	@Failure		200				{object}	handler.Response	"参数无效"
//	@Router			/oauth/{provider}/login [get]
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

// OAuthCallback 处理 OAuth2 回调
//
//	@Summary		OAuth2 回调
//	@Description	接收 OAuth2 提供商回调，验证 code 和 state，生成一次性 exchange code 并重定向到前端
//	@Tags			认证 - OAuth2
//	@Param			provider	path	string	true	"OAuth2 提供商 (github/google/qq/custom)"
//	@Param			code		query	string	true	"OAuth2 授权码"
//	@Param			state		query	string	true	"OAuth2 状态参数"
//	@Success		302			"重定向到前端 /auth?code=xxx"
//	@Failure		200			{object}	handler.Response	"参数无效或回调处理失败"
//	@Router			/oauth/{provider}/callback [get]
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

// OAuthBind 绑定 OAuth2 账号到当前用户
//
//	@Summary		绑定 OAuth2 账号
//	@Description	为当前已认证用户绑定指定 OAuth2 提供商的账号，返回授权重定向 URL
//	@Tags			认证 - OAuth2
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string						true	"OAuth2 提供商 (github/google/qq/custom)"
//	@Param			body		body		object{redirect_uri=string}	true	"回调地址"
//	@Success		200			{object}	handler.Response			"绑定 URL"
//	@Failure		200			{object}	handler.Response			"参数无效或绑定失败"
//	@Router			/oauth/{provider}/bind [post]
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

// GetOAuthInfo 获取当前用户的 OAuth2 绑定信息
//
//	@Summary		获取 OAuth2 绑定信息
//	@Description	获取当前已认证用户指定 provider 的 OAuth2 绑定状态
//	@Tags			认证 - OAuth2
//	@Produce		json
//	@Param			provider	query		string				false	"OAuth2 提供商，默认 github"
//	@Success		200			{object}	handler.Response	"OAuth2 绑定信息"
//	@Router			/oauth/info [get]
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
