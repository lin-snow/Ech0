package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/database"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFullAccessOnly_BlocksIntegrationToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDB(t)

	engine := gin.New()
	group := engine.Group("/api")
	group.Use(JWTAuthMiddleware(), FullAccessOnly())
	group.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+mustIssueToken(t, authModel.TokenScopeIntegration))
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestFullAccessOnly_AllowsFullAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDB(t)

	engine := gin.New()
	group := engine.Group("/api")
	group.Use(JWTAuthMiddleware(), FullAccessOnly())
	group.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+mustIssueToken(t, ""))
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func mustIssueToken(t *testing.T, scope string) string {
	t.Helper()

	token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaimsWithScopeAndExpiry(userModel.User{
		ID:       "u1",
		Username: "alice",
	}, scope, 3600))
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}
	return token
}

func initTestDB(t *testing.T) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("init test db failed: %v", err)
	}
	database.SetDB(db)
}
