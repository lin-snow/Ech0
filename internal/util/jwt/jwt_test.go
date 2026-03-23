package util

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

func TestCreateClaims_WithSessionType(t *testing.T) {
	user := userModel.User{
		ID:       "u-1",
		Username: "alice",
	}
	claimsAny := CreateClaims(user)
	claims, ok := claimsAny.(authModel.MyClaims)
	if !ok {
		t.Fatalf("unexpected claims type %T", claimsAny)
	}
	if claims.Type != authModel.TokenTypeSession {
		t.Fatalf("expected typ=%s, got %s", authModel.TokenTypeSession, claims.Type)
	}
}

func TestParseToken_RejectsTokenWithoutType(t *testing.T) {
	legacyToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "u-legacy",
		"username": "legacy",
		"exp":      time.Now().UTC().Add(10 * time.Minute).Unix(),
		"iat":      time.Now().UTC().Unix(),
		"nbf":      time.Now().UTC().Add(-1 * time.Minute).Unix(),
	})
	tokenString, err := legacyToken.SignedString(config.Config().Security.JWTSecret)
	if err != nil {
		t.Fatalf("failed to sign legacy token: %v", err)
	}

	if _, err := ParseToken(tokenString); err == nil {
		t.Fatal("expected ParseToken to reject token without typ")
	}
}

func TestParseOAuthState_RoundTrip(t *testing.T) {
	state, nonce, err := GenerateOAuthState("login", "u1", "https://example.com/auth", "custom")
	if err != nil {
		t.Fatalf("GenerateOAuthState failed: %v", err)
	}
	if nonce == "" {
		t.Fatalf("expected nonce not empty")
	}

	parsed, err := ParseOAuthState(state)
	if err != nil {
		t.Fatalf("ParseOAuthState failed: %v", err)
	}
	if parsed.Action != "login" {
		t.Fatalf("unexpected action: %s", parsed.Action)
	}
	if parsed.UserID != "u1" {
		t.Fatalf("unexpected user id: %s", parsed.UserID)
	}
	if parsed.Provider != "custom" {
		t.Fatalf("unexpected provider: %s", parsed.Provider)
	}
}

func TestParseAndVerifyIDToken_NonceSuccess(t *testing.T) {
	issuer := "https://issuer.example.com"
	clientID := "client-1"
	nonce := "nonce-ok"

	privateKey, jwksServer, kid := prepareJWKS(t)
	defer jwksServer.Close()

	rawToken := signIDToken(t, privateKey, kid, issuer, clientID, nonce)
	claims, err := ParseAndVerifyIDToken(rawToken, issuer, jwksServer.URL, clientID, nonce)
	if err != nil {
		t.Fatalf("ParseAndVerifyIDToken failed: %v", err)
	}
	if claims["sub"] != "user-123" {
		t.Fatalf("unexpected sub: %v", claims["sub"])
	}
}

func TestParseAndVerifyIDToken_NonceMismatch(t *testing.T) {
	issuer := "https://issuer.example.com"
	clientID := "client-1"

	privateKey, jwksServer, kid := prepareJWKS(t)
	defer jwksServer.Close()

	rawToken := signIDToken(t, privateKey, kid, issuer, clientID, "nonce-actual")
	_, err := ParseAndVerifyIDToken(rawToken, issuer, jwksServer.URL, clientID, "nonce-expected")
	if err == nil {
		t.Fatalf("expected nonce mismatch error")
	}
}

func prepareJWKS(t *testing.T) (*rsa.PrivateKey, *httptest.Server, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key failed: %v", err)
	}
	kid := "test-kid"
	n := base64.RawURLEncoding.EncodeToString(privateKey.PublicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes())
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": kid,
				"n":   n,
				"e":   e,
			},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jwks)
	}))
	return privateKey, server, kid
}

func signIDToken(
	t *testing.T,
	privateKey *rsa.PrivateKey,
	kid, issuer, clientID, nonce string,
) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss":   issuer,
		"aud":   clientID,
		"sub":   "user-123",
		"nonce": nonce,
		"iat":   time.Now().Add(-1 * time.Minute).Unix(),
		"exp":   time.Now().Add(10 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	return raw
}
