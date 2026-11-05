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
	fmt.Fprintf(&doc, "> %s — a self-hosted Ech0 microblog instance.\n\n", siteDescription(system))
	doc.WriteString("Ech0 serves its timeline over HTTP and the Model Context Protocol. " +
		"The links below are the entry points for automated clients; every URL is absolute, " +
		"so a crawler can follow it without further context.\n\n")

	doc.WriteString("## Pages and feeds\n\n")
	fmt.Fprintf(&doc, "- [Home timeline](%s/)\n", base)
	fmt.Fprintf(&doc, "- [Federated hub](%s/hub)\n", base)
	fmt.Fprintf(&doc, "- [RSS feed](%s/rss)\n", base)
	fmt.Fprintf(&doc, "- [Sitemap](%s/sitemap.xml)\n\n", base)

	doc.WriteString("## Agent surface\n\n")
	fmt.Fprintf(&doc, "- [AI catalog](%s%s): machine-readable index of this instance's agentic resources, "+
		"also served at `%s`\n", base, AgentCatalogPath, AgentCatalogLegacyPath)
	fmt.Fprintf(&doc, "- [MCP server card](%s%s): transport, protocol versions and authentication for the MCP endpoint\n",
		base, mcp.ServerCardPath)
	fmt.Fprintf(&doc, "- [OpenAPI document](%s/api/openapi.json): full REST API schema\n", base)
	fmt.Fprintf(&doc, "- [API reference UI](%s/api/docs): browsable version of the same schema\n\n", base)

	doc.WriteString("## Endpoints worth knowing\n\n")
	doc.WriteString("- `GET /api/echo/query` — public timeline; a token with the `echo:private` scope also returns private posts\n")
	doc.WriteString("- `POST /mcp` — Streamable HTTP MCP endpoint, requires an access token with the `mcp-remote` audience\n")
	doc.WriteString("- Discovery documents describe the instance, never its content; post text is available " +
		"only through the feeds and APIs above.\n")

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

func siteDescription(system settingModel.SystemSetting) string {
	if description := singleLine(system.ServerName); description != "" {
		return description
	}

	return siteTitle(system)
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
