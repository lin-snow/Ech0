// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lin-snow/ech0/internal/capsule"
)

// spaRoot 是内嵌 SPA 产物在 template.WebFS 里的根目录。
const spaRoot = "dist"

// indexFile 是 SPA 入口；404.html 是它的副本，用于 Pages 类托管的深链兜底。
const (
	indexFile    = "index.html"
	notFoundFile = "404.html"
)

// ensureEmptyDir 保证输出目录存在且为空。拒绝往非空目录里写是刻意的：
// 增量覆盖会留下上一次构建的陈旧资源（改名后的 assets 尤其致命），
// 与其产出一个「半新半旧」的站点，不如让调用方显式清空。
func ensureEmptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	switch {
	case err == nil:
		if len(entries) > 0 {
			return fmt.Errorf("output directory %s is not empty", dir)
		}
		return nil
	case os.IsNotExist(err):
		return os.MkdirAll(dir, 0o755)
	default:
		return fmt.Errorf("inspect output directory %s: %w", dir, err)
	}
}

// writeFile 写入一个产物文件，按需创建父目录。
func writeFile(dir, rel string, data []byte) error {
	dest := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", rel, err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", rel, err)
	}
	return nil
}

// copySPA 把 SPA 产物整棵树原样铺到输出目录。走 all: 嵌入的全部条目
// （含 _plugin-vue_export-helper-*.js 这类下划线开头的文件），漏一个就白屏。
//
// assets 作为参数而非直接取 template.WebFS：内嵌产物由 `pnpm build` 生成、
// 不进版本库，CI 只塞一个占位 index.html。测试若读真产物就只能在「本地恰好
// 构建过前端」时通过，等于把测试结果押在工作区状态上。
func copySPA(dir string, assets fs.FS) error {
	return fs.WalkDir(assets, spaRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := fs.ReadFile(assets, p)
		if readErr != nil {
			return fmt.Errorf("read embedded %s: %w", p, readErr)
		}
		rel := strings.TrimPrefix(p, spaRoot+"/")
		return writeFile(dir, rel, data)
	})
}

// copyMedia 把胶囊 files/ 下的字节铺到 api/files/。位置与 serve 模式的静态
// 路由逐字一致，所以 dataset 里算出来的 URL 在两种模式下同形。
func copyMedia(ctx context.Context, dir string, loaded *capsule.Loaded) (int, error) {
	// map 遍历无序，排一下让失败信息与产出顺序可复现。
	keys := make([]string, 0, len(loaded.MediaPaths))
	for p := range loaded.MediaPaths {
		keys = append(keys, p)
	}
	sort.Strings(keys)

	prefix := capsule.FilesDir + "/"
	n := 0
	for _, p := range keys {
		rel := strings.TrimPrefix(p, prefix)
		if rel == p || rel == "" || rel != path.Clean(rel) || strings.HasPrefix(rel, "../") {
			return n, fmt.Errorf("unexpected media path %q in capsule", p)
		}
		data, err := loaded.Source.ReadFile(ctx, p)
		if err != nil {
			return n, fmt.Errorf("read media %s: %w", p, err)
		}
		if err := writeFile(dir, "api/files/"+rel, data); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// rootPathAttr 匹配指向站点根的资源属性（href="/assets/…"、src="/favicon.svg" 等）。
// 只吃 href / src，且要求 / 之后不是另一个 /——协议相对 URL（//cdn/…）必须放过。
var rootPathAttr = regexp.MustCompile(`\b(href|src)="/([^/"])`)

// rootPathAttrBare 匹配恰好指向根的 href="/"（首页链接）。
var rootPathAttrBare = regexp.MustCompile(`\b(href|src)="/"`)

// firstScriptTag 定位第一个 <script，注入点必须在它之前：SPA 的入口脚本一旦
// 开跑就会读 window.__ECH0_STATIC__，开关晚到等于没有。
var firstScriptTag = regexp.MustCompile(`<script[\s>]`)

// renderIndex 产出改写后的 index.html：注入静态模式开关，并把绝对根路径
// 迁到部署基址下（baseURL 为 / 时不动，避免无谓的字节变化）。
func renderIndex(raw []byte, baseURL string) []byte {
	html := string(raw)

	if baseURL != "/" {
		html = rootPathAttrBare.ReplaceAllString(html, `$1="`+baseURL+`"`)
		html = rootPathAttr.ReplaceAllString(html, `$1="`+baseURL+`$2`)
	}

	// SPA 头部的 <link rel="alternate"> 指向 serve 模式的动态路由 /rss；
	// 静态站上那个路径不存在，改指真正落盘的 rss.xml，否则订阅入口是死链。
	html = strings.Replace(html, `href="`+baseURL+`rss"`, `href="`+baseURL+`rss.xml"`, 1)

	// 开关值走 JSON 编码，baseURL 里的引号 / 反斜杠不会撕开脚本。
	snippet := fmt.Sprintf(
		"<script>window.__ECH0_STATIC__=true;window.__ECH0_STATIC_BASE__=%s;</script>\n    ",
		jsonString(baseURL),
	)

	if loc := firstScriptTag.FindStringIndex(html); loc != nil {
		return []byte(html[:loc[0]] + snippet + html[loc[0]:])
	}
	if i := strings.Index(html, "</head>"); i >= 0 {
		return []byte(html[:i] + snippet + html[i:])
	}
	// 连 </head> 都没有的 index.html 不该存在；真遇上就前置，总比丢开关强。
	return []byte(snippet + html)
}

// jsonString 把字符串编码成 JS 字面量。手写而非 json.Marshal：这里只需要
// 一个不会失败的最小转义，且要顺手挡掉 </script 提前闭合。
func jsonString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '<':
			b.WriteString(`\u003c`)
		case '>':
			b.WriteString(`\u003e`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// writeEntrypoints 改写 index.html 并复制出 404.html。
func writeEntrypoints(dir, baseURL string, assets fs.FS) error {
	raw, err := fs.ReadFile(assets, spaRoot+"/"+indexFile)
	if err != nil {
		return fmt.Errorf("read embedded %s: %w", indexFile, err)
	}
	rendered := renderIndex(raw, baseURL)
	if err := writeFile(dir, indexFile, rendered); err != nil {
		return err
	}
	// 404.html 必须与 index.html 逐字一致：它就是深链回退到的同一个 SPA。
	return writeFile(dir, notFoundFile, rendered)
}
