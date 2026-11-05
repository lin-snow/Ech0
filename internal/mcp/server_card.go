// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	httputil "github.com/lin-snow/ech0/internal/util/httpdoc"
	"github.com/lin-snow/ech0/internal/util/reqinfo"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

// The MCP Server Card extension (SEP-2127) reserves `<streamable-http-url>
// /server-card` as the canonical location, so the card for a server mounted at
// /mcp lives at /mcp/server-card. The card is public, static metadata: it
// carries connection details only, never tools, which stay behind the protocol.
const (
	ServerCardPath      = MCPEndpointPath + "/server-card"
	ServerCardMediaType = "application/mcp-server-card+json"

	serverCardSchemaURI   = "https://static.modelcontextprotocol.io/schemas/v1/server-card.schema.json"
	serverCardServerName  = "ech0"
	serverCardDescription = "Read and write this Ech0 instance's posts, comments, files and stats over MCP."
	serverCardTitleMaxLen = 100
	serverCardAuthName    = "Authorization"
	serverCardAuthValue   = "Bearer {token}"
	serverCardAuthVarName = "token"
)

type ServerCard struct {
	Schema      string             `json:"$schema"`
	Name        string             `json:"name"`
	Title       string             `json:"title,omitempty"`
	Version     string             `json:"version"`
	Description string             `json:"description"`
	WebsiteURL  string             `json:"websiteUrl,omitempty"`
	Icons       []ServerCardIcon   `json:"icons,omitempty"`
	Remotes     []ServerCardRemote `json:"remotes,omitempty"`
}

type ServerCardIcon struct {
	Src      string   `json:"src"`
	MimeType string   `json:"mimeType,omitempty"`
	Sizes    []string `json:"sizes,omitempty"`
}

type ServerCardRemote struct {
	Type                      string             `json:"type"`
	URL                       string             `json:"url"`
	SupportedProtocolVersions []string           `json:"supportedProtocolVersions,omitempty"`
	Headers                   []ServerCardHeader `json:"headers,omitempty"`
}

type ServerCardHeader struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description,omitempty"`
	IsRequired  bool                          `json:"isRequired,omitempty"`
	IsSecret    bool                          `json:"isSecret,omitempty"`
	Value       string                        `json:"value,omitempty"`
	Variables   map[string]ServerCardVariable `json:"variables,omitempty"`
}

type ServerCardVariable struct {
	Description string `json:"description,omitempty"`
	IsRequired  bool   `json:"isRequired,omitempty"`
	IsSecret    bool   `json:"isSecret,omitempty"`
}

func (h *Handler) GetServerCard(ctx *gin.Context) {
	system, err := coreSetting.System.Get(ctx.Request.Context(), h.durableKV)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if system.DisableAgentDiscovery {
		ctx.Status(http.StatusNotFound)
		return
	}

	body, err := json.Marshal(buildServerCard(reqinfo.Origin(ctx.Request), system.SiteTitle))
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	httputil.Write(ctx, ServerCardMediaType, body)
}

func buildServerCard(baseURL, siteTitle string) ServerCard {
	base := strings.TrimRight(baseURL, "/")
	title := strings.Join(strings.Fields(siteTitle), " ")
	if title == "" {
		title = "Ech0"
	}
	title = clipRunes(title, serverCardTitleMaxLen)

	return ServerCard{
		Schema:      serverCardSchemaURI,
		Name:        serverCardName(base),
		Title:       title,
		Version:     versionPkg.Version,
		Description: serverCardDescription,
		WebsiteURL:  base + "/",
		Icons: []ServerCardIcon{{
			Src:      base + "/favicon.svg",
			MimeType: "image/svg+xml",
			Sizes:    []string{"any"},
		}},
		Remotes: []ServerCardRemote{{
			Type:                      TransportStreamableHTTP,
			URL:                       base + MCPEndpointPath,
			SupportedProtocolVersions: SupportedVersions,
			Headers: []ServerCardHeader{{
				Name:        serverCardAuthName,
				Description: "Ech0 access token with the mcp-remote audience, created in Preferences → Access tokens.",
				IsRequired:  true,
				IsSecret:    true,
				Value:       serverCardAuthValue,
				Variables: map[string]ServerCardVariable{
					serverCardAuthVarName: {
						Description: "Access token granted the scopes the client needs",
						IsRequired:  true,
						IsSecret:    true,
					},
				},
			}},
		}},
	}
}

// serverCardName renders the reverse-DNS name the card schema requires.
func serverCardName(baseURL string) string {
	labels := strings.Split(urlUtil.Host(baseURL), ".")
	for i, j := 0, len(labels)-1; i < j; i, j = i+1, j-1 {
		labels[i], labels[j] = labels[j], labels[i]
	}

	return urlUtil.SafeHost(strings.Join(labels, ".")) + "/" + serverCardServerName
}

// clipRunes trims to limit runes without padding the length back up.
func clipRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}

	return string(runes[:limit])
}
