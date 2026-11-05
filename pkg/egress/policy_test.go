// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package egress

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAllowPrivatePermitsLocalTargets(t *testing.T) {
	SetAllowPrivate(true)
	t.Cleanup(func() { SetAllowPrivate(false) })

	for _, rawURL := range []string{
		"http://localhost:6277",
		"http://127.0.0.1:6277",
		"http://[::1]:6277",
		"http://192.168.1.10:6277",
		"http://10.0.0.5:6277",
		"http://100.64.0.1:6277",
		"http://host.docker.internal:6277",
	} {
		if err := Validate(rawURL); err != nil {
			t.Fatalf("Validate(%q) = %v, want nil with AllowPrivate", rawURL, err)
		}
	}

	// Scheme/userinfo validation still applies with the private guard off.
	if err := Validate("ftp://localhost"); err == nil {
		t.Fatal("Validate(ftp://localhost) = nil, want scheme rejection")
	}
	if err := Validate("http://user:pass@localhost"); err == nil {
		t.Fatal("Validate(userinfo) = nil, want userinfo rejection")
	}

	// End-to-end: the guarded dialer must accept a loopback target too.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	body, err := Fetch(srv.URL, http.MethodGet, Header{}, time.Second)
	if err != nil {
		t.Fatalf("Fetch(%q) error = %v, want nil with AllowPrivate", srv.URL, err)
	}
	if string(body) != "ok" {
		t.Fatalf("Fetch body = %q, want %q", body, "ok")
	}
}

func TestAllowPrivateRestoresGuard(t *testing.T) {
	SetAllowPrivate(true)
	SetAllowPrivate(false)

	if err := Validate("http://127.0.0.1:6277"); err == nil {
		t.Fatal("Validate(loopback) = nil after SetAllowPrivate(false)")
	}
	if err := Validate("http://localhost:6277"); err == nil {
		t.Fatal("Validate(localhost) = nil after SetAllowPrivate(false)")
	}
}
