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

type links struct {
	home       string
	echoPrefix string
	absolute   bool
}

func newLinks(serverURL, baseURL string) links {
	origin := strings.TrimRight(strings.TrimSpace(serverURL), "/")
	if origin == "" {
		return links{home: baseURL, echoPrefix: baseURL + "echo/"}
	}
	return links{home: origin + "/", echoPrefix: origin + "/echo/", absolute: true}
}

func (l links) resolve(u string) string {
	if !l.absolute || !strings.HasPrefix(u, "/") || strings.HasPrefix(u, "//") {
		return u
	}
	return strings.TrimSuffix(l.home, "/") + u
}

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
				url := stdhtml.EscapeString(l.resolve(ef.File.URL))
				switch storage.NormalizeCategory(ef.File.Category) {
				case storage.CategoryImage:
					mediaContent = fmt.Appendf(mediaContent,
						"<img src=\"%s\" alt=\"Image\" style=\"max-width:100%%;height:auto;\" />", url)
				case storage.CategoryVideo:
					mediaContent = fmt.Appendf(mediaContent,
						"<video controls src=\"%s\" style=\"max-width:100%%;\"><a href=\"%s\">打开视频</a></video>", url, url)
				case storage.CategoryAudio:
					mediaContent = fmt.Appendf(mediaContent,
						"<audio controls src=\"%s\"><a href=\"%s\">打开音频</a></audio>", url, url)
				default:
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
			Loc:        l.echoPrefix + e.ID,
			LastMod:    time.Unix(e.CreatedAt, 0).UTC().Format(time.DateOnly),
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
