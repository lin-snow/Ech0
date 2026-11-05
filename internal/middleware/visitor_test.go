// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	actorUtil "github.com/lin-snow/ech0/internal/util/actor"
)

func TestVisitorIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	run := func(visitorHeader string) (string, bool) {
		t.Helper()
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.RemoteAddr = "203.0.113.9:1234"
		if visitorHeader != "" {
			c.Request.Header.Set(VisitorIDHeader, visitorHeader)
		}
		VisitorIdentity()(c)
		return actorUtil.FromContext(c.Request.Context()), c.IsAborted()
	}

	t.Run("valid visitor id becomes the actor", func(t *testing.T) {
		value, aborted := run("11111111-1111-4111-8111-111111111111")
		if aborted {
			t.Fatal("middleware must not abort the request")
		}
		if value != "v:11111111-1111-4111-8111-111111111111" {
			t.Fatalf("unexpected actor %q", value)
		}
	})

	t.Run("malformed visitor id falls back to the hashed ip", func(t *testing.T) {
		value, _ := run("not-a-uuid")
		if value == "" || value[:3] != "ip:" {
			t.Fatalf("expected ip fallback, got %q", value)
		}
	})

	t.Run("missing header falls back to the hashed ip", func(t *testing.T) {
		value, _ := run("")
		if value == "" || value[:3] != "ip:" {
			t.Fatalf("expected ip fallback, got %q", value)
		}
	})
}
