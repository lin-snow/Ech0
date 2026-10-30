// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

func TestCreateAccessToken_RejectsUnknownScope(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-scope",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{"admin:root"},
		Audience: authModel.AudiencePublic,
	}
	if err := validateAccessTokenRequest(user, dto); err == nil {
		t.Fatal("expected error for unknown scope")
	}
}

func TestCreateAccessToken_RejectsUnknownAudience(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-audience",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authModel.ScopeEchoRead},
		Audience: "unknown-client",
	}
	if err := validateAccessTokenRequest(user, dto); err == nil {
		t.Fatal("expected error for unknown audience")
	}
}

func TestCreateAccessToken_AcceptsProfileWriteScope(t *testing.T) {
	user := userModel.User{IsAdmin: true}
	dto := &model.AccessTokenSettingDto{
		Name:     "profile-write-token",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authModel.ScopeProfileWrite},
		Audience: authModel.AudiencePublic,
	}
	if err := validateAccessTokenRequest(user, dto); err != nil {
		t.Fatalf("expected profile:write to be accepted, got error: %v", err)
	}
}

func TestCreateAccessToken_RejectsAdminScopeForNonAdminUser(t *testing.T) {
	user := userModel.User{IsAdmin: false}
	dto := &model.AccessTokenSettingDto{
		Name:     "bad-admin-scope",
		Expiry:   model.EIGHT_HOUR_EXPIRY,
		Scopes:   []string{authModel.ScopeAdminSettings},
		Audience: authModel.AudiencePublic,
	}
	if err := validateAccessTokenRequest(user, dto); err == nil {
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
			if got := accessTokenExpirySeconds(testCase.expiry); got != testCase.want {
				t.Fatalf("accessTokenExpirySeconds(%q) = %d, want %d", testCase.expiry, got, testCase.want)
			}
		})
	}
}
