// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

import (
	"encoding/xml"
	"fmt"
	stdhtml "html"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/lin-snow/ech0/internal/storage"
	mdUtil "github.com/lin-snow/ech0/internal/util/md"
)

// links 是站点的链接基址。静态站可能被部署到任意域名下，胶囊里唯一的线索
// 是 site.server_url；它缺席时只能退化成相对链接（RSS/sitemap 里的相对
// 链接不理想，但比编造一个错误的域名强）。
type links struct {
	home       string // 首页链接，恒以 / 结尾
	echoPrefix string // Echo 详情页前缀，恒以 / 结尾
	absolute   bool   // 是否拿到了真实 origin
}

func newLinks(serverURL, baseURL string) links {
	origin := strings.TrimRight(strings.TrimSpace(serverURL), "/")
	if origin == "" {
		return links{home: baseURL, echoPrefix: baseURL + "echo/"}
	}
	return links{home: origin + "/", echoPrefix: origin + "/echo/", absolute: true}
}

// resolve 把站内绝对根路径（/api/files/…）补成完整 URL。RSS 阅读器不在本站
// 上下文里渲染，相对 src 会直接变成死链。
func (l links) resolve(u string) string {
	if !l.absolute || !strings.HasPrefix(u, "/") || strings.HasPrefix(u, "//") {
		return u
	}
	return strings.TrimSuffix(l.home, "/") + u
}

// renderAtom 生成 rss.xml（Atom 格式，与活实例 GET /rss 同形）。
func renderAtom(ds *dataset, l links, generatedAt time.Time) (string, error) {
	title := ds.Settings.SiteTitle
	if title == "" {
		title = "Ech0"
	}
	description := ds.Settings.ServerName
	if description == "" {
		description = title
	}

	feed := &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: l.home},
		Image:       &feeds.Image{Url: l.home + "Ech0.svg"},
		Description: description,
		Author:      &feeds.Author{Name: title},
		Updated:     generatedAt.UTC(),
	}

	for i := range ds.Echos {
		e := &ds.Echos[i]
		renderedContent := mdUtil.MdToHTML([]byte(e.Content))
		createdAt := time.Unix(e.CreatedAt, 0).UTC()
		itemTitle := e.Username + " - " + createdAt.Format(time.DateOnly)

		if len(e.EchoFiles) > 0 {
			var mediaContent []byte
			for _, ef := range e.EchoFiles {
				if ef.File.URL == "" {
					continue
				}
				// URL 进属性、文件名进链接文本都是可能来自 external 的用户可控字段，进入
				// <summary type="html"> 前必须做 HTML 实体转义，阻断订阅器二次解码触发的
				// stored XSS（与下方标签转义同一注入类，GHSA-3v85-fqvh-7rxf）。
				url := stdhtml.EscapeString(l.resolve(ef.File.URL))
				switch storage.NormalizeCategory(ef.File.Category) {
				case storage.CategoryImage:
					mediaContent = fmt.Appendf(mediaContent,
						"<img src=\"%s\" alt=\"Image\" style=\"max-width:100%%;height:auto;\" />", url)
				case storage.CategoryVideo:
					// 内嵌 <a> 兜底：RSS 阅读器若剥离 <video> 标签，仍退化成可点链接，不丢内容。
					mediaContent = fmt.Appendf(mediaContent,
						"<video controls src=\"%s\" style=\"max-width:100%%;\"><a href=\"%s\">打开视频</a></video>", url, url)
				case storage.CategoryAudio:
					mediaContent = fmt.Appendf(mediaContent,
						"<audio controls src=\"%s\"><a href=\"%s\">打开音频</a></audio>", url, url)
				default:
					// pdf / document / file / markdown：给一个可点的下载链接。
					name := stdhtml.EscapeString(ef.File.Name)
					if name == "" {
						name = "下载文件"
					}
					mediaContent = fmt.Appendf(mediaContent, "<p>📎 <a href=\"%s\">%s</a></p>", url, name)
				}
			}
			renderedContent = append(mediaContent, renderedContent...)
		}

		for _, t := range e.Tags {
			// 标签名进入 RSS Atom <summary type="html"> 后会被订阅器二次解码并渲染成 HTML，
			// 必须先做 HTML 实体转义阻断 stored XSS（GHSA-3v85-fqvh-7rxf）。
			renderedContent = fmt.Appendf(
				renderedContent,
				"<br /><span class=\"tag\">#%s</span>",
				stdhtml.EscapeString(t.Name),
			)
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       itemTitle,
			Link:        &feeds.Link{Href: l.echoPrefix + e.ID},
			Description: string(renderedContent),
			Author:      &feeds.Author{Name: e.Username},
			Created:     createdAt,
		})
	}

	return feed.ToAtom()
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// renderSitemap 生成 sitemap.xml：首页 + 每条 Echo 详情页。
func renderSitemap(ds *dataset, l links, generatedAt time.Time) ([]byte, error) {
	set := sitemapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]sitemapURL, 0, len(ds.Echos)+1),
	}
	set.URLs = append(set.URLs, sitemapURL{
		Loc:        l.home,
		LastMod:    generatedAt.UTC().Format(time.DateOnly),
		ChangeFreq: "daily",
		Priority:   "1.0",
	})
	for i := range ds.Echos {
		e := &ds.Echos[i]
		set.URLs = append(set.URLs, sitemapURL{
			Loc:     l.echoPrefix + e.ID,
			LastMod: time.Unix(e.CreatedAt, 0).UTC().Format(time.DateOnly),
			// 静态站是冻结快照，单条内容此后不会再变。
			ChangeFreq: "never",
			Priority:   "0.8",
		})
	}

	body, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal sitemap: %w", err)
	}
	return append([]byte(xml.Header), body...), nil
}
