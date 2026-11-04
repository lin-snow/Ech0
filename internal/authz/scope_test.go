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
		{"comment-read", ScopeCommentRead, true},
		{"comment-write", ScopeCommentWrite, true},
		{"comment-moderate", ScopeCommentMod, true},
		{"file-read", ScopeFileRead, true},
		{"file-write", ScopeFileWrite, true},
		{"connect-read", ScopeConnectRead, true},
		{"connect-write", ScopeConnectWrite, true},
		{"profile-read", ScopeProfileRead, true},
		{"profile-write", ScopeProfileWrite, true},
		{"admin-settings", ScopeAdminSettings, true},
		{"admin-user", ScopeAdminUser, true},
		{"admin-token", ScopeAdminToken, true},
		{"empty", "", false},
		{"unknown", "echo:delete", false},
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
		{"only-admin-settings", []string{ScopeAdminSettings}, true},
		{"only-admin-user", []string{ScopeAdminUser}, true},
		{"only-admin-token", []string{ScopeAdminToken}, true},
		{"non-admin-only", []string{ScopeEchoRead, ScopeCommentWrite, ScopeFileRead}, false},
		{"admin-mixed-with-others", []string{ScopeEchoRead, ScopeAdminUser, ScopeFileRead}, true},
		{"admin-last", []string{ScopeProfileRead, ScopeProfileWrite, ScopeAdminToken}, true},
		{"unknown-scopes-only", []string{"foo", "bar"}, false},
		{"unknown-with-admin", []string{"foo", ScopeAdminSettings}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, HasAdminScope(tc.scopes))
		})
	}
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
		{"none-held", nil, []string{ScopeEchoRead}, false},
		{"duplicate-held", []string{ScopeEchoRead, ScopeEchoRead}, []string{ScopeEchoRead}, true},
		{"required-longer", []string{ScopeEchoRead}, []string{ScopeEchoRead, ScopeEchoWrite}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Allows(tc.held, tc.required))
		})
	}
}
