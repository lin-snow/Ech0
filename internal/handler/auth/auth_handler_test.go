// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	cookieUtil "github.com/lin-snow/ech0/internal/util/cookie"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
)

type fakeAuthService struct {
	isTokenRevokedFn    func(jti string) bool
	revokeTokenFn       func(jti string, ttl time.Duration)
	exchangeOAuthCodeFn func(code string) (*authModel.TokenPair, error)
	accountStateFn      func(userID string) (userModel.AccountState, error)
}

func (f *fakeAuthService) AccountState(_ context.Context, userID string) (userModel.AccountState, error) {
	if f.accountStateFn != nil {
		return f.accountStateFn(userID)
	}
	return userModel.AccountState{Found: true, Role: userModel.RoleMember, Status: userModel.StatusActive}, nil
}

func (f *fakeAuthService) IsTokenRevoked(jti string) bool {
	if f.isTokenRevokedFn != nil {
		return f.isTokenRevokedFn(jti)
	}
	return false
}

func (f *fakeAuthService) RevokeToken(jti string, ttl time.Duration) {
	if f.revokeTokenFn != nil {
		f.revokeTokenFn(jti, ttl)
	}
}

func (f *fakeAuthService) ExchangeOAuthCode(code string) (*authModel.TokenPair, error) {
	if f.exchangeOAuthCodeFn != nil {
		return f.exchangeOAuthCodeFn(code)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeAuthService) Login(context.Context, *authModel.LoginDto) (*authModel.TokenPair, error) {
	panic("not called in auth handler tests")
}

func (f *fakeAuthService) BindOAuth(context.Context, string, string) (string, error) {
	panic("not called")
}

func (f *fakeAuthService) GetOAuthLoginURL(string, string) (string, error) {
	panic("not called")
}

func (f *fakeAuthService) HandleOAuthCallback(string, string, string) (string, error) {
	panic("not called")
}

func (f *fakeAuthService) GetOAuthInfo(context.Context, string) (userModel.OAuthInfoDto, error) {
	panic("not called")
}

func (f *fakeAuthService) PasskeyRegisterBegin(context.Context, string, string, string) (authModel.PasskeyRegisterBeginResp, error) {
	panic("not called")
}

func (f *fakeAuthService) PasskeyRegisterFinish(context.Context, string, string, string, json.RawMessage) error {
	panic("not called")
}

func (f *fakeAuthService) PasskeyLoginBegin(string, string) (authModel.PasskeyLoginBeginResp, error) {
	panic("not called")
}

func (f *fakeAuthService) PasskeyLoginFinish(string, string, string, json.RawMessage) (*authModel.TokenPair, error) {
	panic("not called")
}

func (f *fakeAuthService) ListPasskeys(context.Context) ([]authModel.PasskeyDeviceDto, error) {
	panic("not called")
}
func (f *fakeAuthService) DeletePasskey(context.Context, string) error { panic("not called") }
func (f *fakeAuthService) UpdatePasskeyDeviceName(context.Context, string, string) error {
	panic("not called")
}

func (f *fakeAuthService) PasskeyBoundary(context.Context) (string, []string) { return "", nil }

func (f *fakeAuthService) StartDeviceAuthorization(string, []string) (authModel.DeviceCodeResponse, error) {
	panic("not called in auth handler tests")
}

func (f *fakeAuthService) DeviceAuthorizationInfo(string) (authModel.DevicePendingInfo, error) {
	panic("not called in auth handler tests")
}

func (f *fakeAuthService) ApproveDeviceAuthorization(context.Context, string) error {
	panic("not called in auth handler tests")
}

func (f *fakeAuthService) PollDeviceAuthorization(string) (authModel.DeviceTokenResponse, error) {
	panic("not called in auth handler tests")
}

type fakeUserService struct {
	getUserByIDFn func(id string) (userModel.User, error)
}

func (f *fakeUserService) GetUserByID(id string) (userModel.User, error) {
	if f.getUserByIDFn != nil {
		return f.getUserByIDFn(id)
	}
	return userModel.User{}, errors.New("user not found")
}

func (f *fakeUserService) InitOwner(*authModel.RegisterDto) error { panic("not called") }
func (f *fakeUserService) Register(*authModel.RegisterDto) error  { panic("not called") }
func (f *fakeUserService) UpdateUser(context.Context, userModel.UserInfoDto) error {
	panic("not called")
}

func (f *fakeUserService) UpdateUserRole(context.Context, string, userModel.Role) error {
	panic("not called")
}
func (f *fakeUserService) TransferOwnership(context.Context, string) error { panic("not called") }
func (f *fakeUserService) UpdateUserStatus(context.Context, string, userModel.Status) error {
	panic("not called")
}

func (f *fakeUserService) GetAllUsers(context.Context) ([]userModel.User, error) {
	panic("not called")
}
func (f *fakeUserService) GetOwner() (userModel.User, error) { panic("not called") }
func (f *fakeUserService) DeleteUser(context.Context, string) error {
	panic("not called")
}

var testUser = userModel.User{
	ID:       "test-user-id-001",
	Username: "testuser",
}

func init() {
	gin.SetMode(gin.TestMode)
	cfg := config.Config()
	cfg.Security.JWTSecret = []byte("test-secret-for-auth-handler-unit-tests")
}

func issueRefreshToken(t *testing.T) string {
	t.Helper()
	claims := jwtUtil.CreateRefreshClaims(testUser)
	token, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to issue refresh token: %v", err)
	}
	return token
}

