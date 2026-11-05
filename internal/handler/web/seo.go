// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/util/reqinfo"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const (
	defaultSiteTitle   = "Ech0"
	defaultDescription = "Ech0 is a next-generation open-source self-hosted platform built for individuals. It is lightweight and low-cost, making it easy to publish and share your ideas, writing, and links."
	defaultCoverImage  = "/og-cover.jpg"
	defaultLogoImage   = "/Ech0.png"
	defaultLang        = "zh-CN"

	coverImageWidth  = 1200
	coverImageHeight = 630

	seoStartMarker      = "<!--ECH0_SEO_START-->"
	seoEndMarker        = "<!--ECH0_SEO_END-->"
	noscriptStartMarker = "<!--ECH0_NOSCRIPT_START-->"
	noscriptEndMarker   = "<!--ECH0_NOSCRIPT_END-->"

	robotsIndex   = "index,follow,max-image-preview:large"
	robotsNoIndex = "noindex,nofollow"

	echoTitleMaxLen       = 60
	echoDescriptionMaxLen = 160
	echoNoScriptMaxLen    = 4000
)

// SettingsReader and EchoReader are the narrow slices of the setting and echo
// domains the shell renderer needs. Keeping them as interfaces keeps this
// package testable without a database.
type SettingsReader interface {
	GetSetting(setting *settingModel.SystemSetting) error
}

type EchoReader interface {
	GetEchoById(ctx context.Context, id string) (*echoModel.Echo, error)
}

// spaRoute mirrors the meta of web/src/router/index.ts. A path that is absent
// here is answered with 404 so search engines never index a soft 404; the shell
// is still returned so the client router can render its own not-found view.
type spaRoute struct {
	title       string
	description string
	noindex     bool
}

var notFoundRoute = spaRoute{
	title:       "404",
	description: "Requested page was not found.",
	noindex:     true,
}

var spaRoutes = map[string]spaRoute{
	"/": {
		title:       "Home",
		description: "Ech0 home timeline for publishing and browsing thoughts, notes, and links.",
	},
	"/publish": {
		title:       "Home",
		description: "Ech0 home timeline for publishing and browsing thoughts, notes, and links.",
	},
	"/hub": {
		title:       "Hub",
		description: "Discover and explore curated content in Ech0 hub.",
	},
	"/about": {
		title:       "About",
		description: "Copyright, license and author information for this Ech0 instance.",
	},
	"/auth": {
		title:       "Sign In",
		description: "Sign in to your Ech0 workspace.",
		noindex:     true,
	},
	"/device": {
		title:       "Device Authorization",
		description: "Approve a CLI sign-in request for this Ech0 instance.",
		noindex:     true,
	},
	"/oauth/authorize": {
		title:       "MCP Authorization",
		description: "Approve an MCP client request for this Ech0 instance.",
		noindex:     true,
	},
	"/widget": {
		title:       "Widget",
		description: "Ech0 embeddable widget.",
		noindex:     true,
	},
	"/init": {
		title:       "Initialize",
		description: "Initialize your Ech0 instance.",
		noindex:     true,
	},
	"/chat": {
		title:       "Chat",
		description: "Reflect on and summarize your past echos with AI.",
		noindex:     true,
	},
	"/zen": {
		title:       "Zen",
		description: "Browse all echos in an immersive masonry view.",
		noindex:     true,
	},
}

var panelRoute = spaRoute{
	title:       "Panel",
	description: "Ech0 management panel.",
	noindex:     true,
}

var (
	codeFencePattern   = regexp.MustCompile("(?s)```.*?```")
	htmlTagPattern     = regexp.MustCompile(`(?s)</?[a-zA-Z][^>]*>|<!--.*?-->`)
	imagePattern       = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	linkPattern        = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	listPrefixPattern  = regexp.MustCompile(`(?m)^[ \t]*(?:[-*+]|\d+\.)[ \t]+`)
	spaceRunPattern    = regexp.MustCompile(`[ \t]+`)
	punctuationPattern = regexp.MustCompile(`[ \t]+([.,!?;:，。！？；：])`)
	markdownNoiseChars = strings.NewReplacer(
		"*", "", "_", "", "`", "", "~", "", "#", "", ">", "", "|", " ", "\\", "",
	)
)

