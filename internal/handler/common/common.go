// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/common"
	errorUtil "github.com/lin-snow/ech0/internal/util/err"
	"github.com/lin-snow/ech0/internal/util/reqinfo"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

type (
	GetHeatMapInput struct {
		Timezone string `header:"X-Timezone" doc:"客户端时区（IANA 名），用于按本地日界对齐热力图"`
	}
	HelloInput           struct{}
	GetWebsiteTitleInput struct {
		WebsiteURL string `query:"website_url" required:"true" doc:"目标网站 URL"`
	}

	HelloResponse struct {
		Hello     string `json:"hello"`
		Copyright string `json:"copyright"`
		versionPkg.Info
	}
)

type (
	HeatmapOutput = commonModel.Result[[]commonModel.Heatmap]
	HelloOutput   = commonModel.Result[HelloResponse]
	StringOutput  = commonModel.Result[string]
)

type CommonHandler struct {
	commonService service.Service
	durableKV     kvstore.Store
}

type siteMapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []siteMapURL `xml:"url"`
}

type siteMapURL struct {
	Loc        string `xml:"loc"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func NewCommonHandler(commonService service.Service, durableKV kvstore.Store) *CommonHandler {
	return &CommonHandler{
		commonService: commonService,
		durableKV:     durableKV,
	}
}

func (commonHandler *CommonHandler) GetHeatMap(ctx context.Context, in *GetHeatMapInput) (HeatmapOutput, error) {
	heatMap, err := commonHandler.commonService.GetHeatMap(timezoneUtil.NormalizeTimezone(in.Timezone))
	if err != nil {
		return HeatmapOutput{}, err
	}
	return commonModel.OK(heatMap, commonModel.GET_HEATMAP_SUCCESS), nil
}

func (commonHandler *CommonHandler) GetRss(ctx *gin.Context) {
	doc, err := commonHandler.commonService.GenerateRSS(ctx)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
				Msg: "",
				Err: err,
			})),
		)
		return
	}

	// The same URL yields two representations, so caches must key on Accept.
	ctx.Header("Vary", "Accept")

	if strings.Contains(ctx.GetHeader("Accept"), "text/html") {
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(doc.HTML))
		return
	}
	ctx.Data(http.StatusOK, "application/atom+xml; charset=utf-8", []byte(doc.Atom))
}

func (commonHandler *CommonHandler) HelloEch0(ctx context.Context, _ *HelloInput) (HelloOutput, error) {
	hello := HelloResponse{
		Hello:     "Hello, Ech0! 👋",
		Copyright: versionPkg.Copyright(),
		Info:      versionPkg.Get(),
	}
	return commonModel.OK(hello, commonModel.GET_HELLO_SUCCESS), nil
}

func (commonHandler *CommonHandler) Healthz() gin.HandlerFunc {
	return res.Execute(func(ctx *gin.Context) res.Response {
		return res.Response{
			Msg: commonModel.GET_HEALTHZ_SUCCESS,
			Data: struct {
				Status  string `json:"status"`
				Version string `json:"version"`
			}{
				Status:  "ok",
				Version: versionPkg.Version,
			},
		}
	})
}

func (commonHandler *CommonHandler) GetWebsiteTitle(ctx context.Context, in *GetWebsiteTitleInput) (StringOutput, error) {
	title, err := commonHandler.commonService.GetWebsiteTitle(in.WebsiteURL)
	if err != nil {
		return StringOutput{}, err
	}
	return commonModel.OK(title, commonModel.GET_WEBSITE_TITLE_SUCCESS), nil
}

func resolveBaseURL(ctx *gin.Context) string {
	return reqinfo.Origin(ctx.Request)
}

func (commonHandler *CommonHandler) GetRobotsTxt(ctx *gin.Context) {
	baseURL := strings.TrimRight(resolveBaseURL(ctx), "/")
	lines := []string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /api/",
		"Disallow: /auth",
		"Disallow: /panel",
		"Disallow: /init",
		"Disallow: /healthz",
	}
	// Agentmap advertises the catalog; it goes away with the discovery surface.
	if system, err := commonHandler.agentDiscoverySetting(ctx); err == nil && !system.DisableAgentDiscovery {
		lines = append(lines, fmt.Sprintf("Agentmap: %s%s", baseURL, AgentCatalogPath))
	}
	lines = append(lines, fmt.Sprintf("Sitemap: %s/sitemap.xml", baseURL), "")
	content := strings.Join(lines, "\n")

	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(content))
}

func (commonHandler *CommonHandler) GetSitemap(ctx *gin.Context) {
	baseURL := strings.TrimRight(resolveBaseURL(ctx), "/")
	urlSet := siteMapURLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []siteMapURL{
			{Loc: baseURL + "/", ChangeFreq: "daily", Priority: "1.0"},
			{Loc: baseURL + "/hub", ChangeFreq: "daily", Priority: "0.8"},
			{Loc: baseURL + "/rss", ChangeFreq: "hourly", Priority: "0.6"},
		},
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(urlSet); err != nil {
		ctx.JSON(
			http.StatusOK,
			commonModel.Fail[string](errorUtil.HandleError(&commonModel.ServerError{
				Msg: "",
				Err: err,
			})),
		)
		return
	}

	ctx.Data(http.StatusOK, "application/xml; charset=utf-8", buf.Bytes())
}
