// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// serverCardNamePattern is the grammar the Server Card schema requires.
var serverCardNamePattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+/[a-zA-Z0-9._-]+$`)

func TestBuildServerCard(t *testing.T) {
	card := buildServerCard("https://blog.example.com/", "  Ech0   Blog ")

	assert.Equal(t, serverCardSchemaURI, card.Schema)
	assert.Regexp(t, serverCardNamePattern, card.Name)
	assert.Equal(t, "com.example.blog/ech0", card.Name)
	assert.Equal(t, "Ech0 Blog", card.Title)
	assert.Equal(t, versionPkg.Version, card.Version)
	assert.Equal(t, "https://blog.example.com/", card.WebsiteURL)

	require.Len(t, card.Remotes, 1)
	remote := card.Remotes[0]
	assert.Equal(t, TransportStreamableHTTP, remote.Type)
	assert.Equal(t, "https://blog.example.com"+MCPEndpointPath, remote.URL)
	assert.Equal(t, SupportedVersions, remote.SupportedProtocolVersions)
	assert.Contains(t, remote.SupportedProtocolVersions, ProtocolVersion)

	require.Len(t, remote.Headers, 1)
	header := remote.Headers[0]
	assert.Equal(t, "Authorization", header.Name)
	assert.Equal(t, "Bearer {token}", header.Value)
	assert.True(t, header.IsRequired)
	assert.True(t, header.IsSecret)
	assert.Contains(t, header.Variables, "token")
}

func TestBuildServerCard_KeepsPortInURLsButNotInName(t *testing.T) {
	card := buildServerCard("http://127.0.0.1:6277", "")

	assert.Equal(t, "1.0.0.127/ech0", card.Name)
	assert.Equal(t, "http://127.0.0.1:6277/mcp", card.Remotes[0].URL)
	assert.Equal(t, "Ech0", card.Title)
}

func TestBuildServerCard_IPv6HostFallsBackToLocalhost(t *testing.T) {
	card := buildServerCard("http://[::1]:6277", "Local")

	assert.Equal(t, "localhost/ech0", card.Name)
}

func TestBuildServerCard_TitleFitsSchemaLimit(t *testing.T) {
	long := strings.Repeat("博客 ", 60)
	card := buildServerCard("https://blog.example.com", long)

	assert.LessOrEqual(t, len([]rune(card.Title)), 100)
	assert.NotContains(t, card.Title, "  ")
}

func TestServerCardMetadataFitsSchema(t *testing.T) {
	card := buildServerCard("https://blog.example.com", "Blog")

	// description: minLength 1, maxLength 100 in the Server Card schema.
	assert.NotEmpty(t, card.Description)
	assert.LessOrEqual(t, len(card.Description), 100)
}

func TestGetServerCard(t *testing.T) {
	cases := []struct {
		name       string
		kv         func(*testing.T) kvstore.Store
		wantStatus int
	}{
		{
			name:       "enabled",
			kv:         func(t *testing.T) kvstore.Store { return settingsStore(t, `{"site_title":"Ech0 Blog"}`) },
			wantStatus: http.StatusOK,
		},
		{
			name:       "default-when-unset",
			kv:         func(t *testing.T) kvstore.Store { return settingsStore(t, "") },
			wantStatus: http.StatusOK,
		},
		{
			name:       "disabled",
			kv:         func(t *testing.T) kvstore.Store { return settingsStore(t, `{"disable_agent_discovery":true}`) },
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "settings-failure",
			kv:         func(t *testing.T) kvstore.Store { return failingStore{} },
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := &Handler{durableKV: tc.kv(t)}
			r := gin.New()
			r.GET(ServerCardPath, h.GetServerCard)

			req := httptest.NewRequest(http.MethodGet, ServerCardPath, nil)
			req.Host = "blog.example.com"
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus != http.StatusOK {
				return
			}

			assert.Equal(t, ServerCardMediaType, rec.Header().Get("Content-Type"))
			assert.NotEmpty(t, rec.Header().Get("ETag"))

			var card ServerCard
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &card))
			assert.Equal(t, serverCardSchemaURI, card.Schema)
			assert.Equal(t, "http://blog.example.com/mcp", card.Remotes[0].URL)
		})
	}
}
