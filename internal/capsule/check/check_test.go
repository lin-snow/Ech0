// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lin-snow/ech0/internal/capsule"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

const (
	echoID    = "01947c3e-0000-7000-8000-000000000001"
	commentID = "01947c3e-0000-7000-8000-000000000002"
	echoPath  = "echoes/2026/2026-01-01-01947c3e.md"
	catBytes  = "cat-bytes" // 9 字节，供 size 校验用
)

// buildCapsule 在临时目录里手搭一个胶囊；键是胶囊内相对路径。
func buildCapsule(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	return dir
}

func runCheck(t *testing.T, dir string, opts Options) *Report {
	t.Helper()
	src, err := capsule.Open(dir)
	if err != nil {
		t.Fatalf("open capsule: %v", err)
	}
	t.Cleanup(func() { _ = src.Close() })

	loaded, report, err := Run(context.Background(), src, opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if loaded == nil {
		t.Fatal("Run returned a nil capsule")
	}
	return report
}

// findIssue 按级别 + 坐标定位一条发现，找不到返回 nil。
func findIssue(report *Report, level Level, path, field string) *Issue {
	for i := range report.Issues {
		it := &report.Issues[i]
		if it.Level == level && it.Path == path && it.Field == field {
			return it
		}
	}
	return nil
}

func dumpIssues(report *Report) string {
	var b strings.Builder
	for _, it := range report.Issues {
		b.WriteString("\n  " + it.Level.String() + " " + it.Path + " [" + it.Field + "] " + it.Message)
	}
	return b.String()
}

// TestValidateCleanCapsule 锁定「合法胶囊零噪音」：校验器一旦对正常导出产物
// 报警告，用户就会开始无视报告，整套准入门槛随之失效。
func TestValidateCleanCapsule(t *testing.T) {
	dir := buildCapsule(t, map[string]string{
		capsule.ManifestPath: `schema_version: 1
generator: ech0 test
site:
  site_title: Demo
  server_url: https://demo.example
owner:
  username: alice
`,
		echoPath: `---
id: ` + echoID + `
created_at: 2026-01-01T00:00:00Z
layout: grid
files:
  - key: cat.png
    category: image
    size: 9
extension:
  type: WEBSITE
  payload:
    url: https://elsewhere.example/post
---
hello world
`,
		capsule.CommentsPath: `schema_version: 1
comments:
  - id: ` + commentID + `
    echo_id: ` + echoID + `
    nickname: bob
    content: nice
    status: approved
    created_at: 2026-01-02T03:04:05Z
`,
		"files/images/cat.png": catBytes,
	})

	report := runCheck(t, dir, Options{})
	if len(report.Issues) != 0 {
		t.Fatalf("合法胶囊不应产生任何发现，实际 %d 条:%s", len(report.Issues), dumpIssues(report))
	}
	if report.HasErrors() {
		t.Fatal("HasErrors 与空 Issues 自相矛盾")
	}
}

// TestFixGeneratesMissingID 锁定 spec §7 唯一的自动修复项：补 id 必须真的落盘，
// 且正文逐字不动——回写走的是整文件重写，正文一旦被动过就是数据损坏。
func TestFixGeneratesMissingID(t *testing.T) {
	const body = "line one\n\n  indented line\n中文 🎉\n"
	dir := buildCapsule(t, map[string]string{
		capsule.ManifestPath: `schema_version: 1
site:
  site_title: Demo
owner:
  username: alice
`,
		echoPath: `---
created_at: 2026-01-01T00:00:00Z
layout: grid
---
` + body,
	})

	report := runCheck(t, dir, Options{Fix: true})
	if report.HasErrors() {
		t.Fatalf("--fix 后不应残留 error:%s", dumpIssues(report))
	}
	if len(report.Fixed) != 1 {
		t.Fatalf("Fixed = %v, 期望 1 条修复记录", report.Fixed)
	}

	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(echoPath)))
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	doc, _, err := capsule.DecodeEcho(raw)
	if err != nil {
		t.Fatalf("decode back: %v", err)
	}
	if !uuidUtil.IsValid(doc.ID) {
		t.Fatalf("回写后的 id = %q，不是合法 UUID", doc.ID)
	}
	if !strings.Contains(report.Fixed[0], doc.ID) {
		t.Errorf("Fixed 记录 %q 未提及实际写入的 id %s", report.Fixed[0], doc.ID)
	}
	if doc.Content != body {
		t.Errorf("正文被改动:\n got %q\nwant %q", doc.Content, body)
	}
	if doc.Layout != "grid" || doc.CreatedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("其余 frontmatter 字段未原样保留: %+v", doc)
	}

	// 修复后重跑（不带 --fix）必须干净：修复不是「本次报告好看」而是真落地。
	if again := runCheck(t, dir, Options{}); again.HasErrors() {
		t.Fatalf("修复后重跑仍有 error:%s", dumpIssues(again))
	}
}

