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

type agentProtocolCollapseMigrator struct{}

func NewAgentProtocolCollapseMigrator() Migrator {
	return &agentProtocolCollapseMigrator{}
}

func (m *agentProtocolCollapseMigrator) Name() string {
	return "agent_provider_collapse_migrator"
}

func (m *agentProtocolCollapseMigrator) Key() string {
	return commonModel.AgentProtocolCollapsedKey
}

func (m *agentProtocolCollapseMigrator) CanRerun() bool {
	return false
}

func (m *agentProtocolCollapseMigrator) Migrate(db *gorm.DB) error {
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

	provider, _ := raw["provider"].(string)
	mapped := collapseAgentProtocol(provider)
	if mapped == provider {
		return nil
	}

	raw["provider"] = mapped
	encoded, err := json.Marshal(raw)
	if err != nil {
		return err
	}

	kv.Value = string(encoded)
	return db.Save(&kv).Error
}

func collapseAgentProtocol(old string) string {
	switch old {
	case string(commonModel.Anthropic), "gemini", string(commonModel.OpenAI):
		return old
	default:
		return string(commonModel.OpenAI)
	}
}
