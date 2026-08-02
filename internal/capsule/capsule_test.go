// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"reflect"
	"strings"
	"testing"
	"time"

	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

// TestSiteKeysMatchSystemSetting 守卫「导出即转储」最脆弱的一环：Site 的 yaml 键名
// 必须逐字等于 SystemSetting 的 json tag，只允许缺少行为开关 allow_register。
// SystemSetting 新增字段而这里没跟上时，本用例失败——否则胶囊会静默漏字段。
func TestSiteKeysMatchSystemSetting(t *testing.T) {
	const behaviourOnly = "allow_register"

	want := tagSet(t, reflect.TypeOf(settingModel.SystemSetting{}), "json")
	delete(want, behaviourOnly)
	got := tagSet(t, reflect.TypeOf(Site{}), "yaml")

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("capsule.Site 键集合与 SystemSetting 不一致\nSystemSetting(-allow_register): %v\nSite: %v", keys(want), keys(got))
	}
}

// TestFileRefKeysAreFileColumns 守卫 files[] 与 File 列的 1:1 关系：胶囊里的每个键
// 都必须是真实存在的 File 列，且被明确排除的运行时拓扑列不得混进来。
func TestFileRefKeysAreFileColumns(t *testing.T) {
	fileCols := tagSet(t, reflect.TypeOf(fileModel.File{}), "json")
	excluded := map[string]struct{}{
		"storage_type": {}, "provider": {}, "bucket": {}, "user_id": {}, "created_at": {},
	}

	for k := range tagSet(t, reflect.TypeOf(FileRef{}), "yaml") {
		if _, ok := fileCols[k]; !ok {
			t.Errorf("FileRef.%s 不是 File 的列", k)
		}
		if _, bad := excluded[k]; bad {
			t.Errorf("FileRef.%s 是运行时拓扑列，禁止入胶囊", k)
		}
	}
}

func tagSet(t *testing.T, typ reflect.Type, tag string) map[string]struct{} {
	t.Helper()
	out := make(map[string]struct{}, typ.NumField())
	for i := range typ.NumField() {
		name, _, _ := strings.Cut(typ.Field(i).Tag.Get(tag), ",")
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

// TestEncodeDecodeEchoPreservesContent 锁定 spec §4.3：正文必须逐字往返，
// 包括空正文、首尾换行，以及正文里自带 --- 围栏的情况。
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

// TestDecodeEchoReportsUnknownFields 锁定 spec §8：未知字段不得让解析失败，
// 但必须能被报出来（check 的警告来源）。
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

// TestDecodeEchoRejectsTypeErrors 区分「未知字段」（警告）与「类型错」（硬错）。
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

// TestMediaPath 锁定 spec §6：胶囊内位置是 key 的纯函数，与实例本地布局同构。
func TestMediaPath(t *testing.T) {
	cases := map[string]string{
		"a.png":     "files/images/a.png",
		"clip.mp4":  "files/videos/clip.mp4",
		"song.mp3":  "files/audios/song.mp3",
		"paper.pdf": "files/documents/paper.pdf",
		"notes.txt": "files/files/notes.txt",
		"noext":     "files/files/noext",
		"UPPER.PNG": "files/images/UPPER.PNG",
	}
	for key, want := range cases {
		if got := MediaPath(key); got != want {
			t.Errorf("MediaPath(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestValidateKey(t *testing.T) {
	for _, bad := range []string{"", "sub/dir.png", "../escape.png", "a/../b.png", `win\path.png`} {
		if err := ValidateKey(bad); err == nil {
			t.Errorf("ValidateKey(%q) should fail", bad)
		}
	}
	if err := ValidateKey("ok_1735689600_deadbeef.png"); err != nil {
		t.Errorf("ValidateKey rejected a legit flat key: %v", err)
	}
}

// TestTimeRoundTrip 锁定 spec §11 的无损双射：Unix 秒 → RFC3339 UTC → Unix 秒。
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

	// 输入侧接受任意合法偏移，语义是时刻。
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
	// 2026-08-02T23:30+08:00 == 2026-08-02T15:30Z，年份与日期一律按 UTC 归档。
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
