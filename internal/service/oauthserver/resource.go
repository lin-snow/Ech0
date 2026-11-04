// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"github.com/lin-snow/ech0/internal/authz"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
)

// Resource describes a protected resource exposed by this authorization
// server. Today there is exactly one; adding another means registering a
// descriptor and routing PRM/consent by path.
type Resource struct {
	Path          string
	Name          string
	Audience      string
	TokenPrefix   string
	DefaultScopes []string
}

var mcpResource = Resource{
	Path:          oauthserverModel.MCPResourcePath,
	Name:          "Ech0 MCP",
	Audience:      authz.AudienceMCPRemote,
	TokenPrefix:   "mcp",
	DefaultScopes: oauthserverModel.DefaultScopes,
}

func (s *OAuthServerService) resource() Resource { return mcpResource }
