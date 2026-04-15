package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/visitor"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTemplatesRouter() *gin.Engine {
	r := gin.New()
	r.NoRoute(NewWebHandler(visitor.NewTracker()).Templates())
	return r
}

func TestTemplates_APIPathReturnsNotFound(t *testing.T) {
	r := newTemplatesRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/not-found", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestTemplates_SPAFallbackSetsNoStoreHeaders(t *testing.T) {
	r := newTemplatesRouter()

	req := httptest.NewRequest(http.MethodGet, "/app/dashboard", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
		t.Fatalf("expected Cache-Control no-store header, got %q", got)
	}
	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected Pragma no-cache, got %q", got)
	}
	if got := rec.Header().Get("Expires"); got != "0" {
		t.Fatalf("expected Expires 0, got %q", got)
	}
}

func TestSetCacheControlHeader_DefaultStaticUsesShortCache(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	setCacheControlHeader(ctx, "/favicon.ico")

	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Fatalf("expected short cache header, got %q", got)
	}
}

func TestSetCacheControlHeader_IndexDisablesCacheStorage(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	setCacheControlHeader(ctx, "/index.html")

	if got := rec.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
		t.Fatalf("expected Cache-Control no-store header, got %q", got)
	}
	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected Pragma no-cache, got %q", got)
	}
	if got := rec.Header().Get("Expires"); got != "0" {
		t.Fatalf("expected Expires 0, got %q", got)
	}
}
