// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

const (
	ProtocolVersion = "2026-07-28"
	ServerName      = "ech0-mcp"
)

var SupportedVersions = []string{ProtocolVersion}

const (
	metaKeyProtocolVersion = "io.modelcontextprotocol/protocolVersion"
	metaKeyServerInfo      = "io.modelcontextprotocol/serverInfo"
)

const resultTypeComplete = "complete"

const (
	cacheScopePublic  = "public"
	cacheScopePrivate = "private"
)

const (
	discoverTTLMs = 60 * 60 * 1000
	listTTLMs     = 5 * 60 * 1000
)

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

type ResultEnvelope struct {
	ResultType string         `json:"resultType"`
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

type DiscoverResult struct {
	ResultEnvelope
	SupportedVersions []string           `json:"supportedVersions"`
	Capabilities      ServerCapabilities `json:"capabilities"`
	Instructions      string             `json:"instructions,omitempty"`
	CacheInfo
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
