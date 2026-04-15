package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/lin-snow/ech0/internal/cache"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	httpUtil "github.com/lin-snow/ech0/internal/util/http"
	mdUtil "github.com/lin-snow/ech0/internal/util/md"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
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
	for i := 0; i < 30; i++ {
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
					var imageContent []byte
					for _, ef := range msg.EchoFiles {
						imageContent = fmt.Appendf(
							imageContent,
							"<img src=\"%s\" alt=\"Image\" style=\"max-width:100%%;height:auto;\" />",
							ef.File.URL,
						)
					}
					renderedContent = append(imageContent, renderedContent...)
				}

				if len(msg.Tags) > 0 {
					for _, tag := range msg.Tags {
						renderedContent = fmt.Appendf(renderedContent, "<br /><span class=\"tag\">#%s</span>", tag.Name)
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

			s.commonRepository.TrackRSSCacheKey(cacheKey)
			return atom, nil
		},
	)
}

func (s *CommonService) GetWebsiteTitle(websiteURL string) (string, error) {
	websiteURL = httpUtil.TrimURL(websiteURL)

	body, err := httpUtil.SendSafeRequest(websiteURL, "GET", httpUtil.Header{}, 10*time.Second)
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
