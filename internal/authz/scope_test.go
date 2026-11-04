// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidScope(t *testing.T) {
	cases := []struct {
		name  string
		scope string
		want  bool
	}{
		{"echo-read", ScopeEchoRead, true},
		{"echo-write", ScopeEchoWrite, true},
		{"echo-delete", ScopeEchoDelete, true},
		{"echo-private", ScopeEchoPrivate, true},
		{"comment-read", ScopeCommentRead, true},
		{"comment-write", ScopeCommentWrite, true},
		{"comment-review", ScopeCommentReview, true},
		{"comment-delete", ScopeCommentDelete, true},
		{"comment-settings", ScopeCommentSettings, true},
		{"file-read", ScopeFileRead, true},
		{"file-write", ScopeFileWrite, true},
		{"file-delete", ScopeFileDelete, true},
		{"connect-read", ScopeConnectRead, true},
		{"connect-write", ScopeConnectWrite, true},
		{"connect-delete", ScopeConnectDelete, true},
		{"profile-read", ScopeProfileRead, true},
		{"profile-write", ScopeProfileWrite, true},
		{"admin-site", ScopeAdminSite, true},
		{"admin-storage", ScopeAdminStorage, true},
		{"admin-auth", ScopeAdminAuth, true},
		{"admin-webhook", ScopeAdminWebhook, true},
		{"admin-agent", ScopeAdminAgent, true},
		{"admin-data", ScopeAdminData, true},
		{"admin-ops", ScopeAdminOps, true},
		{"admin-user", ScopeAdminUser, true},
		{"mcp-read", ScopeMCPRead, true},
		{"empty", "", false},
		{"unknown", "echo:destroy", false},
		{"legacy-admin-settings", "admin:settings", false},
		{"legacy-admin-token", "admin:token", false},
		{"legacy-comment-moderate", "comment:moderate", false},
		{"partial-admin", "admin", false},
		{"case-mismatch", "Echo:Read", false},
		{"audience-not-scope", AudienceCLI, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsValidScope(tc.scope))
		})
	}
}

func TestIsValidAudience(t *testing.T) {
	cases := []struct {
		name     string
		audience string
		want     bool
	}{
		{"public", AudiencePublic, true},
		{"cli", AudienceCLI, true},
		{"integration", AudienceIntegration, true},
		{"mcp-remote", AudienceMCPRemote, true},
		{"empty", "", false},
		{"unknown", "android-app", false},
		{"scope-not-audience", ScopeEchoRead, false},
		{"case-mismatch", "CLI", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsValidAudience(tc.audience))
		})
	}
}

func TestHasAdminScope(t *testing.T) {
	cases := []struct {
		name   string
		scopes []string
		want   bool
	}{
		{"nil", nil, false},
		{"empty", []string{}, false},
		{"only-admin-site", []string{ScopeAdminSite}, true},
		{"only-admin-user", []string{ScopeAdminUser}, true},
		{"only-admin-ops", []string{ScopeAdminOps}, true},
		{"non-admin-only", []string{ScopeEchoRead, ScopeCommentWrite, ScopeFileRead, ScopeMCPRead}, false},
		{"admin-mixed-with-others", []string{ScopeEchoRead, ScopeAdminUser, ScopeFileRead}, true},
		{"admin-last", []string{ScopeProfileRead, ScopeProfileWrite, ScopeAdminWebhook}, true},
		{"unknown-scopes-only", []string{"foo", "bar"}, false},
		{"unknown-with-admin", []string{"foo", ScopeAdminData}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, HasAdminScope(tc.scopes))
		})
	}
}

func TestHas(t *testing.T) {
	assert.True(t, Has([]string{ScopeEchoRead, ScopeEchoPrivate}, ScopeEchoPrivate))
	assert.False(t, Has([]string{ScopeEchoRead}, ScopeEchoPrivate))
	assert.False(t, Has(nil, ScopeEchoRead))
}

func TestAllows(t *testing.T) {
	cases := []struct {
		name     string
		held     []string
		required []string
		want     bool
	}{
		{"nothing-required", nil, nil, true},
		{"empty-required", []string{ScopeEchoRead}, []string{}, true},
		{"exact", []string{ScopeEchoRead}, []string{ScopeEchoRead}, true},
		{"superset", []string{ScopeEchoRead, ScopeEchoWrite, ScopeFileRead}, []string{ScopeEchoWrite}, true},
		{"missing-one", []string{ScopeEchoRead}, []string{ScopeEchoRead, ScopeEchoWrite}, false},
		{"write-does-not-imply-delete", []string{ScopeEchoWrite}, []string{ScopeEchoDelete}, false},
		{"read-does-not-imply-private", []string{ScopeEchoRead}, []string{ScopeEchoPrivate}, false},
		{"none-held", nil, []string{ScopeEchoRead}, false},
		{"duplicate-held", []string{ScopeEchoRead, ScopeEchoRead}, []string{ScopeEchoRead}, true},
		{"duplicate-required", []string{ScopeEchoRead}, []string{ScopeEchoRead, ScopeEchoRead}, true},
		{"required-longer", []string{ScopeEchoRead}, []string{ScopeEchoRead, ScopeEchoWrite}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Allows(tc.held, tc.required))
		})
	}
}
