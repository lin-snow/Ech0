package handler

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/user"
	service "github.com/lin-snow/ech0/internal/service/user"
)

type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler UserHandler 的构造函数
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Login 用户登录
//
//	@Summary		用户登录接口
//	@Description	用户通过用户名和密码登录，返回 JWT Token
//	@Tags			用户认证
//	@Accept			application/json
//	@Produce		application/json
//	@Param			login	body		authModel.LoginDto	true	"登录请求体"
//	@Success		200		{object}	res.Response		"登录成功，返回JWT Token"
//	@Failure		200		{object}	res.Response		"登录失败，返回错误信息"
//	@Router			/login [post]
func (userHandler *UserHandler) Login() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 从请求体获取用户名和密码
		var loginDto authModel.LoginDto
		if err := ctx.ShouldBindJSON(&loginDto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		// 调用 Service 层处理登陆
		token, err := userHandler.userService.Login(&loginDto)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		// 返回成功响应， 包含 JWT Token
		return res.Response{
			Data: token,
			Msg:  commonModel.LOGIN_SUCCESS,
		}
	})
}

// Register 用户注册
//
//	@Summary		用户注册
//	@Description	通过提交用户名、密码等信息完成注册
//	@Tags			用户认证
//	@Accept			json
//	@Produce		json
//	@Param			register	body		authModel.RegisterDto	true	"注册请求体"
//	@Success		200			{object}	res.Response			"注册成功，code=1，msg=REGISTER_SUCCESS"
//	@Failure		200			{object}	res.Response			"请求参数错误或注册失败，code=0，msg错误描述"
//	@Router			/register [post]
func (userHandler *UserHandler) Register() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var registerDto authModel.RegisterDto
		if err := ctx.ShouldBindJSON(&registerDto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		// 调用 Service 层处理注册
		if err := userHandler.userService.Register(&registerDto); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.REGISTER_SUCCESS,
		}
	})
}

// UpdateUser 更新用户信息
//
//	@Summary		更新当前用户的信息
//	@Description	接口会根据请求体更新用户相关字段，需携带有效的用户身份（如 JWT）
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			user	body		model.UserInfoDto	true	"用户更新信息"
//	@Success		200		{object}	res.Response		"更新成功，code=1，msg=UPDATE_USER_SUCCESS"
//	@Failure		200		{object}	res.Response		"请求参数错误或更新失败，code=0，msg错误描述"
//	@Security		ApiKeyAuth
//	@Router			/user [put]
func (userHandler *UserHandler) UpdateUser() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 解析用户请求体中的参数
		var userdto model.UserInfoDto
		if err := ctx.ShouldBindJSON(&userdto); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)
		if err := userHandler.userService.UpdateUser(userid, userdto); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_USER_SUCCESS,
		}
	})
}

// UpdateUserAdmin 更新用户权限
//
//	@Summary		更新用户权限（管理员权限）
//	@Description	通过用户ID更新其管理员权限，接口调用者需拥有相应权限
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"用户ID"
//	@Success		200	{object}	res.Response	"更新成功，code=1，msg=UPDATE_USER_SUCCESS"
//	@Failure		200	{object}	res.Response	"参数错误或更新失败，code=0，msg错误描述"
//	@Security		ApiKeyAuth
//	@Router			/user/admin/{id} [put]
func (userHandler *UserHandler) UpdateUserAdmin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
				Err: err,
			}
		}

		if err := userHandler.userService.UpdateUserAdmin(userid, uint(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.UPDATE_USER_SUCCESS,
		}
	})
}

// GetAllUsers 获取所有用户
//
//	@Summary		获取所有用户
//	@Description	获取系统中所有用户的详细信息，接口需要认证
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=[]model.UserInfoDto}	"获取成功，code=1，包含用户列表"
//	@Failure		200	{object}	res.Response							"获取失败，code=0，msg错误描述"
//	@Security		ApiKeyAuth
//	@Router			/allusers [get]
func (userHandler *UserHandler) GetAllUsers() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		allusers, err := userHandler.userService.GetAllUsers()
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: allusers,
			Msg:  commonModel.GET_USER_SUCCESS,
		}
	})
}

