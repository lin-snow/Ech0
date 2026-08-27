// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package egress

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{
			name:    "allow_public_https_domain",
			rawURL:  "https://example.com",
			wantErr: false,
		},
		{
			name:    "allow_public_http_ip",
			rawURL:  "http://93.184.216.34",
			wantErr: false,
		},
		{
			name:    "reject_non_http_scheme",
			rawURL:  "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "reject_loopback_ipv4",
			rawURL:  "http://127.0.0.1",
			wantErr: true,
		},
		{
			name:    "reject_loopback_ipv6",
			rawURL:  "http://[::1]",
			wantErr: true,
		},
		{
			name:    "reject_localhost",
			rawURL:  "http://localhost",
			wantErr: true,
		},
		{
			name:    "reject_private_ipv4",
			rawURL:  "http://10.0.0.1",
			wantErr: true,
		},
		{
			name:    "reject_link_local_metadata",
			rawURL:  "http://169.254.169.254",
			wantErr: true,
		},
		{
			name:    "reject_docker_host_alias",
			rawURL:  "http://host.docker.internal",
			wantErr: true,
		},
		{
			name:    "reject_private_172",
			rawURL:  "http://172.16.0.1",
			wantErr: true,
		},
		{
			name:    "reject_private_192",
			rawURL:  "http://192.168.1.1",
			wantErr: true,
		},
		{
			name:    "reject_url_with_userinfo",
			rawURL:  "http://admin:pass@example.com",
			wantErr: true,
		},
		{
			name:    "reject_empty_scheme",
			rawURL:  "://example.com",
			wantErr: true,
		},
		{
			name:    "reject_no_host",
			rawURL:  "http://",
			wantErr: true,
		},
		{
			name:    "reject_gateway_docker",
			rawURL:  "http://gateway.docker.internal",
			wantErr: true,
		},
		{
			name:    "reject_subdomain_localhost",
			rawURL:  "http://foo.localhost",
			wantErr: true,
		},
		{
			name:    "reject_ipv6_link_local",
			rawURL:  "http://[fe80::1]",
			wantErr: true,
		},
		{
			name:    "reject_ipv6_unique_local",
			rawURL:  "http://[fc00::1]",
			wantErr: true,
		},
		{
			name:    "allow_public_domain_with_path",
			rawURL:  "https://hooks.slack.com/services/T00/B00/xxx",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate(%q) error = %v, wantErr=%v", tt.rawURL, err, tt.wantErr)
			}
		})
	}
}

func TestSecureDialContext_blocks_loopback(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	dial := secureDialContext(2 * time.Second)
	conn, err := dial(context.Background(), "tcp", srv.Listener.Addr().String())
	if conn != nil {
		_ = conn.Close()
	}
	if err == nil {
		t.Fatal("expected secureDialContext to block loopback connection, but got nil error")
	}
}

func TestSecureDialContext_blocks_private_ip(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	dial := secureDialContext(2 * time.Second)
	conn, dialErr := dial(context.Background(), "tcp", ln.Addr().String())
	if conn != nil {
		_ = conn.Close()
	}
	if dialErr == nil {
		t.Fatal("expected secureDialContext to block private IP connection")
	}
}

func TestReadBodyWithLimit(t *testing.T) {
	t.Parallel()

	t.Run("allow_small_body", func(t *testing.T) {
		t.Parallel()
		body := bytes.Repeat([]byte("a"), int(defaultSafeResponseBodyLimitBytes-1))
		got, err := readBodyWithLimit(bytes.NewReader(body), defaultSafeResponseBodyLimitBytes)
		if err != nil {
			t.Fatalf("readBodyWithLimit unexpected error: %v", err)
		}
		if len(got) != len(body) {
			t.Fatalf("unexpected body len: got=%d want=%d", len(got), len(body))
		}
	})

	t.Run("reject_oversized_body", func(t *testing.T) {
		t.Parallel()
		body := bytes.Repeat([]byte("b"), int(defaultSafeResponseBodyLimitBytes+1))
		_, err := readBodyWithLimit(bytes.NewReader(body), defaultSafeResponseBodyLimitBytes)
		if err == nil {
			t.Fatal("expected oversize body error")
		}
	})
}