func issueAccessToken(t *testing.T) string {
	t.Helper()
	claims := jwtUtil.CreateClaims(testUser)
	token, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		t.Fatalf("failed to issue access token: %v", err)
	}
	return token
}

func addRefreshCookie(req *http.Request, token string) {
	req.AddCookie(&http.Cookie{
		Name:  "ech0_refresh_token",
		Value: token,
	})
}

type problemDoc struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func parseProblem(t *testing.T, rec *httptest.ResponseRecorder) problemDoc {
	t.Helper()
	var r problemDoc
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("failed to parse problem document: %v\nbody: %s", err, rec.Body.String())
	}
	return r
}

func assertRefreshCookieCleared(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == "ech0_refresh_token" && c.MaxAge < 0 {
			return
		}
	}
	t.Fatal("expected refresh token cookie to be cleared on failure")
}

func TestRefresh_NoCookie(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.REFRESH_TOKEN_INVALID {
		t.Fatalf("expected detail %q, got %q", commonModel.REFRESH_TOKEN_INVALID, res.Detail)
	}
	if res.Detail == "" {
		t.Fatal("problem document carries no detail")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	addRefreshCookie(req, "invalid.jwt.token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.REFRESH_TOKEN_INVALID {
		t.Fatalf("expected detail %q, got %q", commonModel.REFRESH_TOKEN_INVALID, res.Detail)
	}
	assertRefreshCookieCleared(t, rec)
}

func TestRefresh_TokenRevoked(t *testing.T) {
	auth := &fakeAuthService{
		isTokenRevokedFn: func(_ string) bool { return true },
	}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	addRefreshCookie(req, issueRefreshToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.TOKEN_REVOKED {
		t.Fatalf("expected detail %q, got %q", commonModel.TOKEN_REVOKED, res.Detail)
	}
	assertRefreshCookieCleared(t, rec)
}

func TestRefresh_AccountDisabled(t *testing.T) {
	auth := &fakeAuthService{
		isTokenRevokedFn: func(_ string) bool { return false },
		accountStateFn: func(_ string) (userModel.AccountState, error) {
			return userModel.AccountState{Found: true, Status: userModel.StatusDisabled}, nil
		},
	}
	user := &fakeUserService{
		getUserByIDFn: func(_ string) (userModel.User, error) {
			return testUser, nil
		},
	}
	h := NewAuthHandler(auth, user)
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	addRefreshCookie(req, issueRefreshToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.ACCOUNT_DISABLED {
		t.Fatalf("expected detail %q, got %q", commonModel.ACCOUNT_DISABLED, res.Detail)
	}
	assertRefreshCookieCleared(t, rec)
}

func TestRefresh_AccountStatusReadFailureFailsClosed(t *testing.T) {
	auth := &fakeAuthService{
		isTokenRevokedFn: func(_ string) bool { return false },
		accountStateFn: func(_ string) (userModel.AccountState, error) {
			return userModel.AccountState{}, errors.New("db down")
		},
	}
	user := &fakeUserService{
		getUserByIDFn: func(_ string) (userModel.User, error) {
			return testUser, nil
		},
	}
	h := NewAuthHandler(auth, user)
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	addRefreshCookie(req, issueRefreshToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != "internal server error" {
		t.Fatalf("expected detail %q, got %q", "internal server error", res.Detail)
	}
}

func TestRefresh_UserNotFound(t *testing.T) {
	auth := &fakeAuthService{
		isTokenRevokedFn: func(_ string) bool { return false },
	}
	user := &fakeUserService{
		getUserByIDFn: func(_ string) (userModel.User, error) {
			return userModel.User{}, errors.New("record not found")
		},
	}
	h := NewAuthHandler(auth, user)
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	addRefreshCookie(req, issueRefreshToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.REFRESH_TOKEN_INVALID {
		t.Fatalf("expected detail %q, got %q", commonModel.REFRESH_TOKEN_INVALID, res.Detail)
	}
	assertRefreshCookieCleared(t, rec)
}

func TestRefresh_Success(t *testing.T) {
	auth := &fakeAuthService{
		isTokenRevokedFn: func(_ string) bool { return false },
	}
	user := &fakeUserService{
		getUserByIDFn: func(_ string) (userModel.User, error) {
			return testUser, nil
		},
	}
	h := NewAuthHandler(auth, user)
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	oldToken := issueRefreshToken(t)
	oldClaims, err := jwtUtil.ParseRefreshToken(oldToken)
	if err != nil {
		t.Fatalf("failed to parse seeded refresh token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	addRefreshCookie(req, oldToken)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d\nbody: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if body.AccessToken == "" {
		t.Fatal("expected non-empty access_token")
	}
	if body.ExpiresIn != config.Config().Auth.Jwt.Expires {
		t.Fatalf("expected expires_in %d, got %d", config.Config().Auth.Jwt.Expires, body.ExpiresIn)
	}

	var renewed string
	for _, c := range rec.Result().Cookies() {
		if c.Name != cookieUtil.RefreshTokenCookieName {
			continue
		}
		renewed = c.Value
		if c.MaxAge != config.Config().Auth.Jwt.RefreshExpires {
			t.Fatalf("expected refreshed cookie MaxAge %d, got %d", config.Config().Auth.Jwt.RefreshExpires, c.MaxAge)
		}
	}
	if renewed == "" {
		t.Fatal("expected refresh token cookie to be re-issued on refresh")
	}
	newClaims, err := jwtUtil.ParseRefreshToken(renewed)
	if err != nil {
		t.Fatalf("renewed token must stay a valid refresh token: %v", err)
	}
	if newClaims.ID != oldClaims.ID {
		t.Fatalf("expected renewed session to keep JTI %q, got %q", oldClaims.ID, newClaims.ID)
	}
	if newClaims.ExpiresAt == nil || newClaims.ExpiresAt.Before(time.Now().Add(time.Duration(config.Config().Auth.Jwt.RefreshExpires-5)*time.Second)) {
		t.Fatalf("expected renewed expiry to slide to ~now+%ds, got %v", config.Config().Auth.Jwt.RefreshExpires, newClaims.ExpiresAt)
	}
}

func TestRefresh_AccessTokenCookie_Rejected(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/refresh", h.Refresh())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	addRefreshCookie(req, issueAccessToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("access token should be rejected by refresh endpoint, got status %d", rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.REFRESH_TOKEN_INVALID {
		t.Fatalf("expected detail %q, got %q", commonModel.REFRESH_TOKEN_INVALID, res.Detail)
	}
	assertRefreshCookieCleared(t, rec)
}

func TestLogout_NoCookieNoHeader(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/logout", h.Logout())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}

	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "ech0_refresh_token" && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("expected refresh token cookie to be cleared")
	}
}

func TestLogout_WithRefreshToken_RevokesJTI(t *testing.T) {
	var revokedJTIs []string
	auth := &fakeAuthService{
		revokeTokenFn: func(jti string, _ time.Duration) {
			revokedJTIs = append(revokedJTIs, jti)
		},
	}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/logout", h.Logout())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	addRefreshCookie(req, issueRefreshToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	if len(revokedJTIs) != 1 {
		t.Fatalf("expected 1 revoked JTI, got %d", len(revokedJTIs))
	}
}

func TestLogout_WithBothTokens_RevokesBoth(t *testing.T) {
	var revokedJTIs []string
	auth := &fakeAuthService{
		revokeTokenFn: func(jti string, _ time.Duration) {
			revokedJTIs = append(revokedJTIs, jti)
		},
	}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/logout", h.Logout())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	addRefreshCookie(req, issueRefreshToken(t))
	req.Header.Set("Authorization", "Bearer "+issueAccessToken(t))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	if len(revokedJTIs) != 2 {
		t.Fatalf("expected 2 revoked JTIs (refresh + access), got %d", len(revokedJTIs))
	}
}

// A token without an expiry cannot be verified any more (expiry is required),
// so logout ignores it instead of panicking or revoking an unverified JTI.
func TestLogout_TokenWithoutExpiryIsIgnored(t *testing.T) {
	var revoked []string
	auth := &fakeAuthService{
		revokeTokenFn: func(jti string, _ time.Duration) {
			revoked = append(revoked, jti)
		},
	}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/logout", h.Logout())

	legacy := jwt.NewWithClaims(jwt.SigningMethodHS256, authModel.MyClaims{
		Type:     authModel.TokenTypeSession,
		Issuer:   config.Config().Auth.Jwt.Issuer,
		Subject:  "u",
		Audience: jwt.ClaimStrings{config.Config().Auth.Jwt.Audience},
		ID:       "legacy-jti",
	})
	signed, err := legacy.SignedString(config.Config().Security.JWTSecret)
	if err != nil {
		t.Fatalf("sign legacy token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (no panic), got %d body=%s", rec.Code, rec.Body.String())
	}
	if len(revoked) != 0 {
		t.Fatalf("unverifiable token must not revoke a JTI, got %v", revoked)
	}
}

func TestLogout_InvalidTokens_StillOK(t *testing.T) {
	auth := &fakeAuthService{}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/logout", h.Logout())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	addRefreshCookie(req, "garbage-token")
	req.Header.Set("Authorization", "Bearer garbage-access")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("logout should be best-effort, expected %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestExchange_InvalidBody(t *testing.T) {
	h := NewAuthHandler(&fakeAuthService{}, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/exchange", h.Exchange())

	req := httptest.NewRequest(http.MethodPost, "/api/auth/exchange", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.INVALID_REQUEST_BODY {
		t.Fatalf("expected detail %q, got %q", commonModel.INVALID_REQUEST_BODY, res.Detail)
	}
}

func TestExchange_InvalidCode(t *testing.T) {
	auth := &fakeAuthService{
		exchangeOAuthCodeFn: func(_ string) (*authModel.TokenPair, error) {
			return nil, errors.New("code expired or already used")
		},
	}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/exchange", h.Exchange())

	body, _ := json.Marshal(authModel.ExchangeCodeReq{Code: "expired-code"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/exchange", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	res := parseProblem(t, rec)
	if res.Detail != commonModel.EXCHANGE_CODE_INVALID {
		t.Fatalf("expected detail %q, got %q", commonModel.EXCHANGE_CODE_INVALID, res.Detail)
	}
}

func TestExchange_Success(t *testing.T) {
	auth := &fakeAuthService{
		exchangeOAuthCodeFn: func(_ string) (*authModel.TokenPair, error) {
			return &authModel.TokenPair{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
				ExpiresIn:    900,
			}, nil
		},
	}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/exchange", h.Exchange())

	body, _ := json.Marshal(authModel.ExchangeCodeReq{Code: "valid-code"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/exchange", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d\nbody: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.AccessToken != "new-access-token" {
		t.Fatalf("expected access_token 'new-access-token', got %q", resp.AccessToken)
	}
	if resp.ExpiresIn != 900 {
		t.Fatalf("expected expires_in 900, got %d", resp.ExpiresIn)
	}

	foundCookie := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == cookieUtil.RefreshTokenCookieName {
			foundCookie = true
			if c.Value != "new-refresh-token" {
				t.Fatalf("expected refresh cookie value 'new-refresh-token', got %q", c.Value)
			}
			if !c.HttpOnly {
				t.Fatal("refresh cookie must be HttpOnly")
			}
		}
	}
	if !foundCookie {
		t.Fatal("expected refresh token cookie to be set")
	}
}

func TestExchange_CodeReuse_Rejected(t *testing.T) {
	callCount := 0
	auth := &fakeAuthService{
		exchangeOAuthCodeFn: func(_ string) (*authModel.TokenPair, error) {
			callCount++
			if callCount > 1 {
				return nil, errors.New("code already consumed")
			}
			return &authModel.TokenPair{
				AccessToken:  "token",
				RefreshToken: "refresh",
				ExpiresIn:    900,
			}, nil
		},
	}
	h := NewAuthHandler(auth, &fakeUserService{})
	r := gin.New()
	r.POST("/api/auth/exchange", h.Exchange())

	body, _ := json.Marshal(authModel.ExchangeCodeReq{Code: "one-time-code"})

	req1 := httptest.NewRequest(http.MethodPost, "/api/auth/exchange", bytes.NewBuffer(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first exchange should succeed, got %d", rec1.Code)
	}

	body2, _ := json.Marshal(authModel.ExchangeCodeReq{Code: "one-time-code"})
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/exchange", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("second exchange (code reuse) should fail, got %d", rec2.Code)
	}
}
