// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"testing"
	"time"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeS3SettingDto(t *testing.T) {
	cases := []struct {
		name          string
		in            model.S3SettingDto
		wantSSL       bool
		wantEnd       string
		wantRegion    string
		wantCDN       string
		wantPrefix    string
		wantPathStyle bool
	}{
		{
			name:       "https endpoint forces ssl and strips scheme",
			in:         model.S3SettingDto{Provider: string(commonModel.AWS), Endpoint: "https://s3.example.com", UseSSL: false},
			wantSSL:    true,
			wantEnd:    "s3.example.com",
			wantRegion: "us-east-1",
		},
		{
			name:       "http endpoint disables ssl and strips scheme",
			in:         model.S3SettingDto{Provider: string(commonModel.MINIO), Endpoint: "http://minio.local:9000", UseSSL: true},
			wantSSL:    false,
			wantEnd:    "minio.local:9000",
			wantRegion: "us-east-1",
		},
		{
			name:       "r2 forces ssl and auto region",
			in:         model.S3SettingDto{Provider: string(commonModel.R2), Endpoint: "acc.r2.cloudflarestorage.com", UseSSL: false},
			wantSSL:    true,
			wantEnd:    "acc.r2.cloudflarestorage.com",
			wantRegion: "auto",
		},
		{
			name:       "explicit region is preserved",
			in:         model.S3SettingDto{Provider: string(commonModel.AWS), Endpoint: "s3.eu.example.com", Region: "eu-west-1", UseSSL: true},
			wantSSL:    true,
			wantEnd:    "s3.eu.example.com",
			wantRegion: "eu-west-1",
		},
		{
			name:       "other provider defaults region auto and trims cdn slash",
			in:         model.S3SettingDto{Provider: string(commonModel.OTHER), CDNURL: "https://cdn.example.com/", PathPrefix: "/uploads/"},
			wantSSL:    false,
			wantEnd:    "",
			wantRegion: "auto",
			wantCDN:    "https://cdn.example.com",
			wantPrefix: "uploads",
		},
		{
			name:          "other provider keeps use_path_style",
			in:            model.S3SettingDto{Provider: string(commonModel.OTHER), Endpoint: "s3.selfhosted.example", UsePathStyle: true},
			wantSSL:       false,
			wantEnd:       "s3.selfhosted.example",
			wantRegion:    "auto",
			wantPathStyle: true,
		},
		{
			name:       "non-other provider zeroes use_path_style",
			in:         model.S3SettingDto{Provider: string(commonModel.MINIO), Endpoint: "http://minio.local:9000", UsePathStyle: true},
			wantSSL:    false,
			wantEnd:    "minio.local:9000",
			wantRegion: "us-east-1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeS3SettingDto(&tc.in)
			assert.Equal(t, tc.wantSSL, got.UseSSL, "UseSSL")
			assert.Equal(t, tc.wantEnd, got.Endpoint, "Endpoint")
			assert.Equal(t, tc.wantRegion, got.Region, "Region")
			assert.Equal(t, tc.wantCDN, got.CDNURL, "CDNURL")
			assert.Equal(t, tc.wantPrefix, got.PathPrefix, "PathPrefix")
			assert.Equal(t, tc.wantPathStyle, got.UsePathStyle, "UsePathStyle")
			assert.Equal(t, tc.in.Provider, got.Provider)
		})
	}
}

func TestNormalizeAgentProtocol(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"openai passthrough", string(commonModel.OpenAI), string(commonModel.OpenAI)},
		{
			"openai responses passthrough",
			string(commonModel.OpenAIResponses),
			string(commonModel.OpenAIResponses),
		},
		{"anthropic passthrough", string(commonModel.Anthropic), string(commonModel.Anthropic)},
		{"retired gemini falls back to openai", "gemini", string(commonModel.OpenAI)},
		{"empty falls back to openai", "", string(commonModel.OpenAI)},
		{"unknown falls back to openai", "GPT", string(commonModel.OpenAI)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizeAgentProtocol(tc.in))
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

func TestSanitizeURLList(t *testing.T) {
	got := sanitizeURLList([]string{"https://a.example.com/", "", "   ", "/b/"})
	assert.Equal(t, []string{"https://a.example.com", "b"}, got)

	empty := sanitizeURLList([]string{"", "  "})
	assert.Equal(t, []string{}, empty)
}

func TestRemainingTTLForRevoke(t *testing.T) {
	const neverFallback = 100 * 365 * 24 * time.Hour

	t.Run("nil expiry uses never fallback", func(t *testing.T) {
		assert.Equal(t, neverFallback, remainingTTLForRevoke(nil))
	})

	t.Run("expired token still gets a minimum positive ttl", func(t *testing.T) {
		past := time.Now().UTC().Add(-2 * time.Hour).Unix()
		assert.Equal(t, time.Hour, remainingTTLForRevoke(&past))
	})

	t.Run("future token keeps remaining lifetime", func(t *testing.T) {
		future := time.Now().UTC().Add(2 * time.Hour).Unix()
		got := remainingTTLForRevoke(&future)
		assert.Greater(t, got, time.Hour)
		assert.LessOrEqual(t, got, 2*time.Hour+time.Minute)
	})
}

func TestValidateAccessTokenRequest_EdgeCases(t *testing.T) {
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
			&model.AccessTokenSettingDto{Name: "   ", Scopes: []string{authModel.ScopeEchoRead}, Audience: authModel.AudiencePublic},
			commonModel.INVALID_PARAMS_BODY,
		},
		{
			"empty scopes rejected",
			admin,
			&model.AccessTokenSettingDto{Name: "tok", Scopes: nil, Audience: authModel.AudiencePublic},
			commonModel.INVALID_PARAMS_BODY,
		},
		{
			"valid admin scope for admin accepted",
			admin,
			&model.AccessTokenSettingDto{Name: "tok", Scopes: []string{authModel.ScopeAdminSettings}, Audience: authModel.AudiencePublic},
			"",
		},
		{
			"valid non-admin scope for normal user accepted",
			normal,
			&model.AccessTokenSettingDto{Name: "tok", Scopes: []string{authModel.ScopeEchoRead}, Audience: authModel.AudienceCLI},
			"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAccessTokenRequest(tc.user, tc.dto)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Equal(t, tc.wantErr, err.Error())
		})
	}
}
