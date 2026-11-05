// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/mcp"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	httputil "github.com/lin-snow/ech0/internal/util/httpdoc"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	versionPkg "github.com/lin-snow/ech0/internal/version"
)

// Machine-readable discovery surface: the files describe the instance — feeds,
// APIs, MCP endpoint — never its content.
const (
	LLMsTxtPath = "/llms.txt"

	// AgentCatalogPath is the ARD entry point; the legacy path serves the same
	// document for tooling that still looks there (Lighthouse, AI Catalog).
	AgentCatalogPath       = "/.well-known/ard.json"
	AgentCatalogLegacyPath = "/.well-known/ai-catalog.json"

	llmsTxtContentType      = "text/plain; charset=utf-8"
	agentCatalogContentType = "application/ai-catalog+json"

	agentCatalogSpecVersion = "1.0"
	defaultSiteTitle        = "Ech0"

	agentCatalogNamespace = "mcp"
	agentCatalogServer    = "ech0"
)

func (commonHandler *CommonHandler) GetLlmsTxt(ctx *gin.Context) {
	system, ok := commonHandler.requireAgentDiscovery(ctx)
	if !ok {
		return
	}

	httputil.Write(ctx, llmsTxtContentType, []byte(buildLLMsTxt(resolveBaseURL(ctx), system)))
}

func (commonHandler *CommonHandler) GetAgentCatalog(ctx *gin.Context) {
	system, ok := commonHandler.requireAgentDiscovery(ctx)
	if !ok {
		return
	}

	body, err := json.Marshal(buildAgentCatalog(resolveBaseURL(ctx), system))
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	httputil.Write(ctx, agentCatalogContentType, body)
}

// agentDiscoverySetting loads the settings that shape every discovery document.
func (commonHandler *CommonHandler) agentDiscoverySetting(ctx *gin.Context) (settingModel.SystemSetting, error) {
	return coreSetting.System.Get(ctx.Request.Context(), commonHandler.durableKV)
}

// requireAgentDiscovery answers the request itself when the instance has opted
// out of the discovery surface.
func (commonHandler *CommonHandler) requireAgentDiscovery(ctx *gin.Context) (settingModel.SystemSetting, bool) {
	system, err := commonHandler.agentDiscoverySetting(ctx)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return system, false
	}
	if system.DisableAgentDiscovery {
		ctx.Status(http.StatusNotFound)
		return system, false
	}

	return system, true
}

func buildLLMsTxt(baseURL string, system settingModel.SystemSetting) string {
	base := strings.TrimRight(baseURL, "/")
	title := siteTitle(system)

	var doc strings.Builder
	fmt.Fprintf(&doc, "# %s\n\n", title)
	if description := singleLine(system.ServerName); description != "" && !strings.EqualFold(description, title) {
		fmt.Fprintf(&doc, "> %s — a self-hosted microblogging platform powered by Ech0.\n\n", description)
	} else {
		doc.WriteString("> A self-hosted microblogging platform powered by Ech0.\n\n")
	}
	doc.WriteString("Ech0 is a personal publishing platform that provides content feeds, REST APIs, and " +
		"Model Context Protocol (MCP) integration. Use the following resources to discover and " +
		"interact with this instance.\n\n")

	doc.WriteString("## Content & Feeds\n\n")
	fmt.Fprintf(&doc, "- [Home](%s/): Browse the public timeline.\n", base)
	fmt.Fprintf(&doc, "- [Federated Hub](%s/hub): Explore the federated community.\n", base)
	fmt.Fprintf(&doc, "- [RSS Feed](%s/rss): Subscribe to public updates.\n", base)
	fmt.Fprintf(&doc, "- [Sitemap](%s/sitemap.xml): Discover public pages.\n\n", base)

	doc.WriteString("## Developer & Agent Resources\n\n")
	fmt.Fprintf(&doc, "- [Agent Catalog](%s%s): Machine-readable catalog of available agent services.\n",
		base, AgentCatalogPath)
	fmt.Fprintf(&doc, "- [MCP Server Card](%s%s): MCP endpoint, transport, protocol capabilities, "+
		"and authentication information.\n", base, mcp.ServerCardPath)
	fmt.Fprintf(&doc, "- [OpenAPI Specification](%s/api/openapi.json): Complete REST API specification.\n", base)
	fmt.Fprintf(&doc, "- [API Documentation](%s/api/docs): Interactive API reference.\n\n", base)

	doc.WriteString("## API & Integration\n\n")
	doc.WriteString("- `POST /api/echo/query`: Query public posts. Access to private posts requires " +
		"an Access Token with the `echo:private` scope.\n")
	doc.WriteString("- `POST /mcp`: Streamable HTTP MCP endpoint for AI agent integration. Supports " +
		"Bearer Access Token authentication and OAuth 2.1 authorization.\n\n")

	doc.WriteString("## Integration Guide\n\n")
	doc.WriteString("- **RSS**: Passive content consumption and timeline subscriptions.\n")
	doc.WriteString("- **REST API**: Direct programmatic access to posts and other resources.\n")
	doc.WriteString("- **MCP**: AI agent integration for interacting with this instance.\n\n")

	doc.WriteString("MCP clients can authenticate using either a pre-issued Access Token or an OAuth 2.1 " +
		"authorization flow, depending on their integration requirements.\n\n")
	doc.WriteString("Refer to the corresponding API specifications and MCP Server Card for detailed " +
		"capabilities, authentication requirements, and usage instructions.\n")

	return doc.String()
}

type agentCatalogDocument struct {
	SpecVersion string              `json:"specVersion"`
	Host        agentCatalogHost    `json:"host"`
	Entries     []agentCatalogEntry `json:"entries"`
}

type agentCatalogHost struct {
	DisplayName      string `json:"displayName"`
	DocumentationURL string `json:"documentationUrl,omitempty"`
	LogoURL          string `json:"logoUrl,omitempty"`
}

type agentCatalogEntry struct {
	Identifier            string   `json:"identifier"`
	DisplayName           string   `json:"displayName"`
	Type                  string   `json:"type"`
	URL                   string   `json:"url"`
	Version               string   `json:"version,omitempty"`
	Description           string   `json:"description,omitempty"`
	RepresentativeQueries []string `json:"representativeQueries,omitempty"`
}

func buildAgentCatalog(baseURL string, system settingModel.SystemSetting) agentCatalogDocument {
	base := strings.TrimRight(baseURL, "/")
	title := siteTitle(system)

	return agentCatalogDocument{
		SpecVersion: agentCatalogSpecVersion,
		Host: agentCatalogHost{
			DisplayName:      title,
			DocumentationURL: versionPkg.RepoURL,
			LogoURL:          base + "/favicon.svg",
		},
		Entries: []agentCatalogEntry{{
			Identifier:  "urn:air:" + urlUtil.SafeHost(base) + ":" + agentCatalogNamespace + ":" + agentCatalogServer,
			DisplayName: title + " MCP server",
			Type:        mcp.ServerCardMediaType,
			URL:         base + mcp.ServerCardPath,
			Version:     versionPkg.Version,
			Description: "Read and write this Ech0 instance's posts, comments, files and stats " +
				"over the Model Context Protocol.",
			RepresentativeQueries: []string{
				"find recent posts on this timeline",
				"search this instance for posts mentioning a topic",
				"summarize what this microblog published lately",
			},
		}},
	}
}

func siteTitle(system settingModel.SystemSetting) string {
	if title := singleLine(system.SiteTitle); title != "" {
		return title
	}

	return defaultSiteTitle
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
