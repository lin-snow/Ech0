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
	userModel "github.com/lin-snow/ech0/internal/model/user"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"github.com/lin-snow/ech0/pkg/viewer"
)

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

// requireAdmin admits the site owner and nobody else.
//
// Copilot spends the operator's own model budget and reads and writes the
// owner's Echos, so being signed in is not enough. The admin:agent scope on
// the routes only binds access tokens — a session token passes scope checks by
// design — which is why the check lives here, beside every entry point, the
// same way the other settings services do it.
func (s *CopilotService) requireAdmin(ctx context.Context) (userModel.User, error) {
	user, err := s.userReader.GetUserByID(viewer.MustFromContext(ctx).UserID())
	if err != nil {
		return userModel.User{}, err
	}
	if !user.IsAdmin {
		return userModel.User{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return user, nil
}

// AskStream answers one question over SSE. An error it returns means nothing
// has been written yet, so the caller can still answer with a plain HTTP
// error; once the stream is open, failures travel as SSE error events.
func (s *CopilotService) AskStream(ctx context.Context, question string, locale string, timezone string, w http.ResponseWriter) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}
	currentUser, err := s.requireAdmin(ctx)
	if err != nil {
		return err
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

	userID := currentUser.ID
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

	askEvents := make(chan askEvent, 4)
	ask := &asker{
		registry: s.asks,
		events:   askEvents,
		userID:   userID,
		budget:   time.Duration(config.Config().Agent.AskTimeoutSeconds) * time.Second,
		strs:     askStringsFor(locale),
	}

	tools := func(material int) []agent.Tool {
		return []agent.Tool{
			s.searchEchosTool(allTags, agentSetting.Multimodal, locale, loc, agentSetting.ContextWindow, user),
			s.summarizeEchosTool(allTags, agentSetting, material, locale, loc, user),
			s.statsOverviewTool(allTags, locale, loc, user),
			s.askUserTool(ask, locale),
			s.createEchoTool(ask, locale, loc),
			s.updateEchoTool(ask, locale, loc),
			s.deleteEchoTool(ask, locale, loc),
		}
	}
	// The declarations do not depend on the material budget, so a first build
	// prices them and the second carries the budget they leave room for.
	systemPrompt := buildSystemPrompt(locale, today, tagNames, currentUser.Username)
	plan := planContext(agentSetting.ContextWindow,
		agent.EstimateTokens(systemPrompt)+agent.ToolDefTokens(tools(0))+agent.EstimateTokens(question))

	history := historyForModel(s.loadSession(ctx, userID), locale, plan.History, loc)

	stream, err := agent.Run(ctx, agent.RunRequest{
		Setting:          agentSetting,
		Messages:         buildChatMessages(history, question, locale, today, tagNames, currentUser.Username),
		Tools:            tools(plan.Material),
		MaxRounds:        config.Config().Agent.MaxRounds,
		Strings:          runStringsFor(locale),
		Timeout:          time.Duration(config.Config().Agent.TimeoutSeconds) * time.Second,
		MaxContextTokens: plan.Input,
	})
	if err != nil {
		writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return nil
	}

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	finish := func() {
		endReasoning()
		s.persistTurn(ctx, userID, question, assistantTurn{
			answer: assistantBuf.String(), sources: collectedSources,
			reasoning: reasoningBuf.String(), reasoningMs: reasoningMs,
			asks: ask.exchanges(),
		})
		writeSSE(w, flusher, "done", map[string]bool{"done": true})
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case ev, ok := <-stream:
			if !ok {
				finish()
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
				finish()
				return nil
			case agent.AgentError:
				writeSSE(w, flusher, "error", map[string]string{"message": ev.Err.Error()})
				return nil
			}
		case ev := <-askEvents:
			if ev.Open != nil {
				writeSSE(w, flusher, "ask", ev.Open)
			}
			if ev.Closed != "" {
				writeSSE(w, flusher, "ask_closed", map[string]string{"ask_id": ev.Closed})
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data any) {
	payload, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	flusher.Flush()
}
