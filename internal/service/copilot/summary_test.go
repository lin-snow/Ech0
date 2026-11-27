// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/kvstore"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

func seedRecent(t *testing.T, kv kvstore.Store, fingerprint, text string) {
	t.Helper()
	raw, _ := json.Marshal(recentCache{Fingerprint: fingerprint, Text: text})
	if err := kv.Set(context.Background(), agent.GEN_RECENT, string(raw)); err != nil {
		t.Fatal(err)
	}
}

func TestGetRecent_ServesMatchingCache(t *testing.T) {
	kv := kvstore.NewMemory()
	setting := settingModel.AgentSetting{Enable: true, Protocol: "openai", Model: "gpt-4o", ApiKey: "k"}
	seedAgentSetting(t, kv, setting)
	seedRecent(t, kv, recentFingerprint(setting), "cached summary")

	got, err := (&CopilotService{durableKV: kv}).GetRecent(context.Background())
	if err != nil || got != "cached summary" {
		t.Fatalf("got %q, %v; want the cached summary", got, err)
	}
}

// The endpoint is public: switching the agent off has to stop the summary
// being served, not leave it up until the next post clears the cache.
func TestGetRecent_DisabledIsRefusedEvenWhenCached(t *testing.T) {
	kv := kvstore.NewMemory()
	setting := settingModel.AgentSetting{Enable: false, Protocol: "openai", Model: "gpt-4o"}
	seedAgentSetting(t, kv, setting)
	seedRecent(t, kv, recentFingerprint(setting), "stale summary")

	if _, err := (&CopilotService{durableKV: kv}).GetRecent(context.Background()); err == nil ||
		err.Error() != apperr.ErrAgentDisabled.Error() {
		t.Fatalf("want %q, got %v", apperr.ErrAgentDisabled.Error(), err)
	}
}

func TestRecentFingerprint_TracksWhatShapesTheSummary(t *testing.T) {
	base := settingModel.AgentSetting{Protocol: "openai", Model: "a", Prompt: "p", BaseURL: "u", ApiKey: "k1"}
	for name, mutate := range map[string]func(*settingModel.AgentSetting){
		"model":    func(s *settingModel.AgentSetting) { s.Model = "b" },
		"prompt":   func(s *settingModel.AgentSetting) { s.Prompt = "q" },
		"endpoint": func(s *settingModel.AgentSetting) { s.BaseURL = "v" },
	} {
		changed := base
		mutate(&changed)
		if recentFingerprint(changed) == recentFingerprint(base) {
			t.Fatalf("changing the %s must invalidate the cached summary", name)
		}
	}
	rotated := base
	rotated.ApiKey = "k2"
	if recentFingerprint(rotated) != recentFingerprint(base) {
		t.Fatalf("rotating the API key does not change what the summary says")
	}
}
