// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package viewer

import (
	"context"
	"net/http"
	"testing"
)

func TestWithAndFromContext(t *testing.T) {
	ctx := context.Background()
	v := NewSession("u1", "", nil)
	ctx = WithContext(ctx, v)

	got, ok := FromContext(ctx)
	if !ok {
		t.Fatalf("expected viewer in context")
	}
	if got.UserID() != "u1" {
		t.Fatalf("unexpected user id: %s", got.UserID())
	}
}

func TestMustFromContextFallback(t *testing.T) {
	got := MustFromContext(context.Background())
	if got.UserID() != "" {
		t.Fatalf("expected empty user id fallback viewer")
	}
}

func TestWithRequest(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/ping", nil)
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req = WithRequest(req, NewSession("u1", "", nil))

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
	AttachToRequest(&req, NewSession("u2", "", nil))

	got := MustFromContext(req.Context())
	if got.UserID() != "u2" {
		t.Fatalf("unexpected user id: %s", got.UserID())
	}
}