// DeleteUser 删除用户
//
//	@Summary		删除用户
//	@Description	根据用户ID删除用户，调用者需具备相应权限
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"用户ID"
//	@Success		200	{object}	res.Response	"删除成功，code=1，msg=DELETE_USER_SUCCESS"
//	@Failure		200	{object}	res.Response	"参数错误或删除失败，code=0，msg错误描述"
//	@Security		ApiKeyAuth
//	@Router			/user/{id} [delete]
func (userHandler *UserHandler) DeleteUser() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
				Err: err,
			}
		}

		if err := userHandler.userService.DeleteUser(userid, uint(id)); err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Msg: commonModel.DELETE_USER_SUCCESS,
		}
	})
}

// GetUserInfo 获取当前用户信息
//
//	@Summary		获取当前用户信息
//	@Description	获取当前认证用户的详细信息，密码字段不会返回
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	res.Response{data=model.UserInfoDto}	"获取成功，code=1，包含用户信息"
//	@Failure		200	{object}	res.Response							"获取失败，code=0，msg错误描述"
//	@Security		ApiKeyAuth
//	@Router			/user [get]
func (userHandler *UserHandler) GetUserInfo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 调用 Service 层获取用户信息
		user, err := userHandler.userService.GetUserByID(int(userid))
		user.Password = "" // 不返回密码
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		// 返回成功响应
		return res.Response{
			Data: user,
			Msg:  commonModel.GET_USER_INFO_SUCCESS,
		}
	})
}

// BindGitHub 绑定 GitHub 账号
func (userHandler *UserHandler) BindGitHub() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		type Req struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req Req
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		bindURL, err := userHandler.userService.BindOAuth(
			userid,
			string(commonModel.OAuth2GITHUB),
			req.RedirectURI,
		)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: bindURL,
			Msg:  commonModel.GET_OAUTH_BINGURL_SUCCESS,
		}
	})
}

// GitHubLogin 处理 GitHub OAuth2 登录请求
func (userHandler *UserHandler) GitHubLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取重定向 URL
		redirect_URI := ctx.Query("redirect_uri")

		redirectURL, err := userHandler.userService.GetOAuthLoginURL(
			string(commonModel.OAuth2GITHUB),
			redirect_URI,
		)
		if err != nil {
			return res.Response{
				Msg: commonModel.FAILED_TO_GET_GITHUB_LOGIN_URL,
				Err: err,
			}
		}

		// 重定向到 GitHub 登录页面
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// GitHubCallback 处理 GitHub OAuth2 回调
func (userHandler *UserHandler) GitHubCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
				Err: nil,
			}
		}

		redirectURL := userHandler.userService.HandleOAuthCallback(
			string(commonModel.OAuth2GITHUB),
			code,
			state,
		)
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// BindGoogle 绑定 Google 账号
func (userHandler *UserHandler) BindGoogle() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		type Req struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req Req
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		bindURL, err := userHandler.userService.BindOAuth(
			userid,
			string(commonModel.OAuth2GOOGLE),
			req.RedirectURI,
		)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: bindURL,
			Msg:  commonModel.GET_OAUTH_BINGURL_SUCCESS,
		}
	})
}

