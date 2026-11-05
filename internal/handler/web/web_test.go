// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/visitor"
)

type stubSettings struct {
	setting settingModel.SystemSetting
}

func (s stubSettings) GetSetting(setting *settingModel.SystemSetting) error {
	*setting = s.setting
	return nil
}

type stubEchoes struct {
	echo *echoModel.Echo
	err  error
}

func (s stubEchoes) GetEchoById(_ context.Context, id string) (*echoModel.Echo, error) {
	if s.err != nil {
		return nil, s.err
	}
	if s.echo == nil || s.echo.ID != id {
		return nil, errors.New("echo not found")
	}
	return s.echo, nil
}

const testEchoID = "0190f3f4-0000-7000-8000-000000000001"

func newTemplatesRouter() *gin.Engine {
	return newTemplatesRouterWith(stubSettings{}, stubEchoes{})
}

func newTemplatesRouterWith(settings SettingsReader, echoes EchoReader) *gin.Engine {
	r := gin.New()
	r.NoRoute(NewWebHandler(visitor.NewTracker(), settings, echoes).Templates())
	return r
}

func get(r *gin.Engine, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Host = "blog.example.com"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestTemplates_HomeRendersAbsoluteSocialMeta(t *testing.T) {
	r := newTemplatesRouterWith(stubSettings{
		setting: settingModel.SystemSetting{
			SiteTitle:     "My Blog",
			ServerName:    "My Blog",
			DefaultLocale: "en-US",
		},
	}, stubEchoes{})

	rec := get(r, "/")
	body := rec.Body.String()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	for _, want := range []string{
		`<title>Home | My Blog</title>`,
		`<link rel="canonical" href="http://blog.example.com/" />`,
		`<meta property="og:url" content="http://blog.example.com/" />`,
		`<meta property="og:image" content="http://blog.example.com/og-cover.jpg" />`,
		`<meta property="og:image:type" content="image/jpeg" />`,
		`<meta property="og:image:width" content="1200" />`,
		`<meta property="og:image:height" content="630" />`,
		`<meta property="og:locale" content="en_US" />`,
		`<meta property="og:type" content="website" />`,
		`<meta name="robots" content="index,follow,max-image-preview:large" />`,
		`<html lang="en-US"`,
		`"@type":"WebSite"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q\nbody:\n%s", want, body)
		}
	}
}

func TestTemplates_RequestOriginWinsOverPlaceholderServerURL(t *testing.T) {
	r := newTemplatesRouterWith(stubSettings{
		setting: settingModel.SystemSetting{ServerURL: "https://ech0.example.com"},
	}, stubEchoes{})

	body := get(r, "/hub").Body.String()

	if !strings.Contains(body, `content="http://blog.example.com/hub"`) {
		t.Fatalf("expected canonical to use the request origin, body:\n%s", body)
	}
	if strings.Contains(body, "ech0.example.com") {
		t.Fatalf("placeholder server URL must not leak into canonical URLs")
	}
}

func TestTemplates_EchoPageRendersArticleMeta(t *testing.T) {
	echo := &echoModel.Echo{
		ID:        testEchoID,
		Content:   "Hello **world**\n\nSecond line with <b>markup</b>.",
		Username:  "alice",
		CreatedAt: 1735689600,
		EchoFiles: []echoModel.EchoFile{{
			File: fileModel.File{
				URL:         "/api/files/cover.png",
				ContentType: "image/png",
				Category:    "image",
				Width:       800,
				Height:      600,
			},
		}},
	}
	r := newTemplatesRouterWith(
		stubSettings{setting: settingModel.SystemSetting{SiteTitle: "My Blog"}},
		stubEchoes{echo: echo},
	)

	rec := get(r, "/echo/"+testEchoID)
	body := rec.Body.String()

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	for _, want := range []string{
		`<title>Hello world | My Blog</title>`,
		`<meta property="og:type" content="article" />`,
		`<meta property="article:published_time" content="2025-01-01T00:00:00Z" />`,
		`<meta property="og:image" content="http://blog.example.com/api/files/cover.png" />`,
		`<meta property="og:image:type" content="image/png" />`,
		`<meta property="og:image:width" content="800" />`,
		`<meta property="og:image:height" content="600" />`,
		`<link rel="canonical" href="http://blog.example.com/echo/` + testEchoID + `" />`,
		`"@type":"BlogPosting"`,
		`<noscript><article><h1>Hello world | My Blog</h1>`,
		`Second line with markup.`,
		`"name":"alice"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected body to contain %q\nbody:\n%s", want, body)
		}
	}
}

