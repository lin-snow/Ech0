// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package authz is Ech0's pure authorization vocabulary and policy. It has no
// I/O and no domain dependencies: middleware, the MCP server and the OAuth
// authorization server all decide access through this package.
package authz

import "strings"

const (
	// Content scopes. Write covers create/update; delete is a separate,
	// irreversible capability. echo:private is additive to echo:read.
	ScopeEchoRead    = "echo:read"
	ScopeEchoWrite   = "echo:write"
	ScopeEchoDelete  = "echo:delete"
	ScopeEchoPrivate = "echo:private"

	ScopeCommentRead     = "comment:read"
	ScopeCommentWrite    = "comment:write"
	ScopeCommentReview   = "comment:review"
	ScopeCommentDelete   = "comment:delete"
	ScopeCommentSettings = "comment:settings"

	ScopeFileRead   = "file:read"
	ScopeFileWrite  = "file:write"
	ScopeFileDelete = "file:delete"

	ScopeConnectRead   = "connect:read"
	ScopeConnectWrite  = "connect:write"
	ScopeConnectDelete = "connect:delete"

	// Profile scopes cover the public profile only. Passkey and OAuth identity
	// binding are browser-session operations, never delegated to tokens.
	ScopeProfileRead  = "profile:read"
	ScopeProfileWrite = "profile:write"

	// Administrative scopes are split by blast radius. A token that can manage
	// webhooks must not be able to change S3 credentials, SSO or migration data.
	ScopeAdminSite    = "admin:site"
	ScopeAdminStorage = "admin:storage"
	ScopeAdminAuth    = "admin:auth"
	ScopeAdminWebhook = "admin:webhook"
	ScopeAdminAgent   = "admin:agent"
	ScopeAdminData    = "admin:data"
	ScopeAdminOps     = "admin:ops"
	ScopeAdminUser    = "admin:user"

	// ScopeMCPRead only grants discovery of the MCP manifest; invoking a tool
	// still requires that tool's own scope.
	ScopeMCPRead = "mcp:read"

	AudiencePublic      = "public-client"
	AudienceCLI         = "cli"
	AudienceIntegration = "integration"
	AudienceMCPRemote   = "mcp-remote"
)

var validScopes = map[string]struct{}{
	ScopeEchoRead:        {},
	ScopeEchoWrite:       {},
	ScopeEchoDelete:      {},
	ScopeEchoPrivate:     {},
	ScopeCommentRead:     {},
	ScopeCommentWrite:    {},
	ScopeCommentReview:   {},
	ScopeCommentDelete:   {},
	ScopeCommentSettings: {},
	ScopeFileRead:        {},
	ScopeFileWrite:       {},
	ScopeFileDelete:      {},
	ScopeConnectRead:     {},
	ScopeConnectWrite:    {},
	ScopeConnectDelete:   {},
	ScopeProfileRead:     {},
	ScopeProfileWrite:    {},
	ScopeAdminSite:       {},
	ScopeAdminStorage:    {},
	ScopeAdminAuth:       {},
	ScopeAdminWebhook:    {},
	ScopeAdminAgent:      {},
	ScopeAdminData:       {},
	ScopeAdminOps:        {},
	ScopeAdminUser:       {},
	ScopeMCPRead:         {},
}

// scopeOrder fixes the advertisement order of AllScopes so discovery metadata
// and consent UIs stay deterministic.
var scopeOrder = []string{
	ScopeEchoRead,
	ScopeEchoWrite,
	ScopeEchoDelete,
	ScopeEchoPrivate,
	ScopeCommentRead,
	ScopeCommentWrite,
	ScopeCommentReview,
	ScopeCommentDelete,
	ScopeCommentSettings,
	ScopeFileRead,
	ScopeFileWrite,
	ScopeFileDelete,
	ScopeConnectRead,
	ScopeConnectWrite,
	ScopeConnectDelete,
	ScopeProfileRead,
	ScopeProfileWrite,
	ScopeAdminSite,
	ScopeAdminStorage,
	ScopeAdminAuth,
	ScopeAdminWebhook,
	ScopeAdminAgent,
	ScopeAdminData,
	ScopeAdminOps,
	ScopeAdminUser,
	ScopeMCPRead,
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

// HasAdminScope reports whether the set contains any administrative
// capability. Admin scopes share the `admin:` prefix by construction; using
// the prefix keeps this check correct when a new admin domain is added.
func HasAdminScope(scopes []string) bool {
	for _, scope := range scopes {
		if strings.HasPrefix(scope, "admin:") {
			return true
		}
	}
	return false
}

// Has reports whether held contains scope. It is the single-scope fast path
// for hot readers such as private-echo visibility.
func Has(held []string, scope string) bool {
	for _, have := range held {
		if have == scope {
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