// GoogleLogin 处理 Google OAuth2 登录请求
func (userHandler *UserHandler) GoogleLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取重定向 URL
		redirect_URI := ctx.Query("redirect_uri")

		redirectURL, err := userHandler.userService.GetOAuthLoginURL(
			string(commonModel.OAuth2GOOGLE),
			redirect_URI,
		)
		if err != nil {
			return res.Response{
				Msg: commonModel.FAILED_TO_GET_GOOGLE_LOGIN_URL,
				Err: err,
			}
		}

		// 重定向到 Google 登录页面
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// GoogleCallback 处理 Google OAuth2 回调
func (userHandler *UserHandler) GoogleCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
				Err: nil,
			}
		}

		redirectURL := userHandler.userService.HandleOAuthCallback(
			string(commonModel.OAuth2GOOGLE),
			code,
			state,
		)
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// QQLogin 处理 QQ OAuth2 登录请求
func (userHandler *UserHandler) QQLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取重定向 URL
		redirect_URI := ctx.Query("redirect_uri")

		redirectURL, err := userHandler.userService.GetOAuthLoginURL(
			string(commonModel.OAuth2QQ),
			redirect_URI,
		)
		if err != nil {
			return res.Response{
				Msg: commonModel.FAILED_TO_GET_QQ_LOGIN_URL,
				Err: err,
			}
		}

		// 重定向到 QQ 登录页面
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// QQCallback 处理 QQ OAuth2 回调
func (userHandler *UserHandler) QQCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
				Err: nil,
			}
		}

		redirectURL := userHandler.userService.HandleOAuthCallback(
			string(commonModel.OAuth2QQ),
			code,
			state,
		)
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// BindQQ 绑定 QQ 账号
func (userHandler *UserHandler) BindQQ() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		type Req struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req Req
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		bindURL, err := userHandler.userService.BindOAuth(
			userid,
			string(commonModel.OAuth2QQ),
			req.RedirectURI,
		)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: bindURL,
			Msg:  commonModel.GET_OAUTH_BINGURL_SUCCESS,
		}
	})
}

// CustomOAuthLogin 处理自定义 OAuth2 登录请求
func (userHandler *UserHandler) CustomOAuthLogin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取重定向 URL
		redirect_URI := ctx.Query("redirect_uri")

		redirectURL, err := userHandler.userService.GetOAuthLoginURL(
			string(commonModel.OAuth2CUSTOM),
			redirect_URI,
		)
		if err != nil {
			return res.Response{
				Msg: commonModel.FAILED_TO_GET_CUSTOM_LOGIN_URL,
				Err: err,
			}
		}

		// 重定向到自定义 OAuth2 登录页面
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// CustomOAuthCallback 处理自定义 OAuth2 回调
func (userHandler *UserHandler) CustomOAuthCallback() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			return res.Response{
				Msg: commonModel.INVALID_PARAMS,
				Err: nil,
			}
		}

		redirectURL := userHandler.userService.HandleOAuthCallback(
			string(commonModel.OAuth2CUSTOM),
			code,
			state,
		)
		ctx.Redirect(302, redirectURL)
		return res.Response{}
	})
}

// BindCustomOAuth 绑定自定义 OAuth2 账号
func (userHandler *UserHandler) BindCustomOAuth() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		type Req struct {
			RedirectURI string `json:"redirect_uri"`
		}
		var req Req
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{
				Msg: commonModel.INVALID_REQUEST_BODY,
				Err: err,
			}
		}

		bindURL, err := userHandler.userService.BindOAuth(
			userid,
			string(commonModel.OAuth2CUSTOM),
			req.RedirectURI,
		)
		if err != nil {
			return res.Response{
				Msg: "",
				Err: err,
			}
		}

		return res.Response{
			Data: bindURL,
			Msg:  commonModel.GET_OAUTH_BINGURL_SUCCESS,
		}
	})
}

// GetOAuthInfo 获取 OAuth2 配置信息
func (userHandler *UserHandler) GetOAuthInfo() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		// 获取当前用户 ID
		userid := ctx.MustGet("userid").(uint)

		// 获取 provider 参数
		provider := ctx.Query("provider")
		switch provider {
		case string(commonModel.OAuth2GITHUB),
			string(commonModel.OAuth2GOOGLE),
			string(commonModel.OAuth2QQ),
			string(commonModel.OAuth2CUSTOM):
			// 保持原值
		default:
			provider = string(commonModel.OAuth2GITHUB) // 默认使用 GitHub
		}

		// 调用 Service 层获取 OAuth2 信息
		oauthInfo, _ := userHandler.userService.GetOAuthInfo(userid, provider)

		return res.Response{
			Data: oauthInfo,
			Msg:  commonModel.GET_OAUTH_INFO_SUCCESS,
		}
	})
}

