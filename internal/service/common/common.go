// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"fmt"
	stdhtml "html"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/util/egress"
	mdUtil "github.com/lin-snow/ech0/internal/util/md"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"golang.org/x/net/html"
)

type CommonService struct {
	commonRepository CommonRepository
	cache            cache.ICache[string, any]
}

func NewCommonService(
	commonRepository CommonRepository,
	cache cache.ICache[string, any],
) *CommonService {
	return &CommonService{
		commonRepository: commonRepository,
		cache:            cache,
	}
}

func (s *CommonService) CommonGetUserByUserId(ctx context.Context, userId string) (userModel.User, error) {
	return s.commonRepository.GetUserByUserId(ctx, userId)
}

func (s *CommonService) GetOwner() (userModel.User, error) {
	return s.commonRepository.GetOwner(context.Background())
}

func (s *CommonService) GetHeatMap(timezone string) ([]commonModel.Heatmap, error) {
	ctx := context.Background()
	loc := timezoneUtil.LoadLocationOrUTC(timezone)
	nowUser := time.Now().UTC().In(loc)
	startUser := time.Date(nowUser.Year(), nowUser.Month(), nowUser.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, -29)
	endUserExclusive := startUser.AddDate(0, 0, 30)

	createdAtList, err := s.commonRepository.GetHeatMap(ctx, startUser.UTC().Unix(), endUserExclusive.UTC().Unix())
	if err != nil {
		return nil, err
	}

	countMap := make(map[string]int)
	for _, createdAt := range createdAtList {
		day := time.Unix(createdAt, 0).In(loc).Format("2006-01-02")
		countMap[day]++
	}

	var results [30]commonModel.Heatmap
	for i := range 30 {
		date := startUser.AddDate(0, 0, i).Format("2006-01-02")
		results[i] = commonModel.Heatmap{
			Date:  date,
			Count: countMap[date],
		}
	}

	return results[:], nil
}

func (s *CommonService) GenerateRSS(ctx *gin.Context) (string, error) {
	schema := "http"
	if ctx.Request.TLS != nil {
		schema = "https"
	}
	host := ctx.Request.Host
	cacheKey := "rss:" + schema + ":" + host

	return cache.ReadThroughTyped[string](
		s.cache,
		cacheKey,
		1,
		func() (string, error) {
			echos, err := s.commonRepository.GetAllEchos(ctx.Request.Context(), false)
			if err != nil {
				return "", err
			}

			feed := &feeds.Feed{
				Title:       "Ech0",
				Link:        &feeds.Link{Href: fmt.Sprintf("%s://%s/", schema, host)},
				Image:       &feeds.Image{Url: fmt.Sprintf("%s://%s/Ech0.svg", schema, host)},
				Description: "Ech0",
				Author:      &feeds.Author{Name: "Ech0"},
				Updated:     time.Now().UTC(),
			}

			for _, msg := range echos {
				renderedContent := mdUtil.MdToHTML([]byte(msg.Content))
				createdAt := time.Unix(msg.CreatedAt, 0).UTC()
				title := msg.Username + " - " + createdAt.Format("2006-01-02")

				if len(msg.EchoFiles) > 0 {
					var mediaContent []byte
					for _, ef := range msg.EchoFiles {
						if ef.File.URL == "" {
							continue
						}
						url := stdhtml.EscapeString(ef.File.URL)
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

				if len(msg.Tags) > 0 {
					for _, tag := range msg.Tags {
						renderedContent = fmt.Appendf(
							renderedContent,
							"<br /><span class=\"tag\">#%s</span>",
							stdhtml.EscapeString(tag.Name),
						)
					}
				}

				feed.Items = append(feed.Items, &feeds.Item{
					Title:       title,
					Link:        &feeds.Link{Href: fmt.Sprintf("%s://%s/echo/%s", schema, host, msg.ID)},
					Description: string(renderedContent),
					Author:      &feeds.Author{Name: msg.Username},
					Created:     createdAt,
				})
			}

			atom, err := feed.ToAtom()
			if err != nil {
				return "", err
			}

			const xmlDecl = `<?xml version="1.0" encoding="UTF-8"?>`
			const stylesheetPI = `<?xml-stylesheet type="text/xsl" href="/rss.xsl"?>`
			atom = strings.Replace(atom, xmlDecl, xmlDecl+"\n"+stylesheetPI, 1)

			s.commonRepository.TrackRSSCacheKey(cacheKey)
			return atom, nil
		},
	)
}

func (s *CommonService) GetWebsiteTitle(websiteURL string) (string, error) {
	websiteURL = urlUtil.TrimURL(websiteURL)

	body, err := egress.Fetch(websiteURL, "GET", egress.Header{}, 10*time.Second)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("解析 HTML 失败: %w", err)
	}

	title := extractTitle(doc)
	if title == "" {
		return "", errors.New("未找到网站标题")
	}

	return title, nil
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
		return ""
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}
	return ""
}
