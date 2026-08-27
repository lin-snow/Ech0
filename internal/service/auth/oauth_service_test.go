// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package auth

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

func oauthKVWithRedirect(t *testing.T, redirectURI string) kvstore.Store {
	t.Helper()
	kv := kvstore.NewMemory()
	raw, err := json.Marshal(settingModel.OAuth2Setting{RedirectURI: redirectURI})
	if err != nil {
		t.Fatalf("marshal oauth2 setting: %v", err)
	}
	if err := kv.Set(context.Background(), commonModel.OAuth2SettingKey, string(raw)); err != nil {
		t.Fatalf("seed oauth2 setting: %v", err)
	}
	return kv
}

func TestParseAndValidateClientRedirect_Allowed(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	u, err := svc.parseAndValidateClientRedirect("https://app.example.com/auth?from=test")
	if err != nil {
		t.Fatalf("expected allow redirect, got err: %v", err)
	}
	if u.Host != "app.example.com" {
		t.Fatalf("unexpected host: %s", u.Host)
	}
}

func TestParseAndValidateClientRedirect_Denied(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	_, err := svc.parseAndValidateClientRedirect("https://evil.example.net/auth")
	if err == nil {
		t.Fatalf("expected deny redirect")
	}
}

func TestParseAndValidateClientRedirect_PathMismatchDenied(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = []string{"https://app.example.com/auth"}

	svc := &AuthService{}
	_, err := svc.parseAndValidateClientRedirect(
		"https://app.example.com/attacker-chosen-path",
	)
	if err == nil {
		t.Fatalf("expected deny redirect with mismatched path")
	}
}

func TestParseAndValidateClientRedirect_ImplicitSelfAllowed(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = nil

	svc := &AuthService{durableKV: oauthKVWithRedirect(t, "https://m.example.com/oauth/github/callback")}

	for _, target := range []string{
		"https://m.example.com/panel",
		"https://m.example.com/auth?from=test",
	} {
		if _, err := svc.parseAndValidateClientRedirect(target); err != nil {
			t.Fatalf("expected implicit allow for %s, got err: %v", target, err)
		}
	}
}

func TestParseAndValidateClientRedirect_ImplicitSelfOnlyFixedPaths(t *testing.T) {
	cfg := config.Config()
	cfg.Auth.Redirect.AllowedReturnURLs = nil

	svc := &AuthService{durableKV: oauthKVWithRedirect(t, "https://m.example.com/oauth/github/callback")}

	for _, target := range []string{
		"https://m.example.com/attacker-chosen-path",
		"https://evil.example.net/panel",
	} {
		if _, err := svc.parseAndValidateClientRedirect(target); err == nil {
			t.Fatalf("expected deny for %s", target)
		}
	}
}
