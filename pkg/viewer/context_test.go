package viewer

import (
	"context"
	"net/http"
	"testing"
)

func TestWithAndFromContext(t *testing.T) {
	ctx := context.Background()
	v := NewScopedUserViewer("u1", "integration")
	ctx = WithContext(ctx, v)

	got, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("expected viewer in context")
	}
	if got.UserID() != "u1" {
		t.Fatalf("unexpected user id: %s", got.UserID())
	}
	if got.TokenScope() != "integration" {
		t.Fatalf("unexpected token scope: %s", got.TokenScope())
	}
}

func TestMustFromContextFallback(t *testing.T) {
	got := MustFromContext(context.Background())
	if got.UserID() != "" {
		t.Fatalf("expected empty user id fallback viewer")
	}
	if got.TokenScope() != "" {
		t.Fatalf("expected empty token scope fallback viewer")
	}
}

func TestWithRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/ping", nil)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req = WithRequest(req, NewUserViewer("u1"))

	got := MustFromContext(req.Context())
	if got.UserID() != "u1" {
		t.Fatalf("unexpected user id: %s", got.UserID())
	}
}

func TestAttachToRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/ping", nil)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	AttachToRequest(&req, NewUserViewer("u2"))

	got := MustFromContext(req.Context())
	if got.UserID() != "u2" {
		t.Fatalf("unexpected user id: %s", got.UserID())
	}
}
