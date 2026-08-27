// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"errors"
	"strings"
	"testing"
)

func TestCleanKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
		isErr bool
	}{
		{"", "", false},
		{"/", "", false},
		{"a/b/c", "a/b/c", false},
		{"/a/b/c/", "a/b/c", false},
		{"a//b///c", "a/b/c", false},
		{"a/./b", "a/b", false},
		{"..", "", true},
		{"a/../../etc", "", true},
		{"a/../b", "b", false},
	}
	for _, tt := range tests {
		got, err := CleanKey(tt.input)
		if tt.isErr {
			if err == nil {
				t.Errorf("CleanKey(%q) expected error, got %q", tt.input, got)
			} else if !errors.Is(err, ErrInvalidKey) {
				t.Errorf("CleanKey(%q) error should wrap ErrInvalidKey, got %v", tt.input, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("CleanKey(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("CleanKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func FuzzCleanKey(f *testing.F) {
	f.Add("")
	f.Add("/")
	f.Add("a/b/c")
	f.Add("../etc/passwd")
	f.Add("a/../../etc")
	f.Add("a//b///c")
	f.Add("a/./b/../c")
	f.Add(strings.Repeat("a/", 100))
	f.Add("hello\x00world")

	f.Fuzz(func(t *testing.T, input string) {
		cleaned, err := CleanKey(input)
		if err != nil {
			if !errors.Is(err, ErrInvalidKey) {
				t.Errorf("CleanKey(%q) error should wrap ErrInvalidKey, got %v", input, err)
			}
			return
		}

		if cleaned != "" && (cleaned[0] == '/' || cleaned[len(cleaned)-1] == '/') {
			t.Errorf("CleanKey(%q) = %q has leading/trailing slash", input, cleaned)
		}

		if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") || strings.HasSuffix(cleaned, "/..") {
			t.Errorf("CleanKey(%q) = %q contains path traversal", input, cleaned)
		}

		if strings.Contains(cleaned, "//") {
			t.Errorf("CleanKey(%q) = %q contains double slash", input, cleaned)
		}

		again, err := CleanKey(cleaned)
		if err != nil {
			t.Errorf("CleanKey(%q) succeeded, but CleanKey(%q) failed: %v", input, cleaned, err)
		}
		if again != cleaned {
			t.Errorf("CleanKey not idempotent: CleanKey(%q) = %q, CleanKey(%q) = %q", input, cleaned, cleaned, again)
		}
	})
}
