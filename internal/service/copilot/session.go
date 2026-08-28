// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const maxStoredChatMessages = 50

const maxHistoryTokens = 4000

const toolDefTokenEstimate = 640

const minHistoryTokens = 500

func estimateTokens(s string) int { return utf8.RuneCountInString(s) }

func historyForModel(msgs []ChatMessage, locale string, budgetTokens int, loc *time.Location) []agent.Message {
	if len(msgs) == 0 {
		return nil
	}

	lastSourced := -1
	for i, msg := range slices.Backward(msgs) {
		if msg.Role == "assistant" && len(msg.Sources) > 0 {
			lastSourced = i
			break
		}
	}

	contentOf := func(i int) string {
		c := strings.TrimSpace(msgs[i].Content)
		if i != lastSourced {
			return c
		}
		note := fmt.Sprintf(recentSourcesNoteFor(locale), formatSearchResults(msgs[i].Sources, nil, loc))
		if c == "" {
			return note
		}
		return c + "\n\n" + note
	}

	collected := make([]agent.Message, 0, len(msgs))
	used := 0
	for i, msg := range slices.Backward(msgs) {
		content := contentOf(i)
		if content == "" {
			continue
		}
		if t := estimateTokens(content); used+t > budgetTokens && len(collected) > 0 {
			break
		} else {
			used += t
		}
		collected = append(collected, agent.Message{Role: roleFromString(msg.Role), Content: content})
	}

	for l, r := 0, len(collected)-1; l < r; l, r = l+1, r-1 {
		collected[l], collected[r] = collected[r], collected[l]
	}
	return collected
}

func roleFromString(r string) agent.Role {
	if r == "assistant" {
		return agent.RoleAssistant
	}
	return agent.RoleUser
}

type ChatMessage struct {
	Role        string                        `json:"role"`
	Content     string                        `json:"content"`
	Sources     []embeddingModel.SearchResult `json:"sources,omitempty"`
	Reasoning   string                        `json:"reasoning,omitempty"`
	ReasoningMs int64                         `json:"reasoning_ms,omitempty"`
	Asks        []AskExchange                 `json:"asks,omitempty"`
}

func chatSessionKey(userID string) string {
	return commonModel.ChatSessionKeyPrefix + userID
}

func (s *CopilotService) loadSession(ctx context.Context, userID string) []ChatMessage {
	if userID == "" {
		return nil
	}
	raw, err := s.durableKV.Get(ctx, chatSessionKey(userID))
	if err != nil {
		return nil
	}
	var msgs []ChatMessage
	if err := json.Unmarshal([]byte(raw), &msgs); err != nil {
		return nil
	}
	return msgs
}

func (s *CopilotService) appendTurn(ctx context.Context, userID string, turn ...ChatMessage) {
	if userID == "" {
		return
	}
	msgs := append(s.loadSession(ctx, userID), turn...)
	if len(msgs) > maxStoredChatMessages {
		msgs = msgs[len(msgs)-maxStoredChatMessages:]
	}
	payload, err := json.Marshal(msgs)
	if err != nil {
		logUtil.GetLogger().Warn("failed to marshal chat session",
			slog.String("module", "copilot"), logUtil.Err(err))
		return
	}
	if err := s.durableKV.Set(ctx, chatSessionKey(userID), string(payload)); err != nil {
		logUtil.GetLogger().Warn("failed to persist chat session",
			slog.String("module", "copilot"), logUtil.Err(err))
	}
}

type assistantTurn struct {
	answer      string
	sources     []embeddingModel.SearchResult
	reasoning   string
	reasoningMs int64
	asks        []AskExchange
}

// persistTurn writes the turn down, or drops it when there is nothing in it.
//
// An answered question counts as something in it. A turn that ended right after
// a confirmation — approved and applied, or declined — may have produced no
// prose and no sources, and dropping it would erase the one exchange the person
// actually took part in.
func (s *CopilotService) persistTurn(ctx context.Context, userID, question string, turn assistantTurn) {
	if strings.TrimSpace(turn.answer) == "" && len(turn.sources) == 0 && len(turn.asks) == 0 {
		return
	}
	s.appendTurn(ctx, userID,
		ChatMessage{Role: "user", Content: question},
		ChatMessage{
			Role:        "assistant",
			Content:     turn.answer,
			Sources:     turn.sources,
			Reasoning:   turn.reasoning,
			ReasoningMs: turn.reasoningMs,
			Asks:        turn.asks,
		},
	)
}

func (s *CopilotService) GetSession(ctx context.Context) ([]ChatMessage, error) {
	userID := viewer.MustFromContext(ctx).UserID()
	msgs := s.loadSession(ctx, userID)
	if msgs == nil {
		return []ChatMessage{}, nil
	}
	return msgs, nil
}

func (s *CopilotService) ClearSession(ctx context.Context) error {
	userID := viewer.MustFromContext(ctx).UserID()
	if userID == "" {
		return nil
	}
	return s.durableKV.Delete(ctx, chatSessionKey(userID))
}
