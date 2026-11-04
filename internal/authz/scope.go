// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package authz is Ech0's pure authorization vocabulary and policy. It has no
// I/O and no domain dependencies: middleware, the MCP server and the OAuth
// authorization server all decide access through this package.
package authz

const (
	ScopeEchoRead       = "echo:read"
	ScopeEchoWrite      = "echo:write"
	ScopeCommentRead    = "comment:read"
	ScopeCommentWrite   = "comment:write"
	ScopeCommentMod     = "comment:moderate"
	ScopeFileRead       = "file:read"
	ScopeFileWrite      = "file:write"
	ScopeConnectRead    = "connect:read"
	ScopeConnectWrite   = "connect:write"
	ScopeProfileRead    = "profile:read"
	ScopeProfileWrite   = "profile:write"
	ScopeAdminSettings  = "admin:settings"
	ScopeAdminUser      = "admin:user"
	ScopeAdminToken     = "admin:token"
	AudiencePublic      = "public-client"
	AudienceCLI         = "cli"
	AudienceIntegration = "integration"
	AudienceMCPRemote   = "mcp-remote"
)

var validScopes = map[string]struct{}{
	ScopeEchoRead:      {},
	ScopeEchoWrite:     {},
	ScopeCommentRead:   {},
	ScopeCommentWrite:  {},
	ScopeCommentMod:    {},
	ScopeFileRead:      {},
	ScopeFileWrite:     {},
	ScopeConnectRead:   {},
	ScopeConnectWrite:  {},
	ScopeProfileRead:   {},
	ScopeProfileWrite:  {},
	ScopeAdminSettings: {},
	ScopeAdminUser:     {},
	ScopeAdminToken:    {},
}

// scopeOrder fixes the advertisement order of AllScopes so discovery metadata
// and consent UIs stay deterministic.
var scopeOrder = []string{
	ScopeEchoRead,
	ScopeEchoWrite,
	ScopeCommentRead,
	ScopeCommentWrite,
	ScopeCommentMod,
	ScopeFileRead,
	ScopeFileWrite,
	ScopeConnectRead,
	ScopeConnectWrite,
	ScopeProfileRead,
	ScopeProfileWrite,
	ScopeAdminSettings,
	ScopeAdminUser,
	ScopeAdminToken,
}

// AllScopes returns every scope in the fixed vocabulary order.
func AllScopes() []string {
	return append([]string(nil), scopeOrder...)
}

var validAudiences = map[string]struct{}{
	AudiencePublic:      {},
	AudienceCLI:         {},
	AudienceIntegration: {},
	AudienceMCPRemote:   {},
}

func IsValidScope(scope string) bool {
	_, ok := validScopes[scope]
	return ok
}

func IsValidAudience(audience string) bool {
	_, ok := validAudiences[audience]
	return ok
}

func HasAdminScope(scopes []string) bool {
	for _, scope := range scopes {
		switch scope {
		case ScopeAdminSettings, ScopeAdminUser, ScopeAdminToken:
			return true
		}
	}
	return false
}

// Allows reports whether held contains every required scope. Scopes are a
// small fixed vocabulary, so the nested scan is cheaper than allocating a set
// on every request.
func Allows(held, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(required) > len(held) {
		return false
	}
	for _, want := range required {
		found := false
		for _, have := range held {
			if have == want {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
