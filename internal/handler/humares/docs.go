// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const scalarVersion = "1.62.0"

//go:embed assets/scalar.standalone.js.gz
var scalarBundleGz []byte

type DocsRenderer string

const (
	DocsRendererStoplight DocsRenderer = "stoplight"
	DocsRendererScalar    DocsRenderer = "scalar"
)

func ParseDocsRenderer(s string) DocsRenderer {
	if strings.EqualFold(strings.TrimSpace(s), string(DocsRendererScalar)) {
		return DocsRendererScalar
	}
	return DocsRendererStoplight
}

const (
	scalarDocsRoute   = "/docs"
	scalarScriptRoute = "/docs/scalar.standalone.js"
	scalarSpecRoute   = "/openapi.json"
)

func registerScalarDocs(group *gin.RouterGroup, basePath string) {
	html := buildScalarHTML(basePath)
	group.GET(scalarDocsRoute, func(ctx *gin.Context) {
		ctx.Header("Cache-Control", "no-cache")
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", html)
	})
	group.GET(scalarScriptRoute, serveScalarBundle)
}

func buildScalarHTML(basePath string) []byte {
	scriptSrc := basePath + scalarScriptRoute
	specURL := basePath + scalarSpecRoute
	return []byte(`<!doctype html>
<!-- Scalar API Reference (self-hosted) v` + scalarVersion + ` -->
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Ech0 API 文档</title>
</head>
<body>
  <div id="app"></div>
  <script src="` + scriptSrc + `"></script>
  <script>
    Scalar.createApiReference('#app', { url: '` + specURL + `' })
  </script>
</body>
</html>`)
}

func serveScalarBundle(ctx *gin.Context) {
	const contentType = "application/javascript; charset=utf-8"
	ctx.Header("Cache-Control", "public, max-age=86400")
	ctx.Header("Vary", "Accept-Encoding")

	if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
		ctx.Header("Content-Encoding", "gzip")
		ctx.Data(http.StatusOK, contentType, scalarBundleGz)
		return
	}

	reader, err := gzip.NewReader(bytes.NewReader(scalarBundleGz))
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	defer func() { _ = reader.Close() }()
	ctx.Header("Content-Type", contentType)
	ctx.Status(http.StatusOK)
	_, _ = io.Copy(ctx.Writer, reader)
}
