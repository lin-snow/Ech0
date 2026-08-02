// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package build 把一个 Capsule 烘焙成可静态托管的只读站点（spec §10）。
//
// 这是「导出即转储，构建即转换」里的**转换**端：所有面向消费者的派生
// ——排序、聚合、URL 计算、默认值填充、RSS/sitemap 渲染——都只发生在本包，
// 胶囊本身保持 1:1 转储语义。
//
// 产物不需要任何后端：同一份内嵌 SPA bundle 在运行时读 window.__ECH0_STATIC__
// 开关，把请求层换成浏览器内的 dataset.json 查询引擎。因此 build 不跑 Node，
// 也不生成「假 API 文件」（首屏的 POST /echo/query 无法用静态文件冒充）。
//
// 校验由 CLI 统一前置（check.Run → 有 error 则拒绝 → 再调 build.Run），
// 本包刻意不 import check，避免两个消费者互相依赖。
package build

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/capsule"
)

// Options 是一次烘焙的输入参数。
type Options struct {
	// Output 是产物目录；必须不存在或为空目录。
	Output string
	// BaseURL 是站点部署的子路径前缀（如 /blog/）。空值等价于根部署。
	BaseURL string
}

// Result 是一次烘焙的产出摘要。
type Result struct {
	Path     string
	Echoes   int
	Files    int
	Comments int
}

// Run 执行烘焙：拷贝内嵌 SPA、产出 dataset.json 与各类静态端点、
// 铺开媒体字节、改写入口 HTML。
func Run(ctx context.Context, loaded *capsule.Loaded, opts Options) (*Result, error) {
	if loaded == nil {
		return nil, fmt.Errorf("capsule is not loaded")
	}
	if strings.TrimSpace(opts.Output) == "" {
		return nil, fmt.Errorf("output directory is required")
	}
	dir := filepath.Clean(opts.Output)
	baseURL := NormalizeBaseURL(opts.BaseURL)

	if err := ensureEmptyDir(dir); err != nil {
		return nil, err
	}

	ds, err := bake(bakeInput{loaded: loaded, baseURL: baseURL, generatedAt: time.Now()})
	if err != nil {
		return nil, err
	}

	if err := copySPA(dir); err != nil {
		return nil, err
	}
	if err := writeEntrypoints(dir, baseURL); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(ds)
	if err != nil {
		return nil, fmt.Errorf("marshal dataset: %w", err)
	}
	if err := writeFile(dir, "dataset.json", payload); err != nil {
		return nil, err
	}

	// api/connect 是无扩展名文件（spec §10 硬要求）：远端实例的既有探测路径
	// 直接 GET <site>/api/connect 就能拿到与活实例同形的响应体。
	envelope, err := json.Marshal(resultEnvelope{Code: 1, Msg: "", Data: ds.Connect})
	if err != nil {
		return nil, fmt.Errorf("marshal connect payload: %w", err)
	}
	if err := writeFile(dir, "api/connect", envelope); err != nil {
		return nil, err
	}

	files, err := copyMedia(ctx, dir, loaded)
	if err != nil {
		return nil, err
	}

	l := newLinks(ds.Settings.ServerURL, baseURL)
	generatedAt := time.Unix(ds.GeneratedAt, 0).UTC()

	atom, err := renderAtom(ds, l, generatedAt)
	if err != nil {
		return nil, fmt.Errorf("render rss: %w", err)
	}
	if err := writeFile(dir, "rss.xml", []byte(atom)); err != nil {
		return nil, err
	}

	sitemap, err := renderSitemap(ds, l, generatedAt)
	if err != nil {
		return nil, err
	}
	if err := writeFile(dir, "sitemap.xml", sitemap); err != nil {
		return nil, err
	}

	return &Result{
		Path:     dir,
		Echoes:   len(ds.Echos),
		Files:    files,
		Comments: len(ds.Comments),
	}, nil
}

// NormalizeBaseURL 归一化部署基址：恒以 / 开头且以 / 结尾。
// 前端 adapter 与 index.html 注入都直接字符串拼接它，这里定死形状，
// 下游就不必再各自防御 "blog" / "/blog" / "/blog/" 三种写法。
func NormalizeBaseURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "/"
	}
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}
