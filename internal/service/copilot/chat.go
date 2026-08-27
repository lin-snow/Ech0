// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const chatTemperature float32 = 0.4

func (s *CopilotService) agentSetting(ctx context.Context) (settingModel.AgentSetting, error) {
	var setting settingModel.AgentSetting
	raw, err := s.durableKV.Get(ctx, commonModel.AgentSettingKey)
	if err != nil {
		return setting, errors.New(commonModel.AGENT_SETTING_NOT_FOUND)
	}
	if err := json.Unmarshal([]byte(raw), &setting); err != nil {
		return setting, err
	}
	return setting, nil
}

func (s *CopilotService) AskStream(ctx context.Context, question string, locale string, timezone string, w http.ResponseWriter) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	question = strings.TrimSpace(question)
	if question == "" {
		writeSSE(w, flusher, "error", map[string]string{"message": "empty question"})
		return nil
	}

	userID := viewer.MustFromContext(ctx).UserID()
	currentUser, err := s.userReader.GetUserByID(userID)
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}
	user := chatUser{ID: currentUser.ID, Username: currentUser.Username}
	var assistantBuf strings.Builder
	var collectedSources []embeddingModel.SearchResult
	var reasoningBuf strings.Builder
	var reasoningStart time.Time
	var reasoningMs int64
	reasoningEnded := false
	endReasoning := func() {
		if reasoningStart.IsZero() || reasoningEnded {
			return
		}
		reasoningEnded = true
		reasoningMs = time.Since(reasoningStart).Milliseconds()
		writeSSE(w, flusher, "reasoning_done", map[string]int64{"duration_ms": reasoningMs})
	}

	agentSetting, err := s.agentSetting(ctx)
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	allTags, _ := s.echoService.GetAllTags()
	loc := timezoneUtil.LoadLocationOrUTC(timezone)
	today := time.Now().UTC().In(loc).Format("2006-01-02")
	tagNames := tagNamesForPrompt(allTags)

	historyBudget := max(maxHistoryTokens-estimateTokens(buildSystemPrompt(locale, today, tagNames, currentUser.Username))-toolDefTokenEstimate, minHistoryTokens)

	history := historyForModel(s.loadSession(ctx, userID), locale, historyBudget, loc)

	temp := chatTemperature
	stream, err := agent.Run(ctx, agent.RunRequest{
		Setting:  agentSetting,
		Messages: buildChatMessages(history, question, locale, today, tagNames, currentUser.Username),
		Tools: []agent.Tool{
			s.searchEchosTool(allTags, agentSetting.Multimodal, locale, loc, agentSetting.ContextWindow, user),
			s.summarizeEchosTool(allTags, agentSetting, locale, loc, user),
			s.statsOverviewTool(allTags, locale, loc, user),
		},
		MaxRounds:        config.Config().Agent.MaxRounds,
		Temp:             &temp,
		Strings:          runStringsFor(locale),
		Timeout:          time.Duration(config.Config().Agent.TimeoutSeconds) * time.Second,
		MaxContextTokens: chatContextBudgetTokens(agentSetting),
	})
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case ev, ok := <-stream:
			if !ok {
				endReasoning()
				s.persistTurn(ctx, userID, question, assistantTurn{
					answer: assistantBuf.String(), sources: collectedSources,
					reasoning: reasoningBuf.String(), reasoningMs: reasoningMs,
				})
				writeSSE(w, flusher, "done", map[string]bool{"done": true})
				return nil
			}
			switch ev.Kind {
			case agent.AgentDelta:
				if ev.Text != "" {
					endReasoning()
					assistantBuf.WriteString(ev.Text)
					writeSSE(w, flusher, "delta", map[string]string{"text": ev.Text})
				}
			case agent.AgentReasoning:
				if ev.Text != "" {
					if reasoningStart.IsZero() {
						reasoningStart = time.Now()
					}
					reasoningBuf.WriteString(ev.Text)
					writeSSE(w, flusher, "reasoning", map[string]string{"text": ev.Text})
				}
			case agent.AgentSearching:
				writeSSE(w, flusher, "searching", map[string]string{
					"name":  ev.ToolName,
					"query": searchHintOf(ev.ToolArgs),
				})
			case agent.AgentToolResult:
				switch meta := ev.Meta.(type) {
				case []embeddingModel.SearchResult:
					collectedSources = append(collectedSources, meta...)
					writeSSE(w, flusher, "sources", meta)
				case aggregateCoverage:
					writeSSE(w, flusher, "coverage", meta)
				}
			case agent.AgentDone:
				endReasoning()
				s.persistTurn(ctx, userID, question, assistantTurn{
					answer: assistantBuf.String(), sources: collectedSources,
					reasoning: reasoningBuf.String(), reasoningMs: reasoningMs,
				})
				writeSSE(w, flusher, "done", map[string]bool{"done": true})
				return nil
			case agent.AgentError:
				writeSSE(w, flusher, "error", map[string]string{"message": ev.Err.Error()})
				return nil
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	payload, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	flusher.Flush()
}
