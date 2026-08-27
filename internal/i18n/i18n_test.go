// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package i18n

import (
	"strings"
	"testing"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

func TestSanitizeAcceptLanguage(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"legit browser header", "en-US,en;q=0.9,zh-CN;q=0.8", "en-US,en;q=0.9,zh-CN;q=0.8"},
		{"extended subtag", "zh-Hant-TW", "zh-Hant-TW"},
		{"at cap", strings.Repeat("a-", maxAcceptLanguageSeparators), strings.Repeat("a-", maxAcceptLanguageSeparators)},
		{"over cap hyphen", strings.Repeat("a-", maxAcceptLanguageSeparators+1), ""},
		{"underscore bypass (GHSA-mqxv-9rm6-w8qc)", strings.Repeat("_", 10_000), ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitizeAcceptLanguage(tc.in); got != tc.want {
				t.Fatalf("sanitizeAcceptLanguage(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestResolveLocaleLegit(t *testing.T) {
	if got := ResolveLocale("en-US,en;q=0.9,zh-CN;q=0.8"); got != "en-US" {
		t.Fatalf("ResolveLocale legit = %q, want en-US", got)
	}
}

func TestResolveLocaleDefangsQuadraticParse(t *testing.T) {
	malicious := strings.Repeat("_", 1<<20)

	start := time.Now()
	got := ResolveLocale(malicious)
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("ResolveLocale took %v on hostile input; quadratic guard regressed", elapsed)
	}
	if got != string(commonModel.FallbackLocale) {
		t.Fatalf("ResolveLocale(malicious) = %q, want fallback %q", got, commonModel.FallbackLocale)
	}
}
