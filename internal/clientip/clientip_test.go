// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package clientip

import (
	"context"
	"net/http"
	"testing"
)

func TestResolve(t *testing.T) {
	cases := []struct {
		name    string
		peer    string
		headers map[string]string
		want    string
	}{
		{
			name:    "public peer ignores forwarded headers",
			peer:    "203.0.113.10:5555",
			headers: map[string]string{"X-Forwarded-For": "198.51.100.1"},
			want:    "203.0.113.10",
		},
		{
			name:    "loopback peer trusts rightmost public hop",
			peer:    "127.0.0.1:9000",
			headers: map[string]string{"X-Forwarded-For": "203.0.113.9, 10.0.0.1"},
			want:    "203.0.113.9",
		},
		{
			name:    "docker gateway peer",
			peer:    "172.17.0.1:9000",
			headers: map[string]string{"X-Forwarded-For": "198.51.100.4"},
			want:    "198.51.100.4",
		},
		{
			name:    "skips private hops",
			peer:    "10.1.2.3:9000",
			headers: map[string]string{"X-Forwarded-For": "198.51.100.4, 192.168.1.5, 10.0.0.2"},
			want:    "198.51.100.4",
		},
		{
			name:    "malformed entry stops the walk",
			peer:    "127.0.0.1:9000",
			headers: map[string]string{"X-Forwarded-For": "198.51.100.4, not-an-ip"},
			want:    "127.0.0.1",
		},
		{
			name: "x-real-ip when forwarded chain is internal",
			peer: "127.0.0.1:9000",
			headers: map[string]string{
				"X-Forwarded-For": "10.0.0.1",
				"X-Real-Ip":       "198.51.100.7",
			},
			want: "198.51.100.7",
		},
		{
			name: "x-real-ip is ignored when not usable",
			peer: "127.0.0.1:9000",
			headers: map[string]string{
				"X-Real-Ip": "10.0.0.7",
			},
			want: "127.0.0.1",
		},
		{
			name: "no headers falls back to peer",
			peer: "127.0.0.1:9000",
			want: "127.0.0.1",
		},
		{
			name: "unparsable peer is returned as-is",
			peer: "not-an-ip",
			want: "not-an-ip",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			header := http.Header{}
			for k, v := range tc.headers {
				header.Set(k, v)
			}
			if got := Resolve(tc.peer, header); got != tc.want {
				t.Fatalf("Resolve() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestContextRoundTrip(t *testing.T) {
	if got := FromContext(context.Background()); got != "" {
		t.Fatalf("FromContext(background) = %q, want empty", got)
	}

	ctx := WithClientIP(context.Background(), "198.51.100.7")
	if got := FromContext(ctx); got != "198.51.100.7" {
		t.Fatalf("FromContext() = %q, want 198.51.100.7", got)
	}
}
