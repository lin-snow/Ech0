// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package url

import "testing"

func TestTrimURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty stays empty", in: "", want: ""},
		{name: "no decoration unchanged", in: "example.com", want: "example.com"},
		{name: "strips surrounding whitespace", in: "  example.com  ", want: "example.com"},
		{name: "strips single leading slash", in: "/api", want: "api"},
		{name: "strips single trailing slash", in: "api/", want: "api"},
		{name: "strips one slash each side", in: "/api/", want: "api"},
		{name: "only strips one layer of slashes", in: "//x//", want: "/x/"},
		{name: "whitespace then slashes", in: "  /api/  ", want: "api"},
		{name: "keeps inner slashes", in: "/a/b/c/", want: "a/b/c"},
		{name: "full url preserved", in: "https://example.com/", want: "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrimURL(tt.in); got != tt.want {
				t.Fatalf("TrimURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
