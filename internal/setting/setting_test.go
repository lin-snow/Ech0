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
	got, err := demoSpec.Get(context.Background(), kvstore.NewMemory())
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
	got, err := demoSpec.Get(context.Background(), kv)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Name != "stored" || got.Tag != "x" {
		t.Fatalf("want {stored,x}, got %+v", got)
	}
}

func TestGet_AbsentFieldKeepsDefault(t *testing.T) {
	kv := kvstore.NewMemory()
	_ = kv.Set(context.Background(), demoSpec.Key, `{"name":"stored"}`)
	got, err := demoSpec.Get(context.Background(), kv)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Name != "stored" || got.Tag != "t" {
		t.Fatalf("want absent field to keep default tag t, got %+v", got)
	}
}

func TestGet_BackendErrorReturnsDefaultAndError(t *testing.T) {
	got, err := demoSpec.Get(context.Background(), boomStore{})
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
	got, err := demoSpec.Get(context.Background(), kv)
	if err == nil {
		t.Fatal("want unmarshal error to propagate")
	}
	if got.Name != "default" || got.Tag != "t" {
		t.Fatalf("want usable default on parse failure, got %+v", got)
	}
}

func TestSet_RoundTrip(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := demoSpec.Set(context.Background(), kv, demo{Name: "x", Tag: "y"}); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := demoSpec.Get(context.Background(), kv)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "x" || got.Tag != "y" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestSet_AppliesNormalizeOnWrite(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := demoSpec.Set(context.Background(), kv, demo{Name: "x"}); err != nil {
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

func TestSet_RejectsInvalidValue(t *testing.T) {
	kv := kvstore.NewMemory()
	spec := Spec[demo]{
		Key:     "validated_key",
		Default: func() demo { return demo{} },
		Validate: func(d demo) error {
			if d.Name == "" {
				return errors.New("name required")
			}
			return nil
		},
	}
	if err := spec.Set(context.Background(), kv, demo{Tag: "t"}); err == nil {
		t.Fatal("want validation error")
	}
	if _, err := kv.Get(context.Background(), spec.Key); !errors.Is(err, kvstore.ErrNotFound) {
		t.Fatalf("invalid value must not be persisted, got err=%v", err)
	}
	if err := spec.Set(context.Background(), kv, demo{Name: "ok"}); err != nil {
		t.Fatalf("valid set: %v", err)
	}
}

func TestMigrate_DropsLegacyServerURLKey(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := kv.Set(context.Background(), "server_url", "https://legacy.example.com"); err != nil {
		t.Fatalf("seed legacy key: %v", err)
	}
	if err := Migrate(context.Background(), kv); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := kv.Get(context.Background(), "server_url"); !errors.Is(err, kvstore.ErrNotFound) {
		t.Fatalf("want legacy server_url removed, got err=%v", err)
	}
}

func TestMigrate_DoesNotMaterializeDefaults(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := Migrate(context.Background(), kv); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	for _, key := range []string{
		commonModel.SystemSettingsKey,
		commonModel.OAuth2SettingKey,
		commonModel.S3SettingKey,
		commonModel.PasskeySettingKey,
		commonModel.AgentSettingKey,
		commonModel.SnapshotScheduleKey,
		commonModel.EmbeddingSettingKey,
		commentModel.CommentSystemSettingKey,
	} {
		if _, err := kv.Get(context.Background(), key); !errors.Is(err, kvstore.ErrNotFound) {
			t.Errorf("key %q must stay unwritten until a write, got err=%v", key, err)
		}
	}
}

func TestSystem_UnstoredDefaults(t *testing.T) {
	got, err := System.Get(context.Background(), kvstore.NewMemory())
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.SiteTitle == "" {
		t.Fatal("want a non-empty default site title")
	}
	if !got.EnableAgentDiscovery {
		t.Fatal("agent discovery must default on")
	}
}

func TestComment_GetDefaultsSMTPPort(t *testing.T) {
	kv := kvstore.NewMemory()
	if err := Migrate(context.Background(), kv); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	got, err := Comment.Get(context.Background(), kv)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.EnableComment || got.EmailNotify.SMTPPort != 587 {
		t.Fatalf("want EnableComment + SMTPPort 587, got %+v", got)
	}
}

func TestMigrate_PasskeyFromLegacyOAuth2(t *testing.T) {
	kv := kvstore.NewMemory()
	_ = kv.Set(context.Background(), commonModel.OAuth2SettingKey,
		`{"webauthn_rp_id":"example.com","webauthn_allowed_origins":["https://example.com"]}`)

	if err := Migrate(context.Background(), kv); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	got, err := Passkey.Get(context.Background(), kv)
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

func TestMigrate_AgentDiscoveryFlag(t *testing.T) {
	cases := []struct {
		name      string
		store     string
		want      bool
		rewritten bool
	}{
		{
			name:      "legacy-opt-out",
			store:     `{"site_title":"Legacy","disable_agent_discovery":true}`,
			want:      false,
			rewritten: true,
		},
		{
			name:      "legacy-enabled",
			store:     `{"disable_agent_discovery":false}`,
			want:      true,
			rewritten: true,
		},
		{
			name:  "flag-absent",
			store: `{"site_title":"Legacy"}`,
			want:  true,
		},
		{
			name:  "current-flag-wins",
			store: `{"enable_agent_discovery":false,"disable_agent_discovery":true}`,
			want:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kv := kvstore.NewMemory()
			if err := kv.Set(context.Background(), commonModel.SystemSettingsKey, tc.store); err != nil {
				t.Fatalf("write blob: %v", err)
			}
			if err := Migrate(context.Background(), kv); err != nil {
				t.Fatalf("migrate: %v", err)
			}
			raw, err := kv.Get(context.Background(), commonModel.SystemSettingsKey)
			if err != nil {
				t.Fatalf("read back: %v", err)
			}
			if tc.rewritten {
				if strings.Contains(raw, "disable_agent_discovery") || !strings.Contains(raw, `"enable_agent_discovery":`) {
					t.Fatalf("want the opt-out rewritten to the positive flag, got %s", raw)
				}
			} else if raw != tc.store {
				t.Fatalf("blob must stay untouched, got %s", raw)
			}

			got, err := System.Get(context.Background(), kv)
			if err != nil {
				t.Fatalf("get: %v", err)
			}
			if got.EnableAgentDiscovery != tc.want {
				t.Fatalf("EnableAgentDiscovery = %v, want %v (blob %s)", got.EnableAgentDiscovery, tc.want, raw)
			}
		})
	}
}