// TestValidateFileRefErrors 覆盖 files[] 的三类拦截项（spec §4.2 / §6）。
func TestValidateFileRefErrors(t *testing.T) {
	dir := buildCapsule(t, map[string]string{
		capsule.ManifestPath: `schema_version: 1
owner:
  username: alice
`,
		echoPath: `---
id: ` + echoID + `
created_at: 2026-01-01T00:00:00Z
files:
  - key: cat.png
    url: https://cdn.example/cat.png
  - key: ghost.png
  - url: https://cdn.example/doc.pdf
    category: document
---
body
`,
		"files/images/cat.png": catBytes,
	})

	report := runCheck(t, dir, Options{})
	if !report.HasErrors() {
		t.Fatal("期望拦截级发现，实际为零")
	}
	// category 不认得只是渲染退化，属警告；真正拦截的是 key/url 互斥违规与缺字节。
	if got := report.Count(LevelError); got != 2 {
		t.Fatalf("error 数 = %d，期望 2:%s", got, dumpIssues(report))
	}
	for _, field := range []string{"files[0]", "files[1].key"} {
		if findIssue(report, LevelError, echoPath, field) == nil {
			t.Errorf("缺少 %s 的 error:%s", field, dumpIssues(report))
		}
	}
	if findIssue(report, LevelWarning, echoPath, "files[2].category") == nil {
		t.Errorf("未知 category 应为 warning:%s", dumpIssues(report))
	}
	if it := findIssue(report, LevelError, echoPath, "files[1].key"); it != nil &&
		!strings.Contains(it.Message, "files/images/ghost.png") {
		t.Errorf("缺字节的报错应指出期望路径，实际 %q", it.Message)
	}
}

// TestValidatePrivacyAndWarnings 覆盖隐私红线（禁止字段即 error）与三类警告：
// 悬空媒体、custom_js、非 approved 评论。
func TestValidatePrivacyAndWarnings(t *testing.T) {
	const orphanMedia = "files/images/orphan.png"
	dir := buildCapsule(t, map[string]string{
		capsule.ManifestPath: `schema_version: 1
site:
  site_title: Demo
  custom_js: alert(1)
owner:
  username: alice
`,
		echoPath: `---
id: ` + echoID + `
created_at: 2026-01-01T00:00:00Z
---
body
`,
		capsule.CommentsPath: `schema_version: 1
comments:
  - id: ` + commentID + `
    echo_id: ` + echoID + `
    nickname: bob
    content: nice
    status: pending
    email: bob@example.com
    created_at: 2026-01-02T03:04:05Z
`,
		orphanMedia: "png-bytes",
	})

	report := runCheck(t, dir, Options{})

	if got := report.Count(LevelError); got != 1 {
		t.Fatalf("error 数 = %d，期望仅禁止字段 1 条:%s", got, dumpIssues(report))
	}
	if findIssue(report, LevelError, capsule.CommentsPath, "comments[0].email") == nil {
		t.Errorf("comments.yaml 的 email 必须是 error:%s", dumpIssues(report))
	}
	for _, want := range []struct{ path, field string }{
		{orphanMedia, ""},
		{capsule.ManifestPath, "site.custom_js"},
		{capsule.CommentsPath, "comments[0].status"},
	} {
		if findIssue(report, LevelWarning, want.path, want.field) == nil {
			t.Errorf("缺少 %s [%s] 的 warning:%s", want.path, want.field, dumpIssues(report))
		}
	}

	// 排序契约：error 一律排在 warning 之前。
	for i := 1; i < len(report.Issues); i++ {
		if report.Issues[i-1].Level > report.Issues[i].Level {
			t.Fatalf("Issues 未按级别排序:%s", dumpIssues(report))
		}
	}
}

// TestValidateManifestAndOrphans 覆盖清单级拦截项与孤儿评论、内嵌实例 URL 警告。
func TestValidateManifestAndOrphans(t *testing.T) {
	dir := buildCapsule(t, map[string]string{
		capsule.ManifestPath: `schema_version: 2
site:
  server_url: https://demo.example
owner: {}
`,
		echoPath: `---
id: not-a-uuid
created_at: 2026-01-01
layout: mosaic
extension:
  type: MUSIC
---
see https://demo.example/echo/1
`,
		capsule.CommentsPath: `schema_version: 1
comments:
  - id: ` + commentID + `
    echo_id: ` + echoID + `
    nickname: bob
    content: nice
    created_at: 2026-01-02T03:04:05Z
`,
	})

	report := runCheck(t, dir, Options{})

	for _, want := range []struct{ path, field string }{
		{capsule.ManifestPath, "schema_version"}, // 高于自身支持必须拒绝
		{capsule.ManifestPath, "owner.username"}, // 归属兜底缺失
		{echoPath, "id"},                         // 非法 UUID
		{echoPath, "created_at"},                 // 非 RFC3339
		{echoPath, "extension.payload"},          // 有 extension 必须有 payload
	} {
		if findIssue(report, LevelError, want.path, want.field) == nil {
			t.Errorf("缺少 %s [%s] 的 error:%s", want.path, want.field, dumpIssues(report))
		}
	}
	// 表现层枚举退化成警告：内容完好，消费者回落默认值即可，不该拦下整个胶囊。
	if findIssue(report, LevelWarning, echoPath, "layout") == nil {
		t.Errorf("未知 layout 应为 warning:%s", dumpIssues(report))
	}
	if findIssue(report, LevelWarning, capsule.CommentsPath, "comments[0].echo_id") == nil {
		t.Errorf("孤儿评论应为 warning:%s", dumpIssues(report))
	}
	if findIssue(report, LevelWarning, echoPath, "content") == nil {
		t.Errorf("正文内嵌实例 URL 应为 warning:%s", dumpIssues(report))
	}
}

// TestFixRejectsArchiveCapsule 锁定 --fix 的前置门：zip 无法就地改写，
// 与其改一半不如在动手前拒绝。
func TestFixRejectsArchiveCapsule(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "capsule.zip")
	if err := os.WriteFile(archive, []byte("PK-not-really"), 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	loaded := &capsule.Loaded{Source: &capsule.Source{Path: archive}}

	if _, err := Validate(context.Background(), loaded, Options{Fix: true}); err == nil {
		t.Fatal("对 zip 形态胶囊使用 --fix 必须报错")
	}
	if _, err := Validate(context.Background(), loaded, Options{}); err != nil {
		t.Fatalf("只读校验不应因形态失败: %v", err)
	}
}
