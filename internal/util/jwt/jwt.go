package util

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

// CreateClaims 创建浏览器会话的 access token claims。
// typ=session, 有效期由 ECH0_JWT_EXPIRES 控制（默认 900s = 15 分钟），
// 每个 token 带唯一 JTI 以支持黑名单吊销。
func CreateClaims(user userModel.User) jwt.Claims {
	leeway := time.Second * 60
	now := time.Now().UTC()
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		Type:     authModel.TokenTypeSession,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   config.Config().Auth.Jwt.Issuer,
			Subject:  user.Username,
			Audience: jwt.ClaimStrings{config.Config().Auth.Jwt.Audience},
			ExpiresAt: jwt.NewNumericDate(
				now.Add(time.Duration(config.Config().Auth.Jwt.Expires) * time.Second),
			),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-leeway)),
			ID:        cryptoUtil.GenerateRandomString(16),
		},
	}

	return claims
}

// CreateRefreshClaims 创建静默刷新专用的 refresh token claims。
// typ=refresh, 有效期由 ECH0_JWT_REFRESH_EXPIRES 控制（默认 604800s = 7 天），
// 通过 HttpOnly Cookie 传递给浏览器，JS 无法读取。
func CreateRefreshClaims(user userModel.User) jwt.Claims {
	leeway := time.Second * 60
	now := time.Now().UTC()
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		Type:     authModel.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   config.Config().Auth.Jwt.Issuer,
			Subject:  user.Username,
			Audience: jwt.ClaimStrings{config.Config().Auth.Jwt.Audience},
			ExpiresAt: jwt.NewNumericDate(
				now.Add(time.Duration(config.Config().Auth.Jwt.RefreshExpires) * time.Second),
			),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-leeway)),
			ID:        cryptoUtil.GenerateRandomString(16),
		},
	}

	return claims
}

// CreateClaims 创建Claims 带过期时间
func CreateClaimsWithExpiry(user userModel.User, expiry int64) jwt.Claims {
	return CreateAccessClaimsWithExpiry(user, expiry, nil, "", "")
}

func CreateAccessClaimsWithExpiry(
	user userModel.User,
	expiry int64,
	scopes []string,
	audience string,
	jti string,
) jwt.Claims {
	leeway := time.Second * 60 // 允许的时间偏差
	audiences := jwt.ClaimStrings{config.Config().Auth.Jwt.Audience}
	if audience != "" {
		audiences = jwt.ClaimStrings{audience}
	}
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		Type:     authModel.TokenTypeAccess,
		Scopes:   scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Config().Auth.Jwt.Issuer,
			Subject:   user.Username,
			Audience:  audiences,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC().Add(-leeway)),
		},
	}

	// expiry = 0 表示永不过期
	if expiry > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(expiry) * time.Second))
	}

	return claims
}

// GenerateToken 生成JWT Token
func GenerateToken(claim jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(config.Config().Security.JWTSecret)
}

