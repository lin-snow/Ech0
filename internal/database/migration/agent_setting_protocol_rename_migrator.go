// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration

import (
	"encoding/json"
	"errors"
	"fmt"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"gorm.io/gorm"
)

type agentSettingProtocolRenameMigrator struct{}

func NewAgentSettingProtocolRenameMigrator() Migrator {
	return &agentSettingProtocolRenameMigrator{}
}

func (m *agentSettingProtocolRenameMigrator) Name() string {
	return "agent_setting_protocol_rename_migrator"
}

func (m *agentSettingProtocolRenameMigrator) Key() string {
	return commonModel.AgentSettingProtocolRenamedKey
}

func (m *agentSettingProtocolRenameMigrator) CanRerun() bool {
	return false
}

func (m *agentSettingProtocolRenameMigrator) Migrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	var kv commonModel.KeyValue
	err := db.Where("key = ?", commonModel.AgentSettingKey).First(&kv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	raw := map[string]any{}
	if err := json.Unmarshal([]byte(kv.Value), &raw); err != nil {
		return fmt.Errorf("agent setting json invalid: %w", err)
	}

	if _, ok := raw["provider"]; !ok {
		return nil
	}

	if _, ok := raw["protocol"]; !ok {
		raw["protocol"] = raw["provider"]
	}
	delete(raw, "provider")

	encoded, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	kv.Value = string(encoded)
	return db.Save(&kv).Error
}