type pageMeta struct {
	Title       string
	Description string
	Canonical   string
	Image       string
	ImageAlt    string
	ImageType   string
	ImageWidth  int
	ImageHeight int
	Type        string
	PublishedAt string
	Robots      string
	Lang        string
	SiteTitle   string
	JSONLD      string
	NoScript    string
	NoIndex     bool
}

// spaShell keeps the byte ranges that are replaced per request. Both marker
// pairs are optional: an older or hand-edited shell falls back to injecting
// before </head> and </body>, so a stale dist can never break rendering.
type spaShell struct {
	doc                []byte
	headStart, headEnd int
	bodyStart, bodyEnd int
}

func loadSPAShell(subFS fs.FS) spaShell {
	doc, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		return newSPAShell([]byte(
			`<!doctype html><html lang="` + defaultLang +
				`"><head></head><body><div id="app"></div></body></html>`,
		))
	}
	return newSPAShell(doc)
}

func newSPAShell(doc []byte) spaShell {
	shell := spaShell{doc: doc}

	shell.headStart, shell.headEnd = markerRange(doc, seoStartMarker, seoEndMarker)
	if shell.headStart < 0 {
		shell.headStart = bytes.Index(doc, []byte("</head>"))
		if shell.headStart < 0 {
			shell.headStart = len(doc)
		}
		shell.headEnd = shell.headStart
	}

	shell.bodyStart, shell.bodyEnd = markerRange(doc, noscriptStartMarker, noscriptEndMarker)
	if shell.bodyStart < 0 {
		shell.bodyStart = bytes.Index(doc, []byte("</body>"))
		if shell.bodyStart < 0 {
			shell.bodyStart = len(doc)
		}
		shell.bodyEnd = shell.bodyStart
	}

	return shell
}

func markerRange(doc []byte, startMarker, endMarker string) (int, int) {
	start := bytes.Index(doc, []byte(startMarker))
	if start < 0 {
		return -1, -1
	}
	offset := bytes.Index(doc[start:], []byte(endMarker))
	if offset < 0 {
		return -1, -1
	}
	return start, start + offset + len(endMarker)
}

func (shell spaShell) render(headHTML, noscriptHTML, lang string) []byte {
	// The body range sits after the head range, so replacing it first leaves the
	// recorded head offsets valid for the second replacement.
	out := replaceRange(shell.doc, shell.bodyStart, shell.bodyEnd, noscriptHTML)
	out = replaceRange(out, shell.headStart, shell.headEnd, headHTML)

	if lang != "" && lang != defaultLang {
		out = bytes.Replace(
			out,
			[]byte(`lang="`+defaultLang+`"`),
			[]byte(`lang="`+lang+`"`),
			1,
		)
	}

	return out
}

func replaceRange(doc []byte, start, end int, replacement string) []byte {
	if start < 0 || end < start {
		return doc
	}

	out := make([]byte, 0, len(doc)-(end-start)+len(replacement))
	out = append(out, doc[:start]...)
	out = append(out, replacement...)
	out = append(out, doc[end:]...)
	return out
}

func (webHandler *WebHandler) serveRoute(ctx *gin.Context, routePath string) {
	if echoID, ok := echoRouteID(routePath); ok {
		echo, err := webHandler.publicEcho(ctx.Request.Context(), echoID)
		if err != nil {
			webHandler.servePage(ctx, routePath, http.StatusNotFound, nil, notFoundRoute)
			return
		}
		webHandler.servePage(ctx, routePath, http.StatusOK, echo, spaRoute{})
		return
	}

	if route, ok := lookupRoute(routePath); ok {
		webHandler.servePage(ctx, routePath, http.StatusOK, nil, route)
		return
	}

	// A missing static asset must stay a plain 404: answering it with an HTML
	// shell is how SPAs end up with images indexed as soft 404 pages.
	if path.Ext(routePath) != "" {
		ctx.Status(http.StatusNotFound)
		return
	}

	webHandler.servePage(ctx, routePath, http.StatusNotFound, nil, notFoundRoute)
}