func getOriginAndRPID(ctx *gin.Context) (origin string, rpID string) {
	origin = strings.TrimSpace(ctx.GetHeader("Origin"))
	if origin == "" {
		// fallback：尽量从 Referer 推导；再不行用 scheme+Host
		ref := strings.TrimSpace(ctx.GetHeader("Referer"))
		if ref != "" {
			if u, err := url.Parse(ref); err == nil && u.Scheme != "" && u.Host != "" {
				origin = u.Scheme + "://" + u.Host
			}
		}
		if origin == "" {
			scheme := "http"
			if ctx.Request.TLS != nil {
				scheme = "https"
			}
			origin = scheme + "://" + ctx.Request.Host
		}
	}

	if u, err := url.Parse(origin); err == nil {
		if h := u.Hostname(); h != "" {
			rpID = h
		}
	}
	if rpID == "" {
		host := ctx.Request.Host
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}
		rpID = host
	}
	return origin, rpID
}

// PasskeyLoginBegin 开始 Passkey 登录（discoverable / resident key）
func (userHandler *UserHandler) PasskeyLoginBegin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		origin, rpID := getOriginAndRPID(ctx)

		data, err := userHandler.userService.PasskeyLoginBegin(rpID, origin)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

// PasskeyLoginFinish 完成 Passkey 登录
func (userHandler *UserHandler) PasskeyLoginFinish() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}
		origin, rpID := getOriginAndRPID(ctx)

		token, err := userHandler.userService.PasskeyLoginFinish(
			rpID,
			origin,
			req.Nonce,
			req.Credential,
		)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: token}
	})
}

// PasskeyRegisterBegin 开始 Passkey 绑定（仅已登录用户）
func (userHandler *UserHandler) PasskeyRegisterBegin() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)

		var req authModel.PasskeyRegisterBeginReq
		_ = ctx.ShouldBindJSON(&req) // device_name 可选

		origin, rpID := getOriginAndRPID(ctx)
		data, err := userHandler.userService.PasskeyRegisterBegin(
			userid,
			rpID,
			origin,
			req.DeviceName,
		)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: data}
	})
}

// PasskeyRegisterFinish 完成 Passkey 绑定
func (userHandler *UserHandler) PasskeyRegisterFinish() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)

		var req authModel.PasskeyFinishReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		origin, rpID := getOriginAndRPID(ctx)
		if err := userHandler.userService.PasskeyRegisterFinish(userid, rpID, origin, req.Nonce, req.Credential); err != nil {
			return res.Response{Err: err}
		}

		return res.Response{}
	})
}

// ListPasskeys 获取当前用户已绑定的 Passkey 设备列表
func (userHandler *UserHandler) ListPasskeys() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)
		devs, err := userHandler.userService.ListPasskeys(userid)
		if err != nil {
			return res.Response{Err: err}
		}
		return res.Response{Data: devs}
	})
}

// DeletePasskey 删除当前用户某个 Passkey 设备
func (userHandler *UserHandler) DeletePasskey() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)

		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}

		if err := userHandler.userService.DeletePasskey(userid, uint(id)); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}

// UpdatePasskeyDeviceName 更新 Passkey 设备名称
func (userHandler *UserHandler) UpdatePasskeyDeviceName() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		userid := ctx.MustGet("userid").(uint)

		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return res.Response{Msg: commonModel.INVALID_PARAMS, Err: err}
		}

		var req authModel.PasskeyUpdateDeviceNameReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			return res.Response{Msg: commonModel.INVALID_REQUEST_BODY, Err: err}
		}

		if err := userHandler.userService.UpdatePasskeyDeviceName(userid, uint(id), req.DeviceName); err != nil {
			return res.Response{Err: err}
		}
		return res.Response{}
	})
}
