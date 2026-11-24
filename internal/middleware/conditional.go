// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ConditionalGet answers If-None-Match for JSON reads with a 304, so a client
// polling the public timeline revalidates instead of re-downloading it.
//
// Only bodies the server is allowed to cache participate: a response carrying
// Cache-Control: no-store (every authenticated route) is never buffered, and
// neither is anything that is not JSON or that a handler flushes before
// finishing.
func ConditionalGet() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet:
		default:
			c.Next()
			return
		}

		writer := &conditionalWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()
		writer.finish(c)
	}
}

// conditionalWriter buffers a cacheable JSON body so the ETag can be derived
// from its bytes; everything else is forwarded untouched.
type conditionalWriter struct {
	gin.ResponseWriter

	buf         bytes.Buffer
	decided     bool
	buffer      bool
	wroteHeader bool
}

func (w *conditionalWriter) decide() {
	if w.decided {
		return
	}
	w.decided = true

	header := w.Header()
	contentType := header.Get("Content-Type")
	cacheControl := strings.ToLower(header.Get("Cache-Control"))
	w.buffer = strings.HasPrefix(contentType, "application/json") &&
		!strings.Contains(cacheControl, "no-store") &&
		header.Get("Content-Encoding") == ""
}

func (w *conditionalWriter) WriteHeader(code int) {
	w.decide()
	if w.buffer && code == http.StatusOK {
		w.wroteHeader = true
		return
	}
	w.buffer = false
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *conditionalWriter) WriteHeaderNow() {
	w.decide()
	if w.buffer {
		w.wroteHeader = true
		return
	}
	w.ResponseWriter.WriteHeaderNow()
}

func (w *conditionalWriter) Write(p []byte) (int, error) {
	w.decide()
	if !w.buffer {
		return w.ResponseWriter.Write(p)
	}
	w.wroteHeader = true
	return w.buf.Write(p)
}

func (w *conditionalWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *conditionalWriter) Written() bool {
	if w.buffer {
		return w.wroteHeader
	}
	return w.ResponseWriter.Written()
}

func (w *conditionalWriter) Status() int {
	if w.buffer && w.wroteHeader {
		return http.StatusOK
	}
	return w.ResponseWriter.Status()
}

func (w *conditionalWriter) Size() int {
	if w.buffer {
		return w.buf.Len()
	}
	return w.ResponseWriter.Size()
}

// Flush means the handler streams, so the body must not be buffered.
func (w *conditionalWriter) Flush() {
	w.decided, w.buffer = true, false
	w.ResponseWriter.Flush()
}

func (w *conditionalWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.decided, w.buffer = true, false
	return w.ResponseWriter.Hijack()
}

func (w *conditionalWriter) finish(c *gin.Context) {
	if !w.buffer || !w.wroteHeader {
		return
	}

	body := w.buf.Bytes()
	if len(body) == 0 {
		w.ResponseWriter.WriteHeader(http.StatusOK)
		return
	}

	tag := weakETag(body)
	w.Header().Set("ETag", tag)
	if matchesETag(c.Request.Header.Get("If-None-Match"), tag) {
		c.Status(http.StatusNotModified)
		return
	}

	w.ResponseWriter.WriteHeader(http.StatusOK)
	_, _ = w.ResponseWriter.Write(body)
}

func weakETag(body []byte) string {
	sum := sha256.Sum256(body)
	return `W/"` + hex.EncodeToString(sum[:8]) + `"`
}

// matchesETag accepts the weak comparison If-None-Match requires: any listed
// tag that strips to the same opaque value, plus the `*` wildcard.
func matchesETag(header, tag string) bool {
	if header == "" {
		return false
	}
	for candidate := range strings.SplitSeq(header, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || strings.TrimPrefix(candidate, "W/") == strings.TrimPrefix(tag, "W/") {
			return true
		}
	}
	return false
}
