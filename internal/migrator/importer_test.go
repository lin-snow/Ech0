// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"testing"
)

func TestParseMigratedS3Setting(t *testing.T) {
	report := map[string]any{
		"source_s3_setting": map[string]any{
			"enable":      true,
			"provider":    "r2",
			"endpoint":    "example.r2.cloudflarestorage.com",
			"access_key":  "ak",
			"secret_key":  "sk",
			"bucket_name": "bucket-a",
			"region":      "auto",
			"use_ssl":     true,
		},
	}
	setting, ok, err := parseMigratedS3Setting(report)
	if err != nil {
		t.Fatalf("parseMigratedS3Setting returned error: %v", err)
	}
	if !ok || setting == nil {
		t.Fatalf("expected valid migrated s3 setting")
	}
	if setting.Provider != "r2" || setting.BucketName != "bucket-a" {
		t.Fatalf("unexpected setting parsed: %+v", setting)
	}
}

func TestParseMigratedS3Setting_Invalid(t *testing.T) {
	report := map[string]any{
		"source_s3_setting": map[string]any{
			"provider": "r2",
		},
	}
	setting, ok, err := parseMigratedS3Setting(report)
	if err != nil {
		t.Fatalf("parseMigratedS3Setting returned error: %v", err)
	}
	if ok || setting != nil {
		t.Fatalf("expected invalid migrated s3 setting to be ignored")
	}
}

func TestParseMigratedCommonSettings(t *testing.T) {
	report := map[string]any{
		"source_system_setting": map[string]any{
			"site_title":     "My Echo",
			"allow_register": true,
			"server_url":     "https://example.com",
			"footer_content": "hello",
			"footer_link":    "https://example.com/about",
			"meting_api":     "https://meting.example.com",
			"custom_css":     "body{color:#fff;}",
			"custom_js":      "console.log('ok')",
			"server_logo":    "",
			"server_name":    "ech0",
			"ICP_number":     "",
		},
		"source_comment_setting": map[string]any{
			"enable_comment":   true,
			"require_approval": true,
			"captcha_enabled":  false,
		},
		"source_oauth2_setting": map[string]any{
			"enable":        true,
			"provider":      "github",
			"client_id":     "id-a",
			"client_secret": "sec-a",
		},
	}

	systemSetting, ok, err := parseMigratedSystemSetting(report)
	if err != nil || !ok || systemSetting == nil {
		t.Fatalf("parseMigratedSystemSetting failed: ok=%v err=%v", ok, err)
	}
	if systemSetting.SiteTitle != "My Echo" || !systemSetting.AllowRegister {
		t.Fatalf("unexpected system setting: %+v", systemSetting)
	}

	commentSetting, ok, err := parseMigratedCommentSetting(report)
	if err != nil || !ok || commentSetting == nil {
		t.Fatalf("parseMigratedCommentSetting failed: ok=%v err=%v", ok, err)
	}
	if !commentSetting.EnableComment || !commentSetting.RequireApproval {
		t.Fatalf("unexpected comment setting: %+v", commentSetting)
	}

	oauth2Setting, ok, err := parseMigratedOAuth2Setting(report)
	if err != nil || !ok || oauth2Setting == nil {
		t.Fatalf("parseMigratedOAuth2Setting failed: ok=%v err=%v", ok, err)
	}
	if !oauth2Setting.Enable || oauth2Setting.Provider != "github" {
		t.Fatalf("unexpected oauth2 setting: %+v", oauth2Setting)
	}
}

func TestParseMigratedCommonSettings_InvalidIgnored(t *testing.T) {
	report := map[string]any{
		"source_system_setting":  "not-json-object",
		"source_comment_setting": 1,
		"source_oauth2_setting":  true,
	}

	if setting, ok, err := parseMigratedSystemSetting(report); err == nil || ok || setting != nil {
		t.Fatalf("expected invalid system setting to return error and be ignored, got ok=%v err=%v", ok, err)
	}
	if setting, ok, err := parseMigratedCommentSetting(report); err == nil || ok || setting != nil {
		t.Fatalf("expected invalid comment setting to return error and be ignored, got ok=%v err=%v", ok, err)
	}
	if setting, ok, err := parseMigratedOAuth2Setting(report); err == nil || ok || setting != nil {
		t.Fatalf("expected invalid oauth2 setting to return error and be ignored, got ok=%v err=%v", ok, err)
	}
}
