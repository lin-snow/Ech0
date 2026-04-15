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
	"/swagger",
}

type WebHandler struct {
	visitorTracker *visitor.Tracker
}

// NewWebHandler WebHandler 的构造函数
func NewWebHandler(visitorTracker *visitor.Tracker) *WebHandler {
	return &WebHandler{visitorTracker: visitorTracker}
}

// Templates 返回一个处理前端编译后文件的 gin.HandlerFunc
func (webHandler *WebHandler) Templates() gin.HandlerFunc {
	// 提取 dist 子目录
	subFS, _ := fs.Sub(template.WebFS, "dist")
	fileServer := http.FS(subFS)

	return func(ctx *gin.Context) {
		requestPath := ctx.Request.URL.Path
		if shouldBypassSPAFallback(requestPath) {
			ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if requestPath == "/" {
			requestPath = "/index.html"
		}

		if strings.Contains(requestPath, "..") {
			ctx.Status(http.StatusForbidden)
			return
		}

		fullPath := path.Clean("." + requestPath)
		f, err := fileServer.Open(fullPath)
		if err != nil {
			// fallback 到 index.html
			fallback, err := fileServer.Open("index.html")
			if err != nil {
				ctx.Status(http.StatusNotFound)
				return
			}
			defer func() { _ = fallback.Close() }()
			fallbackStat, _ := fallback.Stat()
			webHandler.visitorTracker.Record(ctx.Request, ctx.ClientIP())
			ctx.Header("Content-Type", "text/html; charset=utf-8")
			setCacheControlHeader(ctx, "/index.html")
			http.ServeContent(
				ctx.Writer,
				ctx.Request,
				"index.html",
				fallbackStat.ModTime(),
				fallback,
			)
			return
		}
		defer func() { _ = f.Close() }()

		// 获取文件信息
		stat, _ := f.Stat()

		// 适配资源压缩Gzip 算法
		encoding := ctx.GetHeader("Accept-Encoding")
		if strings.Contains(encoding, "gzip") {
			gzPath := fullPath + ".gz"
			gzFile, err := fileServer.Open(gzPath)
			if err == nil {
				defer func() { _ = gzFile.Close() }()
				stat, _ := gzFile.Stat()
				ctx.Header("Content-Encoding", "gzip")
				ctx.Header("Content-Type", getMimeType(fullPath))
				setCacheControlHeader(ctx, requestPath)
				http.ServeContent(ctx.Writer, ctx.Request, gzPath, stat.ModTime(), gzFile)
				return
			}
		}

		ctx.Header("Content-Type", getMimeType(fullPath))
		setCacheControlHeader(ctx, requestPath)
		if requestPath == "/index.html" {
			webHandler.visitorTracker.Record(ctx.Request, ctx.ClientIP())
		}
		http.ServeContent(ctx.Writer, ctx.Request, fullPath, stat.ModTime(), f)
	}
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

// getMimeType 根据文件扩展名返回 MIME 类型，带默认值
func getMimeType(path string) string {
	ext := filepath.Ext(path)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return mimeType
}
