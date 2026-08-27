// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"strings"
	"testing"
)

func TestBuildContextBlock_IdentityLine(t *testing.T) {
	zh := buildContextBlock("zh-CN", "2026-06-02", nil, "Alice")
	if !strings.Contains(zh, "Alice") {
		t.Fatalf("zh context block should mention the display name: %q", zh)
	}
	if !strings.Contains(zh, "本人发布") {
		t.Fatalf("zh context block should scope retrieval to the user: %q", zh)
	}
	if strings.Contains(zh, "回顾") {
		t.Fatalf("identity line must be task-neutral (no 回顾): %q", zh)
	}

	en := buildContextBlock("en", "2026-06-02", nil, "Alice")
	if !strings.Contains(en, "Alice") || !strings.Contains(en, "posted themselves") {
		t.Fatalf("en context block should mention name + scope: %q", en)
	}

	if got := buildContextBlock("zh-CN", "2026-06-02", nil, ""); strings.Contains(got, "当前与你对话的是") {
		t.Fatalf("empty display name should omit identity line: %q", got)
	}
}

func TestChatSystemPrompt_GeneralFraming(t *testing.T) {
	zh := chatSystemPromptFor("zh-CN")
	if strings.Contains(zh, "回顾助手") {
		t.Fatalf("system prompt should not be framed as a review-only assistant: %q", zh)
	}
	if !strings.Contains(zh, "私人助手") {
		t.Fatalf("system prompt should be framed as a general personal assistant: %q", zh)
	}
	for _, tool := range []string{"search_echos", "summarize_echos", "stats_overview"} {
		if !strings.Contains(zh, tool) {
			t.Fatalf("system prompt should still declare tool %q: %q", tool, zh)
		}
	}

	en := chatSystemPromptFor("en")
	if !strings.Contains(en, "personal assistant") {
		t.Fatalf("en system prompt should be framed as a personal assistant: %q", en)
	}
}
