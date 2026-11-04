// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package reqinfo

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func request(host string, tlsOn bool, proto, forwardedHost string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = host
	if tlsOn {
		r.TLS = &tls.ConnectionState{}
	}
	if proto != "" {
		r.Header.Set(forwardedProtoHeader, proto)
	}
	if forwardedHost != "" {
		r.Header.Set(forwardedHostHeader, forwardedHost)
	}
	return r
}

func TestScheme(t *testing.T) {
	cases := []struct {
		name string
		req  *http.Request
		want string
	}{
		{"plain http", request("internal:6277", false, "", ""), "http"},
		{"tls wins", request("internal:6277", true, "http", ""), "https"},
		{"forwarded https", request("internal:6277", false, "https", ""), "https"},
		{"forwarded list takes first", request("internal", false, "https, http", ""), "https"},
		{"forwarded case and spaces", request("internal", false, " HTTPS ", ""), "https"},
		{"unknown proto ignored", request("internal", false, "ftp", ""), "http"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Scheme(tc.req); got != tc.want {
				t.Fatalf("Scheme() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHostAndOrigin(t *testing.T) {
	r := request("internal:6277", false, "https, http", "ech0.app, proxy.local")
	if got := Host(r); got != "ech0.app" {
		t.Fatalf("Host() = %q", got)
	}
	if got := Origin(r); got != "https://ech0.app" {
		t.Fatalf("Origin() = %q", got)
	}

	plain := request("ech0.example", false, "", "")
	if got := Origin(plain); got != "http://ech0.example" {
		t.Fatalf("Origin() = %q", got)
	}
}

func TestIsSameOrigin(t *testing.T) {
	r := request("internal:6277", false, "https", "ech0.app")
	cases := []struct {
		origin string
		want   bool
	}{
		{"https://ech0.app", true},
		{"https://ech0.app/", true},
		{"https://ECH0.APP", true},
		{"http://ech0.app", false},
		{"https://evil.example", false},
		{"", false},
		{"not a url", false},
	}
	for _, tc := range cases {
		if got := IsSameOrigin(tc.origin, r); got != tc.want {
			t.Errorf("IsSameOrigin(%q) = %v, want %v", tc.origin, got, tc.want)
		}
	}
}