func (webHandler *WebHandler) servePage(
	ctx *gin.Context,
	routePath string,
	status int,
	echo *echoModel.Echo,
	route spaRoute,
) {
	meta := webHandler.buildPageMeta(ctx.Request, routePath, echo, route)
	body := webHandler.shell.render(meta.headHTML(), meta.NoScript, meta.Lang)

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	setCacheControlHeader(ctx, "/index.html")
	if meta.NoIndex {
		ctx.Header("X-Robots-Tag", robotsNoIndex)
	}

	ctx.Status(status)
	_, _ = ctx.Writer.Write(body)
}

func (webHandler *WebHandler) buildPageMeta(
	request *http.Request,
	routePath string,
	echo *echoModel.Echo,
	route spaRoute,
) pageMeta {
	settings := webHandler.siteSettings()
	origin := resolveSiteOrigin(request)
	lang := normalizeLang(settings.DefaultLocale)
	canonical := origin + routePath

	meta := pageMeta{
		Canonical:   canonical,
		Image:       origin + defaultCoverImage,
		ImageAlt:    settings.SiteTitle,
		ImageType:   "image/jpeg",
		ImageWidth:  coverImageWidth,
		ImageHeight: coverImageHeight,
		Type:        "website",
		Robots:      robotsIndex,
		Lang:        lang,
		SiteTitle:   settings.SiteTitle,
		NoIndex:     route.noindex,
	}

	if route.noindex {
		meta.Robots = robotsNoIndex
	}

	if echo != nil {
		meta.Title = pageTitle(settings.SiteTitle, echoTitle(echo))
		meta.Description = echoDescription(echo)
		meta.Type = "article"
		meta.PublishedAt = time.Unix(echo.CreatedAt, 0).UTC().Format(time.RFC3339)

		if image, imageType, width, height := echoShareImage(echo, origin); image != "" {
			meta.Image = image
			meta.ImageType = imageType
			meta.ImageAlt = meta.Title
			meta.ImageWidth = width
			meta.ImageHeight = height
		}

		meta.JSONLD = echoJSONLD(echo, canonical, origin, settings.SiteTitle, meta.Image, lang, meta.Description)
		meta.NoScript = echoNoScript(echo, canonical, meta.Title)
		return meta
	}

	meta.Title = pageTitle(settings.SiteTitle, route.title)
	meta.Description = route.description
	if strings.TrimSpace(meta.Description) == "" {
		meta.Description = defaultDescription
	}
	meta.JSONLD = websiteJSONLD(settings.SiteTitle, meta.Description, origin)

	return meta
}

func (webHandler *WebHandler) siteSettings() settingModel.SystemSetting {
	var settings settingModel.SystemSetting
	if webHandler.settings != nil {
		_ = webHandler.settings.GetSetting(&settings)
	}
	if strings.TrimSpace(settings.SiteTitle) == "" {
		settings.SiteTitle = defaultSiteTitle
	}
	return settings
}

// publicEcho resolves a shareable post as an anonymous visitor. Private posts
// are reported as missing so the shell never confirms that an id exists.
func (webHandler *WebHandler) publicEcho(
	ctx context.Context,
	echoID string,
) (*echoModel.Echo, error) {
	if webHandler.echoes == nil {
		return nil, errEchoUnavailable
	}

	echo, err := webHandler.echoes.GetEchoById(
		viewer.WithContext(ctx, viewer.NewNoopViewer()),
		echoID,
	)
	if err != nil {
		return nil, err
	}
	if echo == nil || echo.Private {
		return nil, errEchoUnavailable
	}
	return echo, nil
}

var errEchoUnavailable = fmt.Errorf("echo unavailable")

func normalizeRoutePath(raw string) string {
	if raw == "" {
		return "/"
	}
	if len(raw) > 1 {
		raw = strings.TrimSuffix(raw, "/")
	}
	return raw
}

func lookupRoute(routePath string) (spaRoute, bool) {
	if route, ok := spaRoutes[routePath]; ok {
		return route, true
	}
	if routePath == "/panel" || strings.HasPrefix(routePath, "/panel/") {
		return panelRoute, true
	}
	return spaRoute{}, false
}

