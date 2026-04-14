package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestJWTAuthMiddleware_RejectsTokenWithoutType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware(nil))
	r.GET("/api/user", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	legacyToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "legacy",
		"username": "legacy",
		"exp":      time.Now().UTC().Add(10 * time.Minute).Unix(),
	})
	tokenString, err := legacyToken.SignedString(config.Config().Security.JWTSecret)
	if err != nil {
		t.Fatalf("sign legacy token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestJWTAuthMiddleware_AllowsAnonymousPublicEchoEvenWithInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware(nil))
	r.GET("/api/echo/page", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/echo/page", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestJWTAuthMiddleware_RejectsAnonymousS3Settings(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware(nil))
	r.GET("/api/s3/settings", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/s3/settings", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestJWTAuthMiddleware_RejectsAdminScopeTokenFromQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware(nil))
	r.GET("/api/settings", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "admin-1", Username: "admin"}
	claims := jwtUtil.CreateAccessClaimsWithExpiry(
		user,
		int64(time.Hour),
		[]string{authModel.ScopeAdminSettings},
		authModel.AudiencePublic,
		"jti-admin-query",
	)
	tokenString, err := jwtUtil.GenerateToken(claims)
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/settings?token="+tokenString, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
	if got := parseAuthErrorCode(rec.Body.Bytes()); got != "TOKEN_TRANSPORT_FORBIDDEN" {
		t.Fatalf("expected error code TOKEN_TRANSPORT_FORBIDDEN, got %s", got)
	}
}

func TestJWTAuthMiddleware_AllowsSessionType(t *testing.T) {
	initMiddlewareTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuthMiddleware(nil))
	r.GET("/api/user", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	user := userModel.User{ID: "user-1", Username: "alice"}
	tokenString, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	if err != nil {
		t.Fatalf("generate token failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func initMiddlewareTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
	database.SetDB(db)
}

func parseAuthErrorCode(body []byte) string {
	var payload struct {
		ErrorCode string `json:"error_code"`
	}
	_ = json.Unmarshal(body, &payload)
	return payload.ErrorCode
}
