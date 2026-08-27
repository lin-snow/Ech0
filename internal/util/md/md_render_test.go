// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"strings"
	"testing"
)

func TestMdToHTML(t *testing.T) {
	t.Run("strips raw script tag", func(t *testing.T) {
		out := string(MdToHTML([]byte("<script>alert(1)</script>\n\nhello\n")))
		if strings.Contains(out, "<script") || strings.Contains(out, "alert(1)") {
			t.Fatalf("raw <script> should be stripped, got: %q", out)
		}
		if !strings.Contains(out, "hello") {
			t.Fatalf("surrounding text should survive, got: %q", out)
		}
	})

	t.Run("renders bold emphasis", func(t *testing.T) {
		out := string(MdToHTML([]byte("normal **bold** text\n")))
		if !strings.Contains(out, "<strong>bold</strong>") {
			t.Fatalf("expected bold rendering, got: %q", out)
		}
	})

	t.Run("autolinks bare url", func(t *testing.T) {
		out := string(MdToHTML([]byte("Visit https://example.com now\n")))
		if !strings.Contains(out, `href="https://example.com"`) {
			t.Fatalf("expected autolink href, got: %q", out)
		}
	})

	t.Run("renders github flavored table", func(t *testing.T) {
		out := string(MdToHTML([]byte("| a | b |\n| - | - |\n| 1 | 2 |\n")))
		for _, want := range []string{"<table>", "<thead>", "<th>a</th>", "<td>1</td>"} {
			if !strings.Contains(out, want) {
				t.Fatalf("expected table fragment %q, got: %q", want, out)
			}
		}
	})

	t.Run("links carry safe rel and target attributes", func(t *testing.T) {
		out := string(MdToHTML([]byte("[link](https://example.com)\n")))
		if !strings.Contains(out, `target="_blank"`) {
			t.Fatalf("expected target=_blank, got: %q", out)
		}
		if !strings.Contains(out, "noopener") || !strings.Contains(out, "noreferrer") {
			t.Fatalf("expected rel to contain noopener and noreferrer, got: %q", out)
		}
		if !strings.Contains(out, "rel=") {
			t.Fatalf("expected a rel attribute, got: %q", out)
		}
	})

	t.Run("empty input yields empty output", func(t *testing.T) {
		if out := string(MdToHTML([]byte{})); strings.TrimSpace(out) != "" {
			t.Fatalf("expected empty output for empty input, got: %q", out)
		}
	})
}