func echoRouteID(routePath string) (string, bool) {
	const prefix = "/echo/"

	if !strings.HasPrefix(routePath, prefix) {
		return "", false
	}
	id := strings.TrimPrefix(routePath, prefix)
	if id == "" || strings.Contains(id, "/") {
		return "", false
	}
	return id, true
}

// resolveSiteOrigin returns the origin the crawler actually visited. RSS,
// sitemap and the discovery documents already derive URLs the same way, and a
// canonical that disagrees with the fetched URL is the one thing Google will
// not trust. `server_url` is deliberately not consulted: it ships with a
// placeholder default that would poison every share card on a fresh install.
func resolveSiteOrigin(request *http.Request) string {
	return reqinfo.Origin(request)
}

func normalizeLang(raw string) string {
	lang := strings.TrimSpace(raw)
	if lang == "" {
		return defaultLang
	}
	return lang
}

func pageTitle(siteTitle, pageTitle string) string {
	title := strings.TrimSpace(pageTitle)
	if title == "" || title == siteTitle {
		return siteTitle
	}
	return title + " | " + siteTitle
}

func echoTitle(echo *echoModel.Echo) string {
	for _, line := range strings.Split(plainText(echo.Content), "\n") {
		if line == "" {
			continue
		}
		return truncateRunes(line, echoTitleMaxLen)
	}
	return "Echo"
}

func echoDescription(echo *echoModel.Echo) string {
	text := strings.ReplaceAll(plainText(echo.Content), "\n", " ")
	if text == "" {
		return defaultDescription
	}
	return truncateRunes(text, echoDescriptionMaxLen)
}

func echoShareImage(echo *echoModel.Echo, origin string) (string, string, int, int) {
	for _, echoFile := range echo.EchoFiles {
		file := echoFile.File
		if file.Category != "image" && !strings.HasPrefix(file.ContentType, "image/") {
			continue
		}
		absolute := absoluteAssetURL(file.URL, origin)
		if absolute == "" {
			continue
		}

		contentType := file.ContentType
		if contentType == "" {
			contentType = mime.TypeByExtension(path.Ext(absolute))
		}
		return absolute, contentType, file.Width, file.Height
	}
	return "", "", 0, 0
}

func absoluteAssetURL(raw, origin string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err == nil && parsed.IsAbs() {
		return raw
	}
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	if strings.HasPrefix(raw, "/") {
		return origin + raw
	}
	return origin + "/" + raw
}

