package model

import (
	"github.com/golang-jwt/jwt/v5"
)

// MyClaims 是自定义的 JWT 声明结构体
type MyClaims struct {
	Userid   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

const (
	// MAX_USER_COUNT 定义最大用户数量
	MAX_USER_COUNT = 5
	// NO_USER_LOGINED 定义未登录用户的 ID
	NO_USER_LOGINED = uint(0)
)

type OAuth2Action string

const (
	// OAuth2ActionLogin 表示登录操作
	OAuth2ActionLogin OAuth2Action = "login"
	// OAuth2ActionRegister 表示注册操作
	OAuth2ActionRegister OAuth2Action = "register"
	// OAuth2ActionBind 表示绑定操作
	OAuth2ActionBind OAuth2Action = "bind"
)

type OAuthState struct {
	Action   string `json:"action"`
	UserID   uint   `json:"user_id,omitempty"`
	Nonce    string `json:"nonce"`
	Redirect string `json:"redirect,omitempty"`
	Exp      int64  `json:"exp"`
	Provider string `json:"provider,omitempty"`
}

// GitHubTokenResponse GitHub token 响应结构
type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// GitHubUser GitHub 用户信息
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// GoogleTokenResponse Google token 响应结构
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
}

// GoogleUser Google 用户信息
type GoogleUser struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// QQTokenResponse QQ token 响应结构
// OpenID需要通过单独的API调用获取，不在token响应中
type QQTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// QQOpenIDResponse QQ OpenID 响应结构
type QQOpenIDResponse struct {
	ClientID string `json:"client_id"`
	OpenID   string `json:"openid"`
}

// QQUser QQ 用户信息
// Ret和Msg字段用于QQ互联API的错误处理
type QQUser struct {
	Ret          int    `json:"ret"`            // 返回码，0表示成功
	Msg          string `json:"msg"`            // 错误信息
	Nickname     string `json:"nickname"`       // 用户昵称
	FigureURL    string `json:"figureurl"`      // 30x30头像
	FigureURL1   string `json:"figureurl_1"`    // 50x50头像
	FigureURL2   string `json:"figureurl_2"`    // 100x100头像
	FigureURLQQ1 string `json:"figureurl_qq_1"` // 40x40头像
	FigureURLQQ2 string `json:"figureurl_qq_2"` // 100x100头像
	Gender       string `json:"gender"`         // 性别
}
