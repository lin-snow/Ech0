// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

const (
	ProtocolVersion = "2026-07-28"
	ServerName      = "ech0-mcp"

	MCPEndpointPath         = "/mcp"
	TransportStreamableHTTP = "streamable-http"
)

const (
	protocolVersion20251125 = "2025-11-25"
	protocolVersion20250618 = "2025-06-18"
	protocolVersion20250326 = "2025-03-26"

	preferredLegacyVersion = protocolVersion20251125
)

var SupportedVersions = []string{
	ProtocolVersion,
	protocolVersion20251125,
	protocolVersion20250618,
	protocolVersion20250326,
}

func isLegacyVersion(version string) bool {
	switch version {
	case protocolVersion20251125, protocolVersion20250618, protocolVersion20250326:
		return true
	default:
		return false
	}
}

type era uint8

const (
	eraLegacy era = iota
	eraModern
)

const (
	metaKeyProtocolVersion    = "io.modelcontextprotocol/protocolVersion"
	metaKeyClientCapabilities = "io.modelcontextprotocol/clientCapabilities"
	metaKeyServerInfo         = "io.modelcontextprotocol/serverInfo"
)

const resultTypeComplete = "complete"

const (
	cacheScopePublic  = "public"
	cacheScopePrivate = "private"
)

const (
	discoverTTLMs = 60 * 60 * 1000
	listTTLMs     = 5 * 60 * 1000
	staticTTLMs   = 60 * 60 * 1000
	liveTTLMs     = 30 * 1000
)

const serverInstructions = "Ech0 personal microblog. Manage posts, tags, comments, files, connects and webhooks via tools; read site data via ech0:// resources."

type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe"`
	ListChanged bool `json:"listChanged"`
}

func serverCapabilities() ServerCapabilities {
	return ServerCapabilities{
		Tools:     &ToolsCapability{ListChanged: false},
		Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
	}
}

type ResultEnvelope struct {
	ResultType string         `json:"resultType,omitempty"`
	Meta       map[string]any `json:"_meta,omitempty"`
}

func (e *ResultEnvelope) complete(info ServerInfo) {
	e.ResultType = resultTypeComplete
	if e.Meta == nil {
		e.Meta = make(map[string]any, 1)
	}
	e.Meta[metaKeyServerInfo] = info
}

type completer interface{ complete(info ServerInfo) }

type CacheInfo struct {
	TTLMs      int64  `json:"ttlMs"`
	CacheScope string `json:"cacheScope"`
}

func publicCache(ttlMs int64) CacheInfo {
	return CacheInfo{TTLMs: ttlMs, CacheScope: cacheScopePublic}
}

func privateCache(ttlMs int64) CacheInfo {
	return CacheInfo{TTLMs: ttlMs, CacheScope: cacheScopePrivate}
}

func (c CacheInfo) normalize() CacheInfo {
	if c.CacheScope != cacheScopePublic {
		c.CacheScope = cacheScopePrivate
	}
	if c.TTLMs < 0 {
		c.TTLMs = 0
	}
	return c
}

type DiscoverResult struct {
	ResultEnvelope
	SupportedVersions []string           `json:"supportedVersions"`
	Capabilities      ServerCapabilities `json:"capabilities"`
	Instructions      string             `json:"instructions,omitempty"`
	*CacheInfo
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
