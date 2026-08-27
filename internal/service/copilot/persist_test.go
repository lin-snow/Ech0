// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	"github.com/lin-snow/ech0/internal/test/helpers"
)

func TestLoadSession_BestEffort(t *testing.T) {
	t.Run("empty userID", func(t *testing.T) {
		s := &CopilotService{durableKV: kvstore.NewMemory()}
		if got := s.loadSession(context.Background(), ""); got != nil {
			t.Fatalf("empty userID should yield nil, got %#v", got)
		}
	})
	t.Run("kv miss", func(t *testing.T) {
		s := &CopilotService{durableKV: kvstore.NewMemory()}
		if got := s.loadSession(context.Background(), "u1"); got != nil {
			t.Fatalf("kv miss should yield nil, got %#v", got)
		}
	})
	t.Run("corrupt json", func(t *testing.T) {
		kv := kvstore.NewMemory()
		if err := kv.Set(context.Background(), chatSessionKey("u1"), "{not json"); err != nil {
			t.Fatalf("seed: %v", err)
		}
		s := &CopilotService{durableKV: kv}
		if got := s.loadSession(context.Background(), "u1"); got != nil {
			t.Fatalf("corrupt json should yield nil, got %#v", got)
		}
	})
}

func TestPersistTurn_SkipsEmpty(t *testing.T) {
	kv := kvstore.NewMemory()
	s := &CopilotService{durableKV: kv}

	s.persistTurn(context.Background(), "u1", "问题", assistantTurn{answer: "   "})

	if got := s.loadSession(context.Background(), "u1"); got != nil {
		t.Fatalf("empty turn should not persist, got %#v", got)
	}
}

func TestPersistTurn_WritesUserAndAssistant(t *testing.T) {
	kv := kvstore.NewMemory()
	s := &CopilotService{durableKV: kv}

	s.persistTurn(context.Background(), "u1", "今年读了什么", assistantTurn{
		answer:      "你读了三体",
		sources:     []embeddingModel.SearchResult{{EchoID: "e1", Content: "三体"}},
		reasoning:   "想了想",
		reasoningMs: 1234,
	})

	msgs := s.loadSession(context.Background(), "u1")
	if len(msgs) != 2 {
		t.Fatalf("want 2 messages, got %d (%#v)", len(msgs), msgs)
	}
	if msgs[0].Role != "user" || msgs[0].Content != "今年读了什么" {
		t.Fatalf("first message should be the user turn, got %+v", msgs[0])
	}
	a := msgs[1]
	if a.Role != "assistant" || a.Content != "你读了三体" {
		t.Fatalf("second message should be the assistant turn, got %+v", a)
	}
	if len(a.Sources) != 1 || a.Sources[0].EchoID != "e1" {
		t.Fatalf("assistant sources not persisted: %+v", a.Sources)
	}
	if a.Reasoning != "想了想" || a.ReasoningMs != 1234 {
		t.Fatalf("reasoning metadata not persisted: %+v", a)
	}
}

func TestAppendTurn_CapsAtMax(t *testing.T) {
	kv := kvstore.NewMemory()
	s := &CopilotService{durableKV: kv}

	turns := make([]ChatMessage, 0, maxStoredChatMessages+5)
	for i := range maxStoredChatMessages + 5 {
		turns = append(turns, ChatMessage{Role: "user", Content: string(rune('A' + i%26))})
	}
	s.appendTurn(context.Background(), "u1", turns...)

	msgs := s.loadSession(context.Background(), "u1")
	if len(msgs) != maxStoredChatMessages {
		t.Fatalf("session should be capped at %d, got %d", maxStoredChatMessages, len(msgs))
	}
	if msgs[len(msgs)-1].Content != turns[len(turns)-1].Content {
		t.Fatalf("cap should keep the most recent turn, got %q", msgs[len(msgs)-1].Content)
	}
}

func TestAppendTurn_EmptyUserSkips(t *testing.T) {
	kv := kvstore.NewMemory()
	s := &CopilotService{durableKV: kv}

	s.appendTurn(context.Background(), "", ChatMessage{Role: "user", Content: "x"})

	if got := s.loadSession(context.Background(), "u1"); got != nil {
		t.Fatalf("empty userID append should not persist anything, got %#v", got)
	}
}

func TestGetSession(t *testing.T) {
	t.Run("none returns empty slice", func(t *testing.T) {
		s := &CopilotService{durableKV: kvstore.NewMemory()}
		got, err := s.GetSession(helpers.CtxAsUser("u1"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Fatalf("want empty non-nil slice, got %#v", got)
		}
	})
	t.Run("returns persisted", func(t *testing.T) {
		kv := kvstore.NewMemory()
		s := &CopilotService{durableKV: kv}
		s.appendTurn(context.Background(), "u1", ChatMessage{Role: "user", Content: "hi"})

		got, err := s.GetSession(helpers.CtxAsUser("u1"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].Content != "hi" {
			t.Fatalf("unexpected session: %#v", got)
		}
	})
}

func TestClearSession(t *testing.T) {
	t.Run("deletes for user", func(t *testing.T) {
		kv := kvstore.NewMemory()
		s := &CopilotService{durableKV: kv}
		s.appendTurn(context.Background(), "u1", ChatMessage{Role: "user", Content: "hi"})

		if err := s.ClearSession(helpers.CtxAsUser("u1")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := s.loadSession(context.Background(), "u1"); got != nil {
			t.Fatalf("session should be cleared, got %#v", got)
		}
	})
	t.Run("anonymous is noop", func(t *testing.T) {
		s := &CopilotService{durableKV: kvstore.NewMemory()}
		if err := s.ClearSession(helpers.CtxAnonymous()); err != nil {
			t.Fatalf("anonymous clear should be a nil no-op, got %v", err)
		}
	})
}