func plainText(raw string) string {
	text := codeFencePattern.ReplaceAllString(raw, " ")
	text = htmlTagPattern.ReplaceAllString(text, " ")
	text = imagePattern.ReplaceAllString(text, " ")
	text = linkPattern.ReplaceAllString(text, "$1")
	text = listPrefixPattern.ReplaceAllString(text, "")
	text = markdownNoiseChars.Replace(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")

	lines := make([]string, 0, 8)
	for _, line := range strings.Split(text, "\n") {
		line = spaceRunPattern.ReplaceAllString(line, " ")
		line = punctuationPattern.ReplaceAllString(line, "$1")
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}

func (meta pageMeta) headHTML() string {
	var out strings.Builder
	line := func(format string, args ...any) {
		fmt.Fprintf(&out, format+"\n", args...)
	}

	line(`<title>%s</title>`, html.EscapeString(meta.Title))
	line(`<meta name="description" content="%s" />`, html.EscapeString(meta.Description))
	line(`<meta name="robots" content="%s" />`, meta.Robots)
	line(`<link rel="canonical" href="%s" />`, html.EscapeString(meta.Canonical))

	line(`<meta property="og:type" content="%s" />`, meta.Type)
	line(`<meta property="og:site_name" content="%s" />`, html.EscapeString(meta.SiteTitle))
	line(`<meta property="og:title" content="%s" />`, html.EscapeString(meta.Title))
	line(`<meta property="og:description" content="%s" />`, html.EscapeString(meta.Description))
	line(`<meta property="og:url" content="%s" />`, html.EscapeString(meta.Canonical))
	line(`<meta property="og:image" content="%s" />`, html.EscapeString(meta.Image))
	if meta.ImageType != "" {
		line(`<meta property="og:image:type" content="%s" />`, html.EscapeString(meta.ImageType))
	}
	if meta.ImageWidth > 0 && meta.ImageHeight > 0 {
		line(`<meta property="og:image:width" content="%d" />`, meta.ImageWidth)
		line(`<meta property="og:image:height" content="%d" />`, meta.ImageHeight)
	}
	line(`<meta property="og:image:alt" content="%s" />`, html.EscapeString(meta.ImageAlt))
	line(`<meta property="og:locale" content="%s" />`, html.EscapeString(ogLocale(meta.Lang)))

	if meta.Type == "article" && meta.PublishedAt != "" {
		line(`<meta property="article:published_time" content="%s" />`, meta.PublishedAt)
	}

	line(`<meta name="twitter:card" content="summary_large_image" />`)
	line(`<meta name="twitter:title" content="%s" />`, html.EscapeString(meta.Title))
	line(`<meta name="twitter:description" content="%s" />`, html.EscapeString(meta.Description))
	line(`<meta name="twitter:image" content="%s" />`, html.EscapeString(meta.Image))
	line(`<meta name="twitter:image:alt" content="%s" />`, html.EscapeString(meta.ImageAlt))

	line(`<script id="website-jsonld" type="application/ld+json">%s</script>`, meta.JSONLD)

	return out.String()
}

func ogLocale(lang string) string {
	return strings.ReplaceAll(lang, "-", "_")
}

func websiteJSONLD(siteTitle, description, origin string) string {
	return encodeJSONLD(map[string]any{
		"@context":    "https://schema.org",
		"@type":       "WebSite",
		"name":        siteTitle,
		"description": description,
		"url":         origin + "/",
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  siteTitle,
			"url":   origin,
			"logo":  origin + defaultLogoImage,
		},
		"sameAs": []string{"https://github.com/lin-snow/Ech0"},
	})
}

func echoJSONLD(
	echo *echoModel.Echo,
	canonical string,
	origin string,
	siteTitle string,
	image string,
	lang string,
	description string,
) string {
	published := time.Unix(echo.CreatedAt, 0).UTC().Format(time.RFC3339)
	author := strings.TrimSpace(echo.Username)
	if author == "" {
		author = siteTitle
	}

	headline := truncateRunes(echoTitle(echo), 110)

	return encodeJSONLD(map[string]any{
		"@context":         "https://schema.org",
		"@type":            "BlogPosting",
		"headline":         headline,
		"description":      description,
		"datePublished":    published,
		"dateModified":     published,
		"inLanguage":       lang,
		"url":              canonical,
		"mainEntityOfPage": map[string]any{"@type": "WebPage", "@id": canonical},
		"image":            image,
		"author":           map[string]any{"@type": "Person", "name": author},
		"publisher": map[string]any{
			"@type": "Organization",
			"name":  siteTitle,
			"url":   origin,
			"logo":  origin + defaultLogoImage,
		},
	})
}

func echoNoScript(echo *echoModel.Echo, canonical, title string) string {
	text := truncateRunes(strings.ReplaceAll(plainText(echo.Content), "\n", " "), echoNoScriptMaxLen)
	if text == "" {
		return ""
	}

	var out strings.Builder
	out.WriteString(`<noscript><article><h1>`)
	out.WriteString(html.EscapeString(title))
	out.WriteString(`</h1><p>`)
	out.WriteString(html.EscapeString(text))
	out.WriteString(`</p><p><a href="`)
	out.WriteString(html.EscapeString(canonical))
	out.WriteString(`">Read the full post</a></p></article></noscript>`)

	return out.String()
}

func encodeJSONLD(payload any) string {
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	// Escape the three characters that could terminate the surrounding
	// <script> element when the payload carries user content.
	return strings.NewReplacer(
		"<", `\u003c`,
		">", `\u003e`,
		"&", `\u0026`,
	).Replace(string(raw))
}
