// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package access

import (
	"testing"

	"github.com/lin-snow/ech0/internal/authz"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAccessToken_RejectsUnknownScope(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-scope",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{"admin:root"},
		Audience: authz.AudiencePublic,
	}
	if err := validateRequest(user, dto); err == nil {
		t.Fatal("expected error for unknown scope")
	}
}

func TestCreateAccessToken_RejectsUnknownAudience(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-audience",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authz.ScopeEchoRead},
		Audience: "unknown-client",
	}
	if err := validateRequest(user, dto); err == nil {
		t.Fatal("expected error for unknown audience")
	}
}

func TestCreateAccessToken_AcceptsProfileWriteScope(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "profile-write-token",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authz.ScopeProfileWrite},
		Audience: authz.AudiencePublic,
	}
	if err := validateRequest(user, dto); err != nil {
		t.Fatalf("expected profile:write to be accepted, got error: %v", err)
	}
}

func TestCreateAccessToken_RejectsAdminScopeForNonAdminUser(t *testing.T) {
	user := userModel.User{IsAdmin: false}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-admin-scope",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authz.ScopeAdminSettings},
		Audience: authz.AudiencePublic,
	}
	if err := validateRequest(user, dto); err == nil {
		t.Fatal("expected error for admin scope on non-admin user")
	}
}

func TestAccessTokenExpirySeconds(t *testing.T) {
	cases := []struct {
		name   string
		expiry string
		want   int64
	}{
		{name: "eight hours", expiry: model.EIGHT_HOUR_EXPIRY, want: 8 * 3600},
		{name: "one month", expiry: model.ONE_MONTH_EXPIRY, want: 30 * 24 * 3600},
		{name: "never", expiry: model.NEVER_EXPIRY, want: 0},
		{name: "unknown falls back to eight hours", expiry: "bogus", want: 8 * 3600},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := expirySeconds(testCase.expiry); got != testCase.want {
				t.Fatalf("expirySeconds(%q) = %d, want %d", testCase.expiry, got, testCase.want)
			}
		})
	}
}

func TestNormalizeScopes(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"dedup preserves first-seen order", []string{"a", "b", "a", "c", "b"}, []string{"a", "b", "c"}},
		{"no duplicates is unchanged", []string{"x", "y"}, []string{"x", "y"}},
		{"empty stays empty", []string{}, []string{}},
		{"nil stays empty", nil, []string{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeScopes(tc.in))
		})
	}
}

func TestValidateRequest_EdgeCases(t *testing.T) {
	admin := userModel.User{IsAdmin: true}
	normal := userModel.User{IsAdmin: false}

	cases := []struct {
		name    string
		user    userModel.User
		dto     *model.AccessTokenSettingDto
		wantErr string
	}{
		{"nil dto rejected", admin, nil, commonModel.INVALID_PARAMS_BODY},
		{
			"blank name rejected",
			admin,
			&model.AccessTokenSettingDto{Name: "   ", Scopes: []string{authz.ScopeEchoRead}, Audience: authz.AudiencePublic},
			commonModel.INVALID_PARAMS_BODY,
		},
		{
			"empty scopes rejected",
			admin,
			&model.AccessTokenSettingDto{Name: "tok", Scopes: nil, Audience: authz.AudiencePublic},
			commonModel.INVALID_PARAMS_BODY,
		},
		{
			"valid admin scope for admin accepted",
			admin,
			&model.AccessTokenSettingDto{Name: "tok", Scopes: []string{authz.ScopeAdminSettings}, Audience: authz.AudiencePublic},
			"",
		},
		{
			"valid non-admin scope for normal user accepted",
			normal,
			&model.AccessTokenSettingDto{Name: "tok", Scopes: []string{authz.ScopeEchoRead}, Audience: authz.AudienceCLI},
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRequest(tc.user, tc.dto)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, tc.wantErr, err.Error())
		})
	}
}
