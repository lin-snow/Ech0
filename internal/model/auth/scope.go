// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

const (
	TokenTypeSession = "session"
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

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