func TestTemplates_MissingEchoIsNotFoundAndNoIndex(t *testing.T) {
	r := newTemplatesRouterWith(stubSettings{}, stubEchoes{})

	rec := get(r, "/echo/does-not-exist")
	body := rec.Body.String()

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
	if got := rec.Header().Get("X-Robots-Tag"); got != robotsNoIndex {
		t.Fatalf("expected X-Robots-Tag %q, got %q", robotsNoIndex, got)
	}
	if !strings.Contains(body, `content="noindex,nofollow"`) {
		t.Fatalf("expected noindex robots meta, body:\n%s", body)
	}
	if strings.Contains(body, `property="og:type" content="article"`) {
		t.Fatalf("missing echo must not render article meta")
	}
}

func TestTemplates_UnknownRouteIsNotFound(t *testing.T) {
	r := newTemplatesRouter()

	rec := get(r, "/some/deep/link")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `<div id="app">`) {
		t.Fatalf("expected the shell so the client router can render its 404 view")
	}
}

func TestTemplates_MissingAssetStaysPlainNotFound(t *testing.T) {
	r := newTemplatesRouter()

	rec := get(r, "/missing-image.png")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Fatalf("missing assets must not be answered with the SPA shell")
	}
}

func TestTemplates_PrivateRouteGetsNoIndexHeader(t *testing.T) {
	r := newTemplatesRouter()

	rec := get(r, "/panel/dashboard")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	if got := rec.Header().Get("X-Robots-Tag"); got != robotsNoIndex {
		t.Fatalf("expected X-Robots-Tag %q, got %q", robotsNoIndex, got)
	}
	if !strings.Contains(rec.Body.String(), `<title>Panel | Ech0</title>`) {
		t.Fatalf("expected panel title, body:\n%s", rec.Body.String())
	}
}

func TestTemplates_EscapesUserContentInHead(t *testing.T) {
	echo := &echoModel.Echo{
		ID:      testEchoID,
		Content: `1 < 2 & 3 > 2 and <img src=x onerror=alert(1)>`,
	}
	r := newTemplatesRouterWith(stubSettings{}, stubEchoes{echo: echo})

	body := get(r, "/echo/"+testEchoID).Body.String()

	if strings.Contains(body, `<img src=x`) {
		t.Fatalf("echo content must never be emitted as live markup")
	}
	if !strings.Contains(body, `\u003c`) {
		t.Fatalf("JSON-LD must escape angle brackets, body:\n%s", body)
	}
}

func TestTemplates_APIPathReturnsNotFound(t *testing.T) {
	r := newTemplatesRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/not-found", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestTemplates_UnknownWellKnownReturnsNotFound(t *testing.T) {
	r := newTemplatesRouter()

	req := httptest.NewRequest(http.MethodGet, "/.well-known/unknown.json", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d, body: %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestTemplates_SPAFallbackSetsNoStoreHeaders(t *testing.T) {
	r := newTemplatesRouter()

	rec := get(r, "/panel/dashboard")

	if got := rec.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
		t.Fatalf("expected Cache-Control no-store header, got %q", got)
	}
	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected Pragma no-cache, got %q", got)
	}
	if got := rec.Header().Get("Expires"); got != "0" {
		t.Fatalf("expected Expires 0, got %q", got)
	}
}

func TestSetCacheControlHeader_DefaultStaticUsesShortCache(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	setCacheControlHeader(ctx, "/favicon.ico")

	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Fatalf("expected short cache header, got %q", got)
	}
}

func TestSetCacheControlHeader_IndexDisablesCacheStorage(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	setCacheControlHeader(ctx, "/index.html")

	if got := rec.Header().Get("Cache-Control"); got != "no-cache, no-store, must-revalidate" {
		t.Fatalf("expected Cache-Control no-store header, got %q", got)
	}
	if got := rec.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("expected Pragma no-cache, got %q", got)
	}
	if got := rec.Header().Get("Expires"); got != "0" {
		t.Fatalf("expected Expires 0, got %q", got)
	}
}

func TestPlainTextStripsMarkup(t *testing.T) {
	got := plainText("## Title\n\n![img](/a.png) [link](https://x.test) **bold** `code`\n\n- item")
	want := "Title\nlink bold code\nitem"
	if got != want {
		t.Fatalf("plainText() = %q, want %q", got, want)
	}
}

func TestTruncateRunesKeepsMultibyteText(t *testing.T) {
	got := truncateRunes("你好世界你好世界", 5)
	if got != "你好世界…" {
		t.Fatalf("truncateRunes() = %q", got)
	}
}
