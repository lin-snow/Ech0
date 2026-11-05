// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// emptySettings reads config defaults, where agent discovery is enabled.
func emptySettings() kvstore.Store {
	return kvstore.NewMemory()
}

func TestGetHeatMap(t *testing.T) {
	cases := []struct {
		name       string
		inputTZ    string
		wantNormTZ string
	}{
		{name: "valid-iana", inputTZ: "Asia/Shanghai", wantNormTZ: "Asia/Shanghai"},
		{name: "empty-falls-back-utc", inputTZ: "", wantNormTZ: "UTC"},
		{name: "garbage-falls-back-utc", inputTZ: "Not/AZone", wantNormTZ: "UTC"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := commonmock.NewMockService(t)
			want := []commonModel.Heatmap{{Date: "2026-06-30", Count: 3}}
			svc.EXPECT().GetHeatMap(tc.wantNormTZ).Return(want, nil).Once()

			h := commonHandler.NewCommonHandler(svc, emptySettings())
			out, err := h.GetHeatMap(context.Background(), &commonHandler.GetHeatMapInput{Timezone: tc.inputTZ})

			require.NoError(t, err)
			assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
			assert.Equal(t, commonModel.GET_HEATMAP_SUCCESS, out.Message)
			assert.Equal(t, want, out.Data)
		})
	}
}

func TestGetHeatMap_ServiceError(t *testing.T) {
	svc := commonmock.NewMockService(t)
	sentinel := errors.New("db down")
	svc.EXPECT().GetHeatMap(mock.Anything).Return(nil, sentinel).Once()

	h := commonHandler.NewCommonHandler(svc, emptySettings())
	out, err := h.GetHeatMap(context.Background(), &commonHandler.GetHeatMapInput{Timezone: "UTC"})

	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, commonHandler.HeatmapOutput{}, out)
}

func TestHelloEch0(t *testing.T) {
	svc := commonmock.NewMockService(t)
	h := commonHandler.NewCommonHandler(svc, emptySettings())

	out, err := h.HelloEch0(context.Background(), &commonHandler.HelloInput{})

	require.NoError(t, err)
	assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	assert.Equal(t, commonModel.GET_HELLO_SUCCESS, out.Message)
	assert.Equal(t, "Hello, Ech0! 👋", out.Data.Hello)
	assert.Equal(t, versionPkg.Copyright(), out.Data.Copyright)
	assert.Equal(t, versionPkg.Version, out.Data.Version)
	assert.Equal(t, versionPkg.RepoURL, out.Data.RepoURL)
	assert.Equal(t, versionPkg.License, out.Data.License)
}

func TestGetWebsiteTitle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := commonmock.NewMockService(t)
		svc.EXPECT().GetWebsiteTitle("https://example.com").Return("Example Domain", nil).Once()

		h := commonHandler.NewCommonHandler(svc, emptySettings())
		out, err := h.GetWebsiteTitle(context.Background(), &commonHandler.GetWebsiteTitleInput{WebsiteURL: "https://example.com"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.GET_WEBSITE_TITLE_SUCCESS, out.Message)
		assert.Equal(t, "Example Domain", out.Data)
	})

	t.Run("service-error", func(t *testing.T) {
		svc := commonmock.NewMockService(t)
		sentinel := errors.New("dns failure")
		svc.EXPECT().GetWebsiteTitle(mock.Anything).Return("", sentinel).Once()

		h := commonHandler.NewCommonHandler(svc, emptySettings())
		out, err := h.GetWebsiteTitle(context.Background(), &commonHandler.GetWebsiteTitleInput{WebsiteURL: "https://bad"})

		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, commonHandler.StringOutput{}, out)
	})
}

func TestHealthz(t *testing.T) {
	svc := commonmock.NewMockService(t)
	h := commonHandler.NewCommonHandler(svc, emptySettings())
	r := gin.New()
	r.GET("/healthz", h.Healthz())

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	res := helpers.ParseResult(t, rec)
	assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, res.Code)

	var data struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	helpers.DecodeData(t, res.Data, &data)
	assert.Equal(t, "ok", data.Status)
	assert.Equal(t, versionPkg.Version, data.Version)
}

