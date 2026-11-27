// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package authz is Ech0's pure authorization vocabulary and policy. It has no
// I/O and no domain dependencies: middleware, the MCP server and the OAuth
// authorization server all decide access through this package.
package authz

import "github.com/lin-snow/ech0/pkg/role"

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

// ScopeCount is the size of the scope vocabulary without allocating a copy, so
// request caps can stay tied to the vocabulary instead of a magic number.
func ScopeCount() int {
	return len(scopeOrder)
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

// ScopesForRole is the capability set of an interactive session. Access tokens
// are never expanded: their scopes were fixed at mint time. Owner equals admin
// on purpose — the owner's extra power is managing administrators and
// transferring ownership, which are rank questions inside user management, not
// a larger capability set.
func ScopesForRole(r role.Role) []string {
	switch r {
	case role.Owner:
		return append([]string(nil), adminScopes...)
	case role.Admin:
		return append([]string(nil), adminScopes...)
	case role.Member:
		return append([]string(nil), memberScopes...)
	default:
		return nil
	}
}

// memberScopes is the baseline every signed-in member holds. It covers creating
// and reading content plus the visibility modifier for private rows; nothing
// governed by rank is included.
var memberScopes = []string{
	ScopeEchoRead, ScopeEchoWrite, ScopeEchoDelete, ScopeEchoPrivate,
	ScopeCommentRead, ScopeCommentWrite,
	ScopeFileRead, ScopeFileWrite, ScopeFileDelete,
	ScopeConnectRead,
	ScopeProfileRead, ScopeProfileWrite,
	ScopeMCPRead,
}

// adminScopes is memberScopes plus every management domain. Splitting by blast
// radius means granting one admin surface never grants the others.
var adminScopes = append(append([]string(nil), memberScopes...),
	ScopeCommentReview, ScopeCommentDelete, ScopeCommentSettings,
	ScopeConnectWrite, ScopeConnectDelete,
	ScopeAdminSite, ScopeAdminStorage, ScopeAdminAuth, ScopeAdminWebhook,
	ScopeAdminAgent, ScopeAdminData, ScopeAdminOps, ScopeAdminUser,
)

// MinimumRole reports the lowest role whose session scope set contains scope.
// It is the single source of truth for "is this an administrative
// capability?", so endpoint declarations and transport rules cannot drift from
// the role derivation above. An unknown scope returns the empty role.
func MinimumRole(scope string) role.Role {
	for _, candidate := range []role.Role{role.Member, role.Admin, role.Owner} {
		if Allows(ScopesForRole(candidate), []string{scope}) {
			return candidate
		}
	}
	return ""
}

// HasAdminScope reports whether the set contains any capability that only an
// administrator holds. It is derived from the role scope sets, not from a
// naming convention: a new management scope cannot silently escape the gate.
func HasAdminScope(scopes []string) bool {
	for _, scope := range scopes {
		if MinimumRole(scope).AtLeast(role.Admin) {
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

// Allows reports whether held contains every required scope. scopes are a
// small fixed vocabulary, so the nested scan is cheaper than allocating a set
// on every request. No length shortcut: required is a requirement list, not a
// set, so duplicates must not turn into a spurious denial.
func Allows(held, required []string) bool {
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
