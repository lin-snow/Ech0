// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package setting

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

type demo struct {
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

var demoSpec = Spec[demo]{
	Key:     "demo_key",
	Default: func() demo { return demo{Name: "default", Tag: "t"} },
	Normalize: func(d *demo) {
		if d.Tag == "" {
			d.Tag = "normalized"
		}
	},
}

type boomStore struct{}

func (boomStore) Get(context.Context, string) (string, error) { return "", errors.New("boom") }
func (boomStore) Set(context.Context, string, string) error   { return nil }
func (boomStore) Delete(context.Context, string) error        { return nil }

func TestGet_MissReturnsNormalizedDefault(t *testing.T) {
	got, err := Get(context.Background(), kvstore.NewMemory(), demoSpec)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Name != "default" || got.Tag != "t" {
		t.Fatalf("want default{default,t}, got %+v", got)
	}
}

func TestGet_ReadsStoredValue(t *testing.T) {
	kv := kvstore.NewMemory()
	_ = kv.Set(context.Background(), demoSpec.Key, `{"name":"stored","tag":"x"}`)
	got, err := Get(context.Background(), kv, demoSpec)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Name != "stored" || got.Tag != "x" {
		t.Fatalf("want {stored,x}, got %+v", got)
	}
}

func TestGet_AppliesNormalizeToStoredValue(t *testing.T) {
	kv := kvstore.NewMemory()
	_ = kv.Set(context.Background(), demoSpec.Key, `{"name":"stored"}`)
	got, err := Get(context.Background(), kv, demoSpec)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Tag != "normalized" {
		t.Fatalf("want tag normalized, got %q", got.Tag)
	}
}

func TestGet_BackendErrorReturnsDefaultAndError(t *testing.T) {
	got, err := Get(context.Background(), boomStore{}, demoSpec)
	if err == nil {
		t.Fatal("want backend error to propagate")
	}
	if got.Name != "default" {
		t.Fatalf("want usable default on error, got %+v", got)
	}
}

func TestGet_UnparseableValueReturnsDefaultAndError(t *testing.T) {
	kv := kvstore.NewMemory()
	_ = kv.Set(context.Background(), demoSpec.Key, `{not valid json`)
	got, err := Get(context.Background(), kv, demoSpec)
	if err == nil {
		t.Fatal("want unmarshal error to propagate")
	}
	if got.Name != "default" || got.Tag != "t" {
		t.Fatalf("want usable default on parse failure, got %+v", got)
	}
}

func TestSet_RoundTrip(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := Set(context.Background(), kv, demoSpec, demo{Name: "x", Tag: "y"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := Get(context.Background(), kv, demoSpec)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "x" || got.Tag != "y" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestSet_AppliesNormalizeOnWrite(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := Set(context.Background(), kv, demoSpec, demo{Name: "x"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	raw, err := kv.Get(context.Background(), demoSpec.Key)
	if err != nil {
		t.Fatalf("get raw: %v", err)
	}
	if !strings.Contains(raw, "normalized") {
		t.Fatalf("expected normalized tag persisted, got %q", raw)
	}
}

func TestSeed_PopulatesMissingKeys(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := Seed(context.Background(), kv); err != nil {
		t.Fatalf("seed: %v", err)
	}
	for _, key := range []string{
		commonModel.SystemSettingsKey,
		commonModel.ServerURLKey,
		commonModel.OAuth2SettingKey,
		commonModel.S3SettingKey,
		commonModel.PasskeySettingKey,
		commonModel.AgentSettingKey,
		commonModel.SnapshotScheduleKey,
		commonModel.EmbeddingSettingKey,
		commentModel.CommentSystemSettingKey,
	} {
		if _, err := kv.Get(context.Background(), key); err != nil {
			t.Errorf("key %q not seeded: %v", key, err)
		}
	}
}

func TestSeed_IdempotentDoesNotOverwrite(t *testing.T) {
	kv := kvstore.NewMemory()
	_ = kv.Set(context.Background(), commonModel.SystemSettingsKey, `{"site_title":"custom"}`)
	if err := Seed(context.Background(), kv); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got, err := Get(context.Background(), kv, System)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.SiteTitle != "custom" {
		t.Fatalf("seed must not overwrite user value, got SiteTitle=%q", got.SiteTitle)
	}
}

func TestSeed_CommentDefaultsSMTPPort(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := Seed(context.Background(), kv); err != nil {
		t.Fatalf("seed: %v", err)
	}
	got, err := Get(context.Background(), kv, Comment)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.EnableComment || got.EmailNotify.SMTPPort != 587 {
		t.Fatalf("want EnableComment + SMTPPort 587, got %+v", got)
	}
}

func TestSeed_PasskeyMigratesFromLegacyOAuth2(t *testing.T) {
	kv := kvstore.NewMemory()
	_ = kv.Set(context.Background(), commonModel.OAuth2SettingKey,
		`{"webauthn_rp_id":"example.com","webauthn_allowed_origins":["https://example.com"]}`)

	if err := Seed(context.Background(), kv); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := Get(context.Background(), kv, Passkey)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.WebAuthnRPID != "example.com" {
		t.Fatalf("want migrated RPID example.com, got %q", got.WebAuthnRPID)
	}
	if len(got.WebAuthnAllowedOrigins) != 1 || got.WebAuthnAllowedOrigins[0] != "https://example.com" {
		t.Fatalf("want migrated origins, got %+v", got.WebAuthnAllowedOrigins)
	}
}

var _ = []seedable{System, serverURLSeed{}, Spec[settingModel.AgentSetting]{}}
