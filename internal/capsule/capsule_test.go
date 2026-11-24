// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"reflect"
	"strings"
	"testing"
	"time"

	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

func TestSiteKeysMatchSystemSetting(t *testing.T) {
	// Behaviour switches stay out of a capsule: importing content must not flip
	// how the receiving instance behaves.
	behaviourOnly := []string{"allow_register", "enable_agent_discovery"}
	// Input-only fields are not content either: server_logo_file_id addresses an
	// upload that the importer re-links from file_links, so a capsule never
	// carries it.
	inputOnly := []string{"server_logo_file_id"}

	want := tagSet(t, reflect.TypeFor[settingModel.SystemSetting](), "json")
	for _, key := range append(behaviourOnly, inputOnly...) {
		delete(want, key)
	}
	got := tagSet(t, reflect.TypeFor[Site](), "yaml")

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("capsule.Site 键集合与 SystemSetting 不一致\nSystemSetting(-行为开关): %v\nSite: %v", keys(want), keys(got))
	}
}

func TestFileRefIsThePortableProjection(t *testing.T) {
	want := map[string]struct{}{
		"id": {}, "key": {}, "url": {}, "name": {}, "content_type": {},
		"size": {}, "width": {}, "height": {}, "category": {},
	}
	got := tagSet(t, reflect.TypeFor[FileRef](), "yaml")
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("FileRef 键集合变了\nwant: %v\ngot:  %v", keys(want), keys(got))
	}

	for _, runtimeOnly := range []string{"storage_id", "object_key", "blob_bytes", "kind", "status", "access"} {
		if _, ok := got[runtimeOnly]; ok {
			t.Errorf("FileRef.%s 是运行时拓扑列，禁止入胶囊", runtimeOnly)
		}
	}
}

func tagSet(t *testing.T, typ reflect.Type, tag string) map[string]struct{} {
	t.Helper()
	out := make(map[string]struct{}, typ.NumField())
	for field := range typ.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get(tag), ",")
		if name == "" || name == "-" {
			continue
		}
		out[name] = struct{}{}
	}
	return out
}

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestEncodeDecodeEchoPreservesContent(t *testing.T) {
	cases := map[string]string{
		"empty":            "",
		"plain":            "hello world",
		"trailing newline": "line one\nline two\n",
		"leading newline":  "\n缩进保留\n",
		"inner fence":      "before\n---\nafter\n",
		"crlf body":        "a\r\nb\r\n",
		"unicode":          "中文 🎉 emoji\n",
	}

	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			in := &EchoDoc{
				ID:        "0198f0a0-0000-7000-8000-000000000001",
				CreatedAt: "2026-08-02T03:04:05Z",
				Username:  "l1nsn0w",
				Tags:      []string{"日记", "ech0"},
				Layout:    "grid",
				Private:   true,
				FavCount:  7,
				Content:   content,
			}
			raw, err := EncodeEcho(in)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			out, unknown, err := DecodeEcho(raw)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if len(unknown) != 0 {
				t.Fatalf("unexpected unknown fields: %v", unknown)
			}
			if out.Content != content {
				t.Fatalf("content not byte-identical\nwant %q\ngot  %q", content, out.Content)
			}
			in.Content = content
			if !reflect.DeepEqual(in, out) {
				t.Fatalf("frontmatter round-trip mismatch\nwant %+v\ngot  %+v", in, out)
			}
		})
	}
}

func TestDecodeEchoReportsUnknownFields(t *testing.T) {
	raw := []byte("---\nid: x\ncreated_at: 2026-01-01T00:00:00Z\nfuture_field: 42\n---\nbody\n")
	doc, unknown, err := DecodeEcho(raw)
	if err != nil {
		t.Fatalf("decode must tolerate unknown fields: %v", err)
	}
	if len(unknown) != 1 || !strings.Contains(unknown[0], "future_field") {
		t.Fatalf("expected future_field reported as unknown, got %v", unknown)
	}
	if doc.Content != "body\n" || doc.ID != "x" {
		t.Fatalf("unexpected doc: %+v", doc)
	}
}

func TestDecodeEchoRejectsTypeErrors(t *testing.T) {
	raw := []byte("---\nid: x\nfav_count: not-a-number\n---\n")
	if _, _, err := DecodeEcho(raw); err == nil {
		t.Fatal("expected a type error for fav_count")
	}
}

func TestDecodeEchoRequiresFrontmatter(t *testing.T) {
	if _, _, err := DecodeEcho([]byte("no fence here\n")); err == nil {
		t.Fatal("expected an error when the leading fence is missing")
	}
}

func TestMediaPath(t *testing.T) {
	cases := map[string]string{
		"2026/11/a.png":           "files/2026/11/a.png",
		"2026/11/clip.mp4":        "files/2026/11/clip.mp4",
		"blob/0193f1a1-aaaa":      "files/blob/0193f1a1-aaaa",
		"2026/11/UPPER.PNG":       "files/2026/11/UPPER.PNG",
		"2026/11/noext":           "files/2026/11/noext",
		"2026/11/deep/nested.bin": "files/2026/11/deep/nested.bin",
	}
	for key, want := range cases {
		if got := MediaPath(key); got != want {
			t.Errorf("MediaPath(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestValidateKey(t *testing.T) {
	for _, bad := range []string{"", "../escape.png", "/abs.png", `win\path.png`, "a/../b.png", "./a.png"} {
		if err := ValidateKey(bad); err == nil {
			t.Errorf("ValidateKey(%q) should fail", bad)
		}
	}
	for _, good := range []string{"2026/11/0193f1a1-aaaa.png", "blob/0193f1a1-aaaa", "a.png"} {
		if err := ValidateKey(good); err != nil {
			t.Errorf("ValidateKey(%q) rejected a clean relative key: %v", good, err)
		}
	}
}

func TestTimeRoundTrip(t *testing.T) {
	const sec int64 = 1_770_000_000
	formatted := FormatUnix(sec)
	if !strings.HasSuffix(formatted, "Z") {
		t.Fatalf("export must normalise to UTC, got %q", formatted)
	}
	back, err := ParseTime(formatted)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if back != sec {
		t.Fatalf("round-trip drift: %d -> %q -> %d", sec, formatted, back)
	}

	offset, err := ParseTime("2026-02-01T08:00:00+08:00")
	if err != nil {
		t.Fatalf("parse offset: %v", err)
	}
	utc, err := ParseTime("2026-02-01T00:00:00Z")
	if err != nil {
		t.Fatalf("parse utc: %v", err)
	}
	if offset != utc {
		t.Fatalf("offset and UTC forms must denote the same instant: %d vs %d", offset, utc)
	}

	if _, err := ParseTime("2026-02-01"); err == nil {
		t.Fatal("a bare date is not RFC3339 and must be rejected")
	}
}

func TestEchoPath(t *testing.T) {
	got := EchoPath("0198f0a0-1111-7000-8000-000000000001", mustTime(t, "2026-08-02T23:30:00+08:00"))
	if want := "echoes/2026/2026-08-02-00000001.md"; got != want {
		t.Fatalf("EchoPath = %q, want %q", got, want)
	}
	if got := EchoPath("short", mustTime(t, "2026-01-01T00:00:00Z")); got != "echoes/2026/2026-01-01-short.md" {
		t.Fatalf("short id must not panic, got %q", got)
	}
}

func mustTime(t *testing.T, rfc3339 string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		t.Fatalf("bad fixture time %q: %v", rfc3339, err)
	}
	return parsed
}
