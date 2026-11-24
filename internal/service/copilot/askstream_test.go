// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/kvstore"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
)

type noFlushWriter struct{ h http.Header }

func (w *noFlushWriter) Header() http.Header         { return w.h }
func (w *noFlushWriter) Write(b []byte) (int, error) { return len(b), nil }
func (w *noFlushWriter) WriteHeader(int)             {}

func seedAgentSetting(t *testing.T, kv kvstore.Store, setting settingModel.AgentSetting) {
	t.Helper()
	raw, err := json.Marshal(setting)
	if err != nil {
		t.Fatalf("marshal setting: %v", err)
	}
	if err := kv.Set(context.Background(), commonModel.AgentSettingKey, string(raw)); err != nil {
		t.Fatalf("seed agent setting: %v", err)
	}
}

func TestAskStream_StreamingUnsupported(t *testing.T) {
	s := &CopilotService{durableKV: kvstore.NewMemory()}
	w := &noFlushWriter{h: http.Header{}}

	err := s.AskStream(helpers.CtxAsUser("u1"), "hi", "zh-CN", "", w)
	if err == nil || !strings.Contains(err.Error(), "streaming unsupported") {
		t.Fatalf("want streaming-unsupported error, got %v", err)
	}
}

func TestAskStream_EmptyQuestion(t *testing.T) {
	s := &CopilotService{durableKV: kvstore.NewMemory(), userReader: adminReader()}
	rec := httptest.NewRecorder()

	if err := s.AskStream(helpers.CtxAsUser("u1"), "   ", "zh-CN", "", rec); err != nil {
		t.Fatalf("AskStream should return nil and report via SSE, got %v", err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "event: error") || !strings.Contains(body, "empty question") {
		t.Fatalf("expected SSE empty-question error, got %q", body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("expected SSE content-type, got %q", ct)
	}
}

func TestAskStream_UserLookupError(t *testing.T) {
	s := &CopilotService{
		durableKV:  kvstore.NewMemory(),
		userReader: &stubUserReader{err: errPropagate},
	}
	rec := httptest.NewRecorder()

	err := s.AskStream(helpers.CtxAsUser("u1"), "你好", "zh-CN", "", rec)
	if err == nil || !strings.Contains(err.Error(), "user gone") {
		t.Fatalf("a failed lookup must be returned before the stream opens, got %v", err)
	}
	if rec.Body.Len() != 0 || rec.Header().Get("Content-Type") == "text/event-stream" {
		t.Fatalf("nothing may be written before the caller can answer with an HTTP error, got %q", rec.Body.String())
	}
}

// A signed-in user who is not the site owner must not reach the owner's model,
// Echos or conversation — whatever token type they hold.
func TestCopilot_NonAdminIsRefused(t *testing.T) {
	s := &CopilotService{
		durableKV:  kvstore.NewMemory(),
		userReader: &stubUserReader{user: userModel.User{ID: "u2", Username: "bob"}},
		asks:       newAskRegistry(),
	}
	ctx := helpers.CtxAsUser("u2")
	rec := httptest.NewRecorder()

	if err := s.AskStream(ctx, "你好", "zh-CN", "", rec); !errors.Is(err, authz.ErrDenied) {
		t.Fatalf("AskStream: want ErrDenied, got %v", err)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("AskStream wrote %q for a refused caller", rec.Body.String())
	}
	if _, err := s.GetSession(ctx); err == nil {
		t.Fatalf("GetSession: want refusal")
	}
	if err := s.ClearSession(ctx); err == nil {
		t.Fatalf("ClearSession: want refusal")
	}
	if err := s.AnswerAsk(ctx, "ask-1", []AskAnswer{{QuestionID: "q"}}); !errors.Is(err, authz.ErrDenied) {
		t.Fatalf("AnswerAsk: want refusal, got %v", err)
	}
}

func TestAskStream_AgentSettingMissingServesDisabledDefault(t *testing.T) {
	s := &CopilotService{
		durableKV:   kvstore.NewMemory(),
		userReader:  &stubUserReader{user: userModel.User{ID: "u1", Username: "alice", Role: userModel.RoleAdmin}},
		echoService: &stubEchoSvc{tags: nil},
	}
	rec := httptest.NewRecorder()

	if err := s.AskStream(helpers.CtxAsUser("u1"), "你好", "zh-CN", "", rec); err != nil {
		t.Fatalf("AskStream should return nil, got %v", err)
	}
	if body := rec.Body.String(); !strings.Contains(body, "event: error") || !strings.Contains(body, commonModel.AGENT_NOT_ENABLED) {
		t.Fatalf("expected SSE agent-not-enabled error, got %q", body)
	}
}

func TestAskStream_AgentRunValidationError(t *testing.T) {
	kv := kvstore.NewMemory()
	seedAgentSetting(t, kv, settingModel.AgentSetting{Enable: false, Protocol: "openai", Model: "gpt"})
	s := &CopilotService{
		durableKV:   kv,
		userReader:  &stubUserReader{user: userModel.User{ID: "u1", Username: "alice", Role: userModel.RoleAdmin}},
		echoService: &stubEchoSvc{tags: nil},
	}
	rec := httptest.NewRecorder()

	if err := s.AskStream(helpers.CtxAsUser("u1"), "你好", "zh-CN", "Asia/Shanghai", rec); err != nil {
		t.Fatalf("AskStream should return nil, got %v", err)
	}
	if body := rec.Body.String(); !strings.Contains(body, "event: error") || !strings.Contains(body, commonModel.AGENT_NOT_ENABLED) {
		t.Fatalf("expected SSE agent-not-enabled error, got %q", body)
	}
}

var errPropagate = userLookupErr("user gone")

type userLookupErr string

func (e userLookupErr) Error() string { return string(e) }
