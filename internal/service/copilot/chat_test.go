// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

func TestAgentSetting_MissingKeyYieldsDisabledDefault(t *testing.T) {
	s := &CopilotService{durableKV: kvstore.NewMemory()}
	got, err := s.agentSetting(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Enable || got.Protocol != string(commonModel.OpenAI) {
		t.Fatalf("want disabled default, got %+v", got)
	}
}

func TestAgentSetting_HappyPath(t *testing.T) {
	want := settingModel.AgentSetting{Enable: true, Protocol: "openai", Model: "gpt-x", ApiKey: "k"}
	raw, _ := json.Marshal(want)
	kv := kvstore.NewMemory()
	if err := kv.Set(context.Background(), commonModel.AgentSettingKey, string(raw)); err != nil {
		t.Fatalf("seed kv: %v", err)
	}
	s := &CopilotService{durableKV: kv}

	got, err := s.agentSetting(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Protocol != want.Protocol || got.Model != want.Model || !got.Enable {
		t.Fatalf("decoded setting mismatch: got %+v want %+v", got, want)
	}
}
