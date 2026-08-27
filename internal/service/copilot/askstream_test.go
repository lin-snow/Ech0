// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	s := &CopilotService{durableKV: kvstore.NewMemory()}
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

	if err := s.AskStream(helpers.CtxAsUser("u1"), "你好", "zh-CN", "", rec); err != nil {
		t.Fatalf("AskStream should return nil, got %v", err)
	}
	if body := rec.Body.String(); !strings.Contains(body, "event: error") || !strings.Contains(body, "user gone") {
		t.Fatalf("expected SSE user-lookup error, got %q", body)
	}
}

func TestAskStream_AgentSettingMissing(t *testing.T) {
	s := &CopilotService{
		durableKV:  kvstore.NewMemory(),
		userReader: &stubUserReader{user: userModel.User{ID: "u1", Username: "alice"}},
	}
	rec := httptest.NewRecorder()

	if err := s.AskStream(helpers.CtxAsUser("u1"), "你好", "zh-CN", "", rec); err != nil {
		t.Fatalf("AskStream should return nil, got %v", err)
	}
	if body := rec.Body.String(); !strings.Contains(body, "event: error") || !strings.Contains(body, commonModel.AGENT_SETTING_NOT_FOUND) {
		t.Fatalf("expected SSE agent-setting-not-found error, got %q", body)
	}
}

func TestAskStream_AgentRunValidationError(t *testing.T) {
	kv := kvstore.NewMemory()
	seedAgentSetting(t, kv, settingModel.AgentSetting{Enable: false, Protocol: "openai", Model: "gpt"})
	s := &CopilotService{
		durableKV:   kv,
		userReader:  &stubUserReader{user: userModel.User{ID: "u1", Username: "alice"}},
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