func TestGetRobotsTxt(t *testing.T) {
	cases := []struct {
		name        string
		host        string
		fwdProto    string
		fwdHost     string
		wantSitemap string
	}{
		{
			name:        "plain-host",
			host:        "example.com",
			wantSitemap: "Sitemap: http://example.com/sitemap.xml",
		},
		{
			name:        "forwarded-proto-and-host",
			host:        "internal:6277",
			fwdProto:    "https",
			fwdHost:     "ech0.app",
			wantSitemap: "Sitemap: https://ech0.app/sitemap.xml",
		},
		{
			name:        "forwarded-multi-value-takes-first",
			host:        "internal",
			fwdProto:    "https, http",
			fwdHost:     "ech0.app, proxy.local",
			wantSitemap: "Sitemap: https://ech0.app/sitemap.xml",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := commonmock.NewMockService(t)
			h := commonHandler.NewCommonHandler(svc, emptySettings())
			r := gin.New()
			r.GET("/robots.txt", h.GetRobotsTxt)

			req := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
			req.Host = tc.host
			if tc.fwdProto != "" {
				req.Header.Set("X-Forwarded-Proto", tc.fwdProto)
			}
			if tc.fwdHost != "" {
				req.Header.Set("X-Forwarded-Host", tc.fwdHost)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
			body := rec.Body.String()
			assert.Contains(t, body, "User-agent: *")
			assert.Contains(t, body, "Disallow: /api/")
			for _, privatePath := range []string{"/auth", "/panel", "/device", "/oauth/", "/widget", "/chat", "/zen"} {
				assert.Contains(t, body, "Disallow: "+privatePath)
			}
			assert.Contains(t, body, tc.wantSitemap)
			assert.Contains(t, body, "Agentmap: ")
		})
	}
}

func TestGetSitemap(t *testing.T) {
	svc := commonmock.NewMockService(t)
	svc.EXPECT().ListSitemapEntries(mock.Anything).Return([]echoModel.EchoRef{
		{ID: "echo-1", CreatedAt: 1735689600},
	}, nil).Once()
	h := commonHandler.NewCommonHandler(svc, emptySettings())
	r := gin.New()
	r.GET("/sitemap.xml", h.GetSitemap)

	req := httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	req.Host = "ech0.app"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/xml; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.True(t, strings.HasPrefix(rec.Body.String(), xml.Header), "sitemap should start with xml header")

	var urlset struct {
		URLs []struct {
			Loc     string `xml:"loc"`
			LastMod string `xml:"lastmod"`
		} `xml:"url"`
	}
	require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &urlset))
	locs := make([]string, 0, len(urlset.URLs))
	for _, u := range urlset.URLs {
		locs = append(locs, u.Loc)
	}
	assert.ElementsMatch(t, []string{
		"http://ech0.app/",
		"http://ech0.app/hub",
		"http://ech0.app/about",
		"http://ech0.app/rss",
		"http://ech0.app/echo/echo-1",
	}, locs)
	assert.Equal(t, "2025-01-01T00:00:00Z", urlset.URLs[len(urlset.URLs)-1].LastMod)
}

func TestGetSitemap_ServiceError(t *testing.T) {
	svc := commonmock.NewMockService(t)
	svc.EXPECT().ListSitemapEntries(mock.Anything).Return(nil, errors.New("db down")).Once()
	h := commonHandler.NewCommonHandler(svc, emptySettings())
	r := gin.New()
	r.GET("/sitemap.xml", h.GetSitemap)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetRss(t *testing.T) {
	t.Run("feed-content-type-for-subscriber", func(t *testing.T) {
		svc := commonmock.NewMockService(t)
		svc.EXPECT().GenerateRSS(mock.Anything).Return(commonService.RSSDocument{
			Atom: "<feed/>",
			HTML: "<html>styled</html>",
		}, nil).Once()
		h := commonHandler.NewCommonHandler(svc, emptySettings())
		r := gin.New()
		r.GET("/rss", h.GetRss)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/rss", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/atom+xml; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "Accept", rec.Header().Get("Vary"))
		assert.Equal(t, "<feed/>", rec.Body.String())
	})

	t.Run("browser-accept-html-gets-server-rendered-page", func(t *testing.T) {
		svc := commonmock.NewMockService(t)
		svc.EXPECT().GenerateRSS(mock.Anything).Return(commonService.RSSDocument{
			Atom: "<feed/>",
			HTML: "<html>styled</html>",
		}, nil).Once()
		h := commonHandler.NewCommonHandler(svc, emptySettings())
		r := gin.New()
		r.GET("/rss", h.GetRss)

		req := httptest.NewRequest(http.MethodGet, "/rss", nil)
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "Accept", rec.Header().Get("Vary"))
		assert.Equal(t, "<html>styled</html>", rec.Body.String())
	})

	t.Run("service-error-returns-fail-envelope", func(t *testing.T) {
		svc := commonmock.NewMockService(t)
		svc.EXPECT().GenerateRSS(mock.Anything).Return(commonService.RSSDocument{}, errors.New("rss boom")).Once()
		h := commonHandler.NewCommonHandler(svc, emptySettings())
		r := gin.New()
		r.GET("/rss", h.GetRss)

		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/rss", nil))

		require.Equal(t, http.StatusOK, rec.Code)
		res := helpers.ParseResult(t, rec)
		assert.Equal(t, commonModel.DEFAULT_FAILED_CODE, res.Code)
	})
}
