// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	stdhtml "html"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/storage"
	mdUtil "github.com/lin-snow/ech0/internal/util/md"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/egress"
	"golang.org/x/net/html"
)

type CommonService struct {
	commonRepository CommonRepository
	cache            *cache.Cache
	durableKV        kvstore.Store
}

func NewCommonService(
	commonRepository CommonRepository,
	cache *cache.Cache,
	durableKV kvstore.Store,
) *CommonService {
	return &CommonService{
		commonRepository: commonRepository,
		cache:            cache,
		durableKV:        durableKV,
	}
}

func (s *CommonService) CommonGetUserByUserId(ctx context.Context, userId string) (userModel.User, error) {
	return s.commonRepository.GetUserByUserId(ctx, userId)
}

func (s *CommonService) GetOwner() (userModel.User, error) {
	return s.commonRepository.GetOwner(context.Background())
}

// ListSitemapEntries returns every public echo as an id + timestamp pair. It is
// deliberately a projection: a sitemap must not pay for file and tag hydration.
func (s *CommonService) ListSitemapEntries(ctx context.Context) ([]echoModel.EchoRef, error) {
	return s.commonRepository.ListPublicEchoRefs(ctx)
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

const rssDefaultTitle = "Ech0"

// rssCacheKey folds the feed identity into the key: the site title is rendered
// into the payload, so renaming the site must miss the cache instead of serving
// the previous name until the next echo mutation clears it.
func rssCacheKey(schema, host, title, description string) string {
	identity := sha256.Sum256([]byte(title + "\x00" + description))
	return fmt.Sprintf("rss:%s:%s:%x", schema, host, identity[:6])
}

func (s *CommonService) GenerateRSS(ctx *gin.Context) (RSSDocument, error) {
	schema := "http"
	if ctx.Request.TLS != nil {
		schema = "https"
	}
	host := ctx.Request.Host

	system, err := coreSetting.System.Get(ctx.Request.Context(), s.durableKV)
	if err != nil {
		return RSSDocument{}, err
	}
	title := strings.TrimSpace(system.SiteTitle)
	if title == "" {
		title = rssDefaultTitle
	}
	description := strings.TrimSpace(system.ServerName)
	if description == "" {
		description = title
	}
	cacheKey := rssCacheKey(schema, host, title, description)

	return s.cache.ReadThrough(
		cacheKey,
		1,
		func() (RSSDocument, error) {
			echos, err := s.commonRepository.GetAllEchos(ctx.Request.Context(), false)
			if err != nil {
				return RSSDocument{}, err
			}

			home := fmt.Sprintf("%s://%s/", schema, host)
			logo := fmt.Sprintf("%s://%s/Ech0.svg", schema, host)
			baseURL := &url.URL{Scheme: schema, Host: host}
			page := rssPage{
				Title:   title,
				Logo:    logo,
				Home:    home,
				FeedURL: fmt.Sprintf("%s://%s/rss", schema, host),
			}

			feed := &feeds.Feed{
				Title:       title,
				Link:        &feeds.Link{Href: home},
				Image:       &feeds.Image{Url: logo},
				Description: description,
				Author:      &feeds.Author{Name: title},
				Updated:     time.Now().UTC(),
			}

			var latest time.Time
			for _, msg := range echos {
				renderedContent := mdUtil.MdToHTML([]byte(msg.Content))
				createdAt := time.Unix(msg.CreatedAt, 0).UTC()
				title := msg.Username + " - " + createdAt.Format("2006-01-02")
				if createdAt.After(latest) {
					latest = createdAt
				}

				if len(msg.EchoFiles) > 0 {
					var mediaContent []byte
					for _, ef := range msg.EchoFiles {
						if ef.File.URL == "" {
							continue
						}
						// Local storage yields a relative URL; absolutizeFeedHTML
						// resolves it against the feed origin for readers.
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

				renderedContent = absolutizeFeedHTML(renderedContent, baseURL)

				entryLink := fmt.Sprintf("%s://%s/echo/%s", schema, host, msg.ID)
				feed.Items = append(feed.Items, &feeds.Item{
					Title:   title,
					Link:    &feeds.Link{Href: entryLink},
					Content: string(renderedContent),
					Author:  &feeds.Author{Name: msg.Username},
					Created: createdAt,
				})
			}

			// The landing page reports liveness only; entry bodies stay in Atom.
			page.Count = len(echos)
			if !latest.IsZero() {
				page.Updated = latest.Format("2006-01-02")
				page.UpdatedISO = latest.Format(time.RFC3339)
			}

			atom, err := feed.ToAtom()
			if err != nil {
				return RSSDocument{}, err
			}
			body, err := renderRSSPage(page)
			if err != nil {
				return RSSDocument{}, err
			}

			s.commonRepository.TrackRSSCacheKey(cacheKey)
			return RSSDocument{Atom: atom, HTML: body}, nil
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
