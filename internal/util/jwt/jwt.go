package util

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
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

// CreateClaims 创建Claims
func CreateClaims(user userModel.User) jwt.Claims {
	leeway := time.Second * 60 // 允许的时间偏差
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		Type:     authModel.TokenTypeSession,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   config.Config().Auth.Jwt.Issuer,
			Subject:  user.Username,
			Audience: jwt.ClaimStrings{config.Config().Auth.Jwt.Audience},
			ExpiresAt: jwt.NewNumericDate(
				time.Now().UTC().Add(time.Duration(config.Config().Auth.Jwt.Expires) * time.Second),
			),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC().Add(-leeway)),
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

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*authModel.MyClaims, error) {
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
		if claims.Type != authModel.TokenTypeSession && claims.Type != authModel.TokenTypeAccess {
			return nil, errors.New("invalid token typ")
		}
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

// validateAudience 校验 aud/azp，确保包含客户端且多 aud 时符合 OIDC 规范
func validateAudience(claims jwt.MapClaims, clientID string) error {
	audRaw, ok := claims["aud"]
	if !ok {
		return errors.New("id_token 缺少 aud 声明")
	}

	hasAud := false
	multiAud := false

	switch v := audRaw.(type) {
	case string:
		if v == clientID {
			hasAud = true
		}
	case []string:
		if len(v) > 1 {
			multiAud = true
		}
		for _, item := range v {
			if item == clientID {
				hasAud = true
				break
			}
		}
	case []any:
		if len(v) > 1 {
			multiAud = true
		}
		for _, item := range v {
			if fmt.Sprint(item) == clientID {
				hasAud = true
				break
			}
		}
	default:
		return errors.New("id_token aud 类型不支持")
	}

	if !hasAud {
		return errors.New("id_token aud 不包含当前客户端")
	}

	if multiAud {
		azp, _ := claims["azp"].(string)
		if azp == "" {
			return errors.New("id_token 多 aud 但缺少 azp")
		}
		if azp != clientID {
			return errors.New("id_token azp 不匹配客户端")
		}
	}

	return nil
}

// jwkKey 表示 JWKS 中的单个 Key
type jwkKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// fetchJWKSPublicKeys 拉取并解析 JWKS 公钥集合
func fetchJWKSPublicKeys(jwksURL string) (map[string]any, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("获取 JWKS 失败: " + string(body))
	}

	var payload struct {
		Keys []jwkKey `json:"keys"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if len(payload.Keys) == 0 {
		return nil, errors.New("JWKS keys 为空")
	}

	keySet := make(map[string]any)
	singleKey := len(payload.Keys) == 1

	for _, key := range payload.Keys {
		if key.Use != "" && key.Use != "sig" {
			continue
		}

		var (
			publicKey any
			parseErr  error
		)

		switch key.Kty {
		case "RSA":
			publicKey, parseErr = buildRSAPublicKey(key.N, key.E)
		case "EC":
			publicKey, parseErr = buildECPublicKey(key.Crv, key.X, key.Y)
		default:
			continue
		}

		if parseErr != nil {
			logUtil.Warn("parse jwks public key failed", zap.String("module", "jwt"), zap.Error(parseErr))
			continue
		}

		kid := key.Kid
		if kid == "" {
			kid = key.Alg
		}
		if kid == "" && singleKey {
			kid = "default"
		}
		if kid == "" {
			continue
		}

		keySet[kid] = publicKey
	}

	if len(keySet) == 0 {
		return nil, errors.New("未能解析任何可用的 JWKS 公钥")
	}

	return keySet, nil
}

// buildRSAPublicKey 根据 JWK 的 n/e 构建 RSA 公钥
func buildRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	if nStr == "" || eStr == "" {
		return nil, errors.New("RSA JWK 缺少 n 或 e")
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	eBig := new(big.Int).SetBytes(eBytes)
	if !eBig.IsInt64() {
		return nil, errors.New("RSA 指数超出范围")
	}

	e := int(eBig.Int64())
	if e <= 0 {
		return nil, errors.New("RSA 指数无效")
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// buildECPublicKey 根据 JWK 的 crv/x/y 构建 EC 公钥
func buildECPublicKey(crv, xStr, yStr string) (*ecdsa.PublicKey, error) {
	if crv == "" || xStr == "" || yStr == "" {
		return nil, errors.New("EC JWK 缺少 crv/x/y")
	}

	curve := map[string]elliptic.Curve{
		"P-256": elliptic.P256(),
		"P-384": elliptic.P384(),
		"P-521": elliptic.P521(),
	}[crv]
	if curve == nil {
		return nil, errors.New("不支持的 EC 曲线: " + crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, err
	}

	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("EC 公钥不在曲线上")
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
