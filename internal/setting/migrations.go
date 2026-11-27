// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package setting

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

// migrations are one-shot wire-format moves that run before the instance
// starts serving. They stay separate from Spec so a default change never
// silently rewrites a stored value: only a rename that would otherwise change
// meaning belongs here.
var migrations = []func(ctx context.Context, kv kvstore.Store) error{
	migratePasskeyFromLegacy,
	migrateAgentDiscoveryFlag,
	dropLegacyServerURLKey,
}

// migratePasskeyFromLegacy moves the WebAuthn fields out of the pre-split
// OAuth2 blob. It writes only when the Passkey block has never been stored.
func migratePasskeyFromLegacy(ctx context.Context, kv kvstore.Store) error {
	if _, err := kv.Get(ctx, commonModel.PasskeySettingKey); err == nil {
		return nil
	} else if !errors.Is(err, kvstore.ErrNotFound) {
		return err
	}

	raw, err := kv.Get(ctx, commonModel.OAuth2SettingKey)
	if errors.Is(err, kvstore.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	var legacy struct {
		WebAuthnRPID           string   `json:"webauthn_rp_id"`
		WebAuthnAllowedOrigins []string `json:"webauthn_allowed_origins"`
	}
	if err := json.Unmarshal([]byte(raw), &legacy); err != nil {
		// An unreadable legacy blob must not block startup; the OAuth2 spec
		// reports the damage on its own read.
		return nil
	}

	migrated := settingModel.PasskeySetting{
		WebAuthnRPID:           strings.TrimSpace(legacy.WebAuthnRPID),
		WebAuthnAllowedOrigins: sanitizeURLList(legacy.WebAuthnAllowedOrigins),
	}
	if migrated.WebAuthnRPID == "" && len(migrated.WebAuthnAllowedOrigins) == 0 {
		return nil
	}
	return Passkey.Set(ctx, kv, migrated)
}

// migrateAgentDiscoveryFlag rewrites the opt-out field into its positive form.
// A stored `disable_agent_discovery: true` would otherwise read as "not set" on
// the new field, whose default is on, and silently reopen the discovery
// surface on an instance that had turned it off.
func migrateAgentDiscoveryFlag(ctx context.Context, kv kvstore.Store) error {
	raw, err := kv.Get(ctx, commonModel.SystemSettingsKey)
	if errors.Is(err, kvstore.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	blob := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(raw), &blob); err != nil {
		// System.Get reports the unreadable blob to every caller that needs it.
		return nil
	}
	if _, ok := blob["enable_agent_discovery"]; ok {
		return nil
	}
	legacy, ok := blob["disable_agent_discovery"]
	if !ok {
		return nil
	}

	var disabled bool
	if err := json.Unmarshal(legacy, &disabled); err != nil {
		return nil
	}
	enabled, err := json.Marshal(!disabled)
	if err != nil {
		return err
	}
	delete(blob, "disable_agent_discovery")
	blob["enable_agent_discovery"] = enabled

	encoded, err := json.Marshal(blob)
	if err != nil {
		return err
	}
	return kv.Set(ctx, commonModel.SystemSettingsKey, string(encoded))
}

// dropLegacyServerURLKey removes the row older versions wrote as a mirror of
// SystemSetting.ServerURL. The system block is the only source now.
func dropLegacyServerURLKey(ctx context.Context, kv kvstore.Store) error {
	const legacyKey = "server_url"
	if err := kv.Delete(ctx, legacyKey); err != nil && !errors.Is(err, kvstore.ErrNotFound) {
		return err
	}
	return nil
}
