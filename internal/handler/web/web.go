// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/visitor"
	"github.com/lin-snow/ech0/template"
)

var spaBypassPrefixes = []string{
	"/api",
	"/ws",
	"/mcp",
	// Well-known URIs are protocol endpoints, not app routes: an unknown one
	// must answer 404 instead of handing a client the SPA shell as 200 HTML.
	"/.well-known",
}

type WebHandler struct {
	visitorTracker *visitor.Tracker
	settings       SettingsReader
	echoes         EchoReader
	files          http.FileSystem
	shell          spaShell
}

func NewWebHandler(
	visitorTracker *visitor.Tracker,
	settings SettingsReader,
	echoes EchoReader,
) *WebHandler {
	subFS, _ := fs.Sub(template.WebFS, "dist")

	return &WebHandler{
		visitorTracker: visitorTracker,
		settings:       settings,
		echoes:         echoes,
		files:          http.FS(subFS),
		shell:          loadSPAShell(subFS),
	}
}

// Templates serves the built SPA. The HTML shell is rendered per request so
// crawlers and link unfurlers — which do not run JavaScript — still receive a
// route-specific title, description, canonical URL and social-card image.
func (webHandler *WebHandler) Templates() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.Path
		if shouldBypassSPAFallback(requestPath) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if strings.Contains(requestPath, "..") {
			ctx.Status(http.StatusForbidden)
			return
		}

		if requestPath == "/" || requestPath == "/index.html" {
			webHandler.visitorTracker.Record(ctx.Request, ctx.ClientIP())
			webHandler.serveRoute(ctx, "/")
			return
		}

		fullPath := path.Clean("." + requestPath)
		f, err := webHandler.files.Open(fullPath)
		if err != nil {
			webHandler.visitorTracker.Record(ctx.Request, ctx.ClientIP())
			webHandler.serveRoute(ctx, normalizeRoutePath(requestPath))
			return
		}
		defer func() { _ = f.Close() }()

		webHandler.serveAsset(ctx, fullPath, requestPath, f)
	}
}

func (webHandler *WebHandler) serveAsset(
	ctx *gin.Context,
	fullPath string,
	requestPath string,
	file http.File,
) {
	stat, _ := file.Stat()

	encoding := ctx.GetHeader("Accept-Encoding")
	if strings.Contains(encoding, "gzip") {
		gzPath := fullPath + ".gz"
		gzFile, err := webHandler.files.Open(gzPath)
		if err == nil {
			defer func() { _ = gzFile.Close() }()
			gzStat, _ := gzFile.Stat()
			ctx.Header("Content-Encoding", "gzip")
			ctx.Header("Content-Type", getMimeType(fullPath))
			setCacheControlHeader(ctx, requestPath)
			http.ServeContent(ctx.Writer, ctx.Request, gzPath, gzStat.ModTime(), gzFile)
			return
		}
	}

	ctx.Header("Content-Type", getMimeType(fullPath))
	setCacheControlHeader(ctx, requestPath)
	http.ServeContent(ctx.Writer, ctx.Request, fullPath, stat.ModTime(), file)
}

func setCacheControlHeader(ctx *gin.Context, requestPath string) {
	if strings.HasPrefix(requestPath, "/assets/") {
		ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
		return
	}
	if requestPath == "/index.html" || requestPath == "/" {
		setNoStoreHeaders(ctx)
		return
	}
	ctx.Header("Cache-Control", "public, max-age=3600")
}

func setNoStoreHeaders(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	ctx.Header("Pragma", "no-cache")
	ctx.Header("Expires", "0")
}

func shouldBypassSPAFallback(requestPath string) bool {
	for _, prefix := range spaBypassPrefixes {
		if requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") {
			return true
		}
	}
	return false
}

func getMimeType(path string) string {
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}
