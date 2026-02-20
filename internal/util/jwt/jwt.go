package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	cryptoUtil "github.com/lin-snow/ech0/internal/util/crypto"
)

// CreateClaims 创建Claims
func CreateClaims(user userModel.User) jwt.Claims {
	leeway := time.Second * 60 // 允许的时间偏差
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   config.Config.Auth.Jwt.Issuer,
			Subject:  user.Username,
			Audience: jwt.ClaimStrings{config.Config.Auth.Jwt.Audience},
			ExpiresAt: jwt.NewNumericDate(
				time.Now().UTC().Add(time.Duration(config.Config.Auth.Jwt.Expires) * time.Second),
			),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(time.Now().UTC().Add(-leeway)),
		},
	}

	return claims
}

// CreateClaims 创建Claims 带过期时间
func CreateClaimsWithExpiry(user userModel.User, expiry int64) jwt.Claims {
	leeway := time.Second * 60 // 允许的时间偏差
	claims := authModel.MyClaims{
		Userid:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    config.Config.Auth.Jwt.Issuer,
			Subject:   user.Username,
			Audience:  jwt.ClaimStrings{config.Config.Auth.Jwt.Audience},
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
	return token.SignedString(config.JWT_SECRET)
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*authModel.MyClaims, error) {
	claims := &authModel.MyClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			return config.JWT_SECRET, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*authModel.MyClaims); ok {
		return claims, nil
	}

	log.Println("unknown claims type, cannot proceed")
	return nil, errors.New("unknown claims type, cannot proceed")
}

// GenerateOAuthState 生成 OAuth2 state token
func GenerateOAuthState(
	action string,
	userID uint,
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
	state, err := token.SignedString(config.JWT_SECRET)
	if err != nil {
		return "", "", err
	}

	return state, nonce, nil
}

// ParseOAuthState 解析并验证 OAuth2 state token
func ParseOAuthState(stateStr string) (*authModel.OAuthState, error) {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(stateStr, claims, func(token *jwt.Token) (interface{}, error) {
		return config.JWT_SECRET, nil
	})
	if err != nil {
		return nil, err
	}

	return &authModel.OAuthState{
		Action:   claims["action"].(string),
		UserID:   uint(claims["user_id"].(float64)),
		Nonce:    claims["nonce"].(string),
		Redirect: claims["redirect"].(string),
		Exp:      int64(claims["exp"].(float64)),
		Provider: claims["provider"].(string),
	}, nil
}

// ParseAndVerifyIDToken 解析并验证 OIDC id_token
func ParseAndVerifyIDToken(idToken, issuer, jwksURL, clientID string) (jwt.MapClaims, error) {
	if idToken == "" {
		return nil, errors.New("id_token 为空")
	}
	if jwksURL == "" {
		return nil, errors.New("JWKS URL 为空")
	}

	keySet, err := fetchJWKSPublicKeys(jwksURL)
	if err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{}
	parser := jwt.NewParser(jwt.WithLeeway(time.Minute))
	validator := jwt.NewValidator(jwt.WithLeeway(time.Minute))

	token, err := parser.ParseWithClaims(
		idToken,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			kid, _ := token.Header["kid"].(string)
			if kid != "" {
				if key, ok := keySet[kid]; ok {
					return key, nil
				}
				return nil, errors.New("未找到匹配 kid 的 JWKS 公钥")
			}

			if len(keySet) == 1 {
				for _, key := range keySet {
					return key, nil
				}
			}

			return nil, errors.New("id_token 缺少 kid 且 JWKS 含多把公钥")
		},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("id_token 验证失败")
	}

	if err := validator.Validate(claims); err != nil {
		return nil, err
	}

	if issuer != "" {
		if iss, ok := claims["iss"].(string); !ok || iss != issuer {
			return nil, errors.New("id_token issuer 不匹配")
		}
	}

	if clientID != "" {
		if err := validateAudience(claims, clientID); err != nil {
			return nil, err
		}
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
			log.Printf("解析 JWKS 公钥失败: %v", parseErr)
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
