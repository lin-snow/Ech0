// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	htmltemplate "html/template"
	"strings"
)

// RSSDocument is one feed render in both wire forms: Atom for feed readers and
// a subscription landing page for browsers.
//
// The browser form is rendered on the server because Chromium — and therefore
// Microsoft Edge in v159 — is removing client-side XSLT; an XML document that
// only looked good through <?xml-stylesheet?> degrades to a raw XML dump once
// that lands. The page carries only the subscription affordance (feed URL) and
// feed liveness (count, last update); entry bodies live in the Atom document
// alone so the landing page never duplicates feed content.
type RSSDocument struct {
	Atom string
	HTML string
}

type rssPage struct {
	Title      string
	Logo       string
	Home       string
	FeedURL    string
	Count      int
	Updated    string
	UpdatedISO string
}

var rssPageTemplate = htmltemplate.Must(htmltemplate.New("rss-page").Parse(rssPageSource))

func renderRSSPage(page rssPage) (string, error) {
	var b strings.Builder
	if err := rssPageTemplate.Execute(&b, page); err != nil {
		return "", err
	}
	return b.String(), nil
}

const rssPageSource = `<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <meta name="robots" content="noindex,follow">
    <title>{{.Title}} · RSS</title>
    <link rel="icon" href="/favicon.svg" type="image/svg+xml">
    <link rel="alternate" type="application/atom+xml" href="{{.FeedURL}}" title="{{.Title}}">
    <style>
      :root {
        --bg-canvas: #f4f1ec;
        --bg-surface: #fff;
        --bg-muted: #fdf9f4;
        --text-primary: #33251d;
        --text-secondary: #5b4f46;
        --text-muted: #8e847d;
        --border-subtle: #e4dfdb;
        --accent: #b84200;
        --accent-soft: #f4e3cd;
      }
      @media (prefers-color-scheme: dark) {
        :root {
          --bg-canvas: #1a1816;
          --bg-surface: #201d1b;
          --bg-muted: #292623;
          --text-primary: #f0e9e0;
          --text-secondary: #cfcac4;
          --text-muted: #a29e99;
          --border-subtle: #3e3c3a;
          --accent: #e7b063;
          --accent-soft: #493e30;
        }
      }
      * { box-sizing: border-box; }
      html, body {
        margin: 0;
        padding: 0;
        background: var(--bg-canvas);
        color: var(--text-primary);
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC",
                     "Hiragino Sans GB", "Microsoft YaHei", Roboto, sans-serif;
        font-size: 16px;
        line-height: 1.65;
        -webkit-font-smoothing: antialiased;
      }
      ::selection { background: var(--accent-soft); color: var(--text-primary); }
      a { color: var(--accent); text-decoration: none; }
      a:hover { text-decoration: underline; }
      .wrap { max-width: 560px; margin: 0 auto; padding: 28px 16px 80px; }
      .masthead {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        padding-bottom: 20px;
        border-bottom: 1px solid var(--border-subtle);
        margin-bottom: 24px;
      }
      .masthead-brand {
        display: flex;
        align-items: center;
        gap: 12px;
        min-width: 0;
      }
      .logo {
        width: 40px;
        height: 40px;
        border-radius: 50%;
        object-fit: cover;
        border: 2px solid var(--bg-surface);
        box-shadow:
          0 0 0 1px var(--border-subtle),
          0 1px 2px rgb(0 0 0 / 6%);
        flex-shrink: 0;
      }
      .masthead-brand h1 {
        margin: 0;
        font-size: 20px;
        font-weight: 700;
        letter-spacing: -0.01em;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
      .masthead-brand h1 .rss-suffix {
        font-weight: 500;
        color: var(--text-muted);
        margin-left: 6px;
      }
      .feed-badge {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        padding: 4px 12px;
        border-radius: 999px;
        font-size: 12px;
        font-weight: 600;
        letter-spacing: 0.06em;
        background: var(--accent-soft);
        color: var(--accent);
        flex-shrink: 0;
      }
      .hint {
        background: var(--bg-muted);
        border: 1px solid var(--border-subtle);
        border-radius: 12px;
        padding: 14px 16px;
        font-size: 13.5px;
        color: var(--text-secondary);
        margin-bottom: 20px;
        display: flex;
        gap: 12px;
        align-items: flex-start;
      }
      .hint svg { flex-shrink: 0; margin-top: 2px; color: var(--accent); }
      .feed-url { display: block; margin-top: 6px; }
      .hint code {
        background: var(--bg-surface);
        border: 1px solid var(--border-subtle);
        padding: 1px 6px;
        border-radius: 4px;
        font-size: 12.5px;
        color: var(--text-primary);
        word-break: break-all;
        -webkit-user-select: all;
                user-select: all;
      }
      .meta {
        margin: 0 0 32px;
        font-size: 13.5px;
        color: var(--text-muted);
        font-variant-numeric: tabular-nums;
      }
      .meta time { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 12.5px; }
      .site-footer {
        margin-top: 40px;
        text-align: center;
        font-size: 12.5px;
        color: var(--text-muted);
      }
      .site-footer a { color: var(--text-secondary); }
      .site-footer .sep { margin: 0 10px; opacity: 0.4; }
      @media (max-width: 520px) {
        .wrap { padding: 20px 14px 60px; }
        .masthead-brand h1 { font-size: 19px; }
      }
    </style>
  </head>
  <body>
    <main class="wrap">
      <header class="masthead">
        <div class="masthead-brand">
          <img class="logo" alt="" src="{{if .Logo}}{{.Logo}}{{else}}/favicon.svg{{end}}">
          <h1>
            {{.Title}}
            <span class="rss-suffix">RSS</span>
          </h1>
        </div>
        <span class="feed-badge">Feed</span>
      </header>

      <div class="hint">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 11a9 9 0 0 1 9 9"/>
          <path d="M4 4a16 16 0 0 1 16 16"/>
          <circle cx="5" cy="19" r="1.5" fill="currentColor" stroke="none"/>
        </svg>
        <div>
          This is an RSS feed. Copy this page's URL into your favorite reader (Feedly, Inoreader, Reeder…) to subscribe.
          <span class="feed-url"><code>{{.FeedURL}}</code></span>
        </div>
      </div>

      <p class="meta">
        {{if .Count}}共 {{.Count}} 条动态{{if .Updated}} · 最近更新 <time datetime="{{.UpdatedISO}}">{{.Updated}}</time>{{end}}{{else}}暂无内容{{end}}
      </p>

      <footer class="site-footer">
        <a href="{{.Home}}" target="_blank" rel="noopener">访问站点 →</a>
        <span class="sep">|</span>
        Powered by Ech0
      </footer>
    </main>
  </body>
</html>
`
