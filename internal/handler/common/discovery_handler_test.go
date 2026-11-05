// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// urnPattern is the identifier grammar the ARD/AI catalog schema enforces.
var urnPattern = regexp.MustCompile(`^urn:air:[a-zA-Z0-9.-]+(:[a-zA-Z0-9._-]+)+$`)

// failingStore makes the settings read fail the way a broken KV backend does.
type failingStore struct{}

func (failingStore) Get(context.Context, string) (string, error) {
	return "", errors.New("kv down")
}
func (failingStore) Set(context.Context, string, string) error { return nil }
func (failingStore) Delete(context.Context, string) error      { return nil }

func settingsStore(t *testing.T, raw string) kvstore.Store {
	t.Helper()

	kv := kvstore.NewMemory()
	if raw == "" {
		return kv
	}
	if err := kv.Set(context.Background(), commonModel.SystemSettingsKey, raw); err != nil {
		t.Fatalf("seed settings: %v", err)
	}

	return kv
}

func discoveryRouter(t *testing.T, kv kvstore.Store) *gin.Engine {
	t.Helper()

	h := commonHandler.NewCommonHandler(commonmock.NewMockService(t), kv)
	r := gin.New()
	r.GET("/llms.txt", h.GetLlmsTxt)
	r.GET("/.well-known/ard.json", h.GetAgentCatalog)
	r.GET("/.well-known/ai-catalog.json", h.GetAgentCatalog)
	r.GET("/robots.txt", h.GetRobotsTxt)

	return r
}

func TestGetLlmsTxt_Document(t *testing.T) {
	r := discoveryRouter(t, settingsStore(t, `{"site_title":"Ech0 Blog","server_name":"Notes from a self-hosted corner"}`))

	req := httptest.NewRequest(http.MethodGet, "/llms.txt", nil)
	req.Host = "blog.example.com"
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.NotEmpty(t, rec.Header().Get("ETag"))
	assert.Contains(t, rec.Header().Get("Cache-Control"), "must-revalidate")

	body := rec.Body.String()
	assert.Contains(t, body, "# Ech0 Blog\n")
	assert.Contains(
		t,
		body,
		"> Notes from a self-hosted corner — a self-hosted microblogging platform powered by Ech0.",
	)
	// The guidance is "at least one link", and every link must be absolute.
	assert.Contains(t, body, "[Agent Catalog](http://blog.example.com/.well-known/ard.json)")
	assert.Contains(t, body, "[MCP Server Card](http://blog.example.com/mcp/server-card)")
	assert.Contains(t, body, "[RSS Feed](http://blog.example.com/rss)")
	assert.Contains(t, body, "`POST /api/echo/query`")
}

func TestGetLlmsTxt_DescriptionFallback(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{name: "no-server-name", raw: `{"site_title":"Ech0 Blog"}`},
		{name: "server-name-repeats-title", raw: `{"site_title":"Ech0 Blog","server_name":"ech0 blog"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := discoveryRouter(t, settingsStore(t, tc.raw))

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/llms.txt", nil))

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Contains(
				t,
				rec.Body.String(),
				"> A self-hosted microblogging platform powered by Ech0.",
			)
		})
	}
}

func TestGetLlmsTxt_UsesForwardedOrigin(t *testing.T) {
	r := discoveryRouter(t, settingsStore(t, `{"site_title":"Ech0 Blog"}`))

	req := httptest.NewRequest(http.MethodGet, "/llms.txt", nil)
	req.Host = "internal:6277"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "blog.example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "https://blog.example.com/.well-known/ard.json")
}

func TestGetAgentCatalog_Document(t *testing.T) {
	r := discoveryRouter(t, settingsStore(t, `{"site_title":"Ech0 Blog"}`))

	for _, path := range []string{"/.well-known/ard.json", "/.well-known/ai-catalog.json"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.Host = "blog.example.com"
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/ai-catalog+json", rec.Header().Get("Content-Type"))

			var doc struct {
				SpecVersion string `json:"specVersion"`
				Host        struct {
					DisplayName string `json:"displayName"`
				} `json:"host"`
				Entries []struct {
					Identifier            string   `json:"identifier"`
					DisplayName           string   `json:"displayName"`
					Type                  string   `json:"type"`
					URL                   string   `json:"url"`
					Version               string   `json:"version"`
					RepresentativeQueries []string `json:"representativeQueries"`
				} `json:"entries"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))

			assert.Equal(t, "1.0", doc.SpecVersion)
			assert.Equal(t, "Ech0 Blog", doc.Host.DisplayName)
			require.Len(t, doc.Entries, 1)

			entry := doc.Entries[0]
			assert.Regexp(t, urnPattern, entry.Identifier)
			assert.Equal(t, "urn:air:blog.example.com:mcp:ech0", entry.Identifier)
			assert.Equal(t, "application/mcp-server-card+json", entry.Type)
			assert.Equal(t, "http://blog.example.com/mcp/server-card", entry.URL)
			assert.NotEmpty(t, entry.Version)
			assert.GreaterOrEqual(t, len(entry.RepresentativeQueries), 2)
			assert.LessOrEqual(t, len(entry.RepresentativeQueries), 5)
		})
	}
}

func TestAgentDiscovery_Disabled(t *testing.T) {
	r := discoveryRouter(t, settingsStore(t, `{"site_title":"Private","disable_agent_discovery":true}`))

	for _, path := range []string{"/llms.txt", "/.well-known/ard.json", "/.well-known/ai-catalog.json"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Host = "blog.example.com"
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code, path)
		assert.Empty(t, rec.Body.String(), path)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, rec.Body.String(), "Agentmap:")
	assert.Contains(t, rec.Body.String(), "Sitemap: ")
}

func TestAgentDiscovery_SettingsFailure(t *testing.T) {
	r := discoveryRouter(t, failingStore{})

	for _, path := range []string{"/llms.txt", "/.well-known/ard.json"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code, path)
	}
}

func TestAgentDiscovery_ETagRevalidation(t *testing.T) {
	r := discoveryRouter(t, settingsStore(t, `{"site_title":"Ech0 Blog"}`))

	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/llms.txt", nil))
	require.Equal(t, http.StatusOK, first.Code)
	etag := first.Header().Get("ETag")
	require.NotEmpty(t, etag)

	req := httptest.NewRequest(http.MethodGet, "/llms.txt", nil)
	req.Header.Set("If-None-Match", etag)
	second := httptest.NewRecorder()
	r.ServeHTTP(second, req)

	assert.Equal(t, http.StatusNotModified, second.Code)
	assert.Empty(t, second.Body.String())
}