// ParseToken 解析 JWT（仅接受 typ=session / typ=access）。
// 用于 JWTAuthMiddleware 鉴权和 AuthHandler.Logout 吊销 access_token。
// 不接受 typ=refresh，防止 refresh_token 被当作 access_token 使用。
func ParseToken(tokenString string) (*authModel.MyClaims, error) {
	claims, err := parseTokenRaw(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != authModel.TokenTypeSession && claims.Type != authModel.TokenTypeAccess {
		return nil, errors.New("invalid token typ")
	}
	return claims, nil
}

// ParseRefreshToken 解析 refresh token（仅接受 typ=refresh）。
// 用于 AuthHandler.Refresh 和 AuthHandler.Logout。
// 不接受 session/access 类型，防止 access_token 被用于刷新。
func ParseRefreshToken(tokenString string) (*authModel.MyClaims, error) {
	claims, err := parseTokenRaw(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.Type != authModel.TokenTypeRefresh {
		return nil, errors.New("invalid token typ: expected refresh")
	}
	return claims, nil
}

// parseTokenRaw 是 ParseToken 和 ParseRefreshToken 的公共底层：
// 验证 JWT 签名和标准 claims（exp/iat/nbf），但不检查 typ 字段。
// typ 检查由上层调用方负责，确保 token 类型不被混用。
func parseTokenRaw(tokenString string) (*authModel.MyClaims, error) {
	claims := &authModel.MyClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return config.Config().Security.JWTSecret, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*authModel.MyClaims); ok {
		return claims, nil
	}

	logUtil.Warn("parse token claims type mismatch", zap.String("module", "jwt"))
	return nil, errors.New("unknown claims type, cannot proceed")
}

// GenerateOAuthState 生成 OAuth2 state token
func GenerateOAuthState(
	action string,
	userID string,
	redirect, provider string,
) (string, string, error) {
	now := time.Now().UTC()
	expiration := now.Add(10 * time.Minute)

	nonce := cryptoUtil.GenerateRandomString(16)

	claims := jwt.MapClaims{
		"action":   action,
		"user_id":  userID,
		"nonce":    nonce,
		"redirect": redirect,
		"exp":      expiration.Unix(),
		"iat":      now.Unix(),
		"provider": provider,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	state, err := token.SignedString(config.Config().Security.JWTSecret)
	if err != nil {
		return "", "", err
	}

	return state, nonce, nil
}

// ParseOAuthState 解析并验证 OAuth2 state token
func ParseOAuthState(stateStr string) (*authModel.OAuthState, error) {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(stateStr, claims, func(token *jwt.Token) (interface{}, error) {
		return config.Config().Security.JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}

	getStringClaim := func(key string) (string, error) {
		v, ok := claims[key]
		if !ok {
			return "", fmt.Errorf("oauth state 缺少 %s", key)
		}
		s, ok := v.(string)
		if !ok || s == "" {
			return "", fmt.Errorf("oauth state %s 非法", key)
		}
		return s, nil
	}

	action, err := getStringClaim("action")
	if err != nil {
		return nil, err
	}
	nonce, err := getStringClaim("nonce")
	if err != nil {
		return nil, err
	}
	redirect, err := getStringClaim("redirect")
	if err != nil {
		return nil, err
	}
	provider, err := getStringClaim("provider")
	if err != nil {
		return nil, err
	}

	expRaw, ok := claims["exp"]
	if !ok {
		return nil, errors.New("oauth state 缺少 exp")
	}
	expFloat, ok := expRaw.(float64)
	if !ok {
		return nil, errors.New("oauth state exp 非法")
	}

	return &authModel.OAuthState{
		Action:   action,
		UserID:   fmt.Sprint(claims["user_id"]),
		Nonce:    nonce,
		Redirect: redirect,
		Exp:      int64(expFloat),
		Provider: provider,
	}, nil
}

// ParseAndVerifyIDToken 解析并验证 OIDC id_token
func ParseAndVerifyIDToken(idToken, issuer, jwksURL, clientID, expectedNonce string) (jwt.MapClaims, error) {
	if idToken == "" {
		return nil, errors.New("id_token 为空")
	}
	if issuer == "" {
		return nil, errors.New("OIDC issuer 为空")
	}
	if jwksURL == "" {
		return nil, errors.New("JWKS URL 为空")
	}
	if clientID == "" {
		return nil, errors.New("OIDC client_id 为空")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	remoteKeySet := oidc.NewRemoteKeySet(ctx, jwksURL)
	verifier := oidc.NewVerifier(issuer, remoteKeySet, &oidc.Config{ClientID: clientID})
	idTokenObj, err := verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{}
	if err := idTokenObj.Claims(&claims); err != nil {
		return nil, err
	}

	if expectedNonce != "" {
		if idTokenObj.Nonce == "" {
			return nil, errors.New("id_token 缺少 nonce 声明")
		}
		if idTokenObj.Nonce != expectedNonce {
			return nil, errors.New("id_token nonce 不匹配")
		}
		claims["nonce"] = idTokenObj.Nonce
	}

	subVal, ok := claims["sub"]
	if !ok {
		return nil, errors.New("id_token 缺少 sub 声明")
	}

	switch v := subVal.(type) {
	case string:
		if v == "" {
			return nil, errors.New("id_token sub 为空")
		}
		claims["sub"] = v
	default:
		subStr := fmt.Sprint(v)
		if subStr == "" || subStr == "<nil>" {
			return nil, errors.New("id_token sub 无法转换为字符串")
		}
		claims["sub"] = subStr
	}

	return claims, nil
}
