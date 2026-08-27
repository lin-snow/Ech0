// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

func respMarshalParams(t *testing.T, req Request) map[string]any {
	t.Helper()
	p := &openaiResponsesProvider{setting: model.AgentSetting{Model: "gpt-5"}}
	params, err := p.buildParams(req)
	if err != nil {
		t.Fatalf("buildParams failed: %v", err)
	}
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params failed: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("unmarshal params failed: %v", err)
	}
	return body
}

func TestResponsesBuildInput_WireShape(t *testing.T) {
	body := respMarshalParams(t, Request{
		Messages: []Message{
			{Role: RoleSystem, Content: "you are a bot"},
			{Role: RoleUser, Content: "看图", Images: []ImagePart{
				{MediaType: "image/png", Base64: "abc"},
				{URL: "https://example.com/x.jpg"},
			}},
			{Role: RoleAssistant, Content: "calling", ToolCalls: []ToolCall{
				{ID: "call_1", Name: "search_echos", Args: []byte(`{"q":"x"}`)},
				{ID: "call_2", Name: "stats_overview"},
			}},
			{Role: RoleTool, ToolCallID: "call_1", Content: "hit"},
		},
	})

	items, ok := body["input"].([]any)
	if !ok {
		t.Fatalf("input should be an item list, got %T", body["input"])
	}
	if len(items) != 6 {
		t.Fatalf("got %d input items, want 6: %v", len(items), items)
	}

	sys := items[0].(map[string]any)
	if sys["role"] != "system" || sys["content"] != "you are a bot" {
		t.Fatalf("system item = %v", sys)
	}

	user := items[1].(map[string]any)
	parts, ok := user["content"].([]any)
	if user["role"] != "user" || !ok || len(parts) != 3 {
		t.Fatalf("user item should carry 1 text + 2 image parts, got %v", user)
	}
	if text := parts[0].(map[string]any); text["type"] != "input_text" || text["text"] != "看图" {
		t.Fatalf("first part = %v, want input_text", text)
	}
	if img := parts[1].(map[string]any); img["type"] != "input_image" ||
		img["image_url"] != "data:image/png;base64,abc" {
		t.Fatalf("base64 image part = %v, want flat data-URL image_url", img)
	}
	if img := parts[2].(map[string]any); img["image_url"] != "https://example.com/x.jpg" {
		t.Fatalf("url image part = %v", img)
	}

	if assistant := items[2].(map[string]any); assistant["role"] != "assistant" || assistant["content"] != "calling" {
		t.Fatalf("assistant text item = %v", assistant)
	}
	call := items[3].(map[string]any)
	if call["type"] != "function_call" || call["call_id"] != "call_1" ||
		call["name"] != "search_echos" || call["arguments"] != `{"q":"x"}` {
		t.Fatalf("function_call item = %v", call)
	}
	if noArgs := items[4].(map[string]any); noArgs["arguments"] != "{}" {
		t.Fatalf("empty args should fall back to {}, got %v", noArgs)
	}
	out := items[5].(map[string]any)
	if out["type"] != "function_call_output" || out["call_id"] != "call_1" || out["output"] != "hit" {
		t.Fatalf("function_call_output item = %v", out)
	}
}

func TestResponsesBuildParams_ToolsAndOptions(t *testing.T) {
	temp := float32(0.4)
	body := respMarshalParams(t, Request{
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Tools:       []ToolDef{{Name: "search_echos", Description: "检索", Parameters: []byte(`{"type":"object"}`)}},
		Temperature: &temp,
		MaxTokens:   128,
	})

	tools, ok := body["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("tools = %v, want 1 entry", body["tools"])
	}
	tool := tools[0].(map[string]any)
	if tool["type"] != "function" || tool["name"] != "search_echos" || tool["description"] != "检索" {
		t.Fatalf("tool should be flat with name/description, got %v", tool)
	}
	if _, nested := tool["function"]; nested {
		t.Fatalf("Responses tools must not nest under \"function\": %v", tool)
	}
	if tool["strict"] != false {
		t.Fatalf("strict should be explicitly false, got %v", tool["strict"])
	}
	if params, ok := tool["parameters"].(map[string]any); !ok || params["type"] != "object" {
		t.Fatalf("parameters should carry the JSON Schema, got %v", tool["parameters"])
	}

	if body["max_output_tokens"] != float64(128) {
		t.Fatalf("max_output_tokens = %v, want 128", body["max_output_tokens"])
	}
	if body["temperature"] != float64(float32(0.4)) {
		t.Fatalf("temperature = %v", body["temperature"])
	}
	if body["store"] != false {
		t.Fatalf("store should be explicitly false, got %v", body["store"])
	}
	if _, has := body["reasoning"]; has {
		t.Fatalf("reasoning must not be sent (unsupported_parameter on non-reasoning models)")
	}
	if _, has := body["max_tokens"]; has {
		t.Fatalf("max_tokens is the Chat Completions field name; Responses uses max_output_tokens")
	}
}

func TestResponsesBuildTools_InvalidSchemaFails(t *testing.T) {
	p := &openaiResponsesProvider{}
	if _, err := p.buildTools([]ToolDef{{Name: "broken", Parameters: []byte(`{`)}}); err == nil {
		t.Fatalf("an unparsable JSON Schema should fail loudly")
	}
}

type respServer struct {
	*httptest.Server
	path string
	body map[string]any
}

func newRespServer(t *testing.T, handler func(w http.ResponseWriter)) *respServer {
	t.Helper()
	s := &respServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.path = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &s.body)
		handler(w)
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *respServer) provider() *openaiResponsesProvider {
	return &openaiResponsesProvider{setting: model.AgentSetting{
		Model:   "gpt-5",
		ApiKey:  "sk-test",
		BaseURL: s.URL + "/v1",
	}}
}

func writeSSE(w http.ResponseWriter, frames ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	for _, f := range frames {
		var typed struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal([]byte(f), &typed)
		_, _ = io.WriteString(w, "event: "+typed.Type+"\ndata: "+f+"\n\n")
		if flusher != nil {
			flusher.Flush()
		}
	}
}

func drainProviderEvents(t *testing.T, ch <-chan Event) []Event {
	t.Helper()
	var got []Event
	timeout := time.After(5 * time.Second)
	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return got
			}
			got = append(got, ev)
		case <-timeout:
			t.Fatalf("timed out draining events, got so far: %+v", got)
		}
	}
}

func TestResponsesStream_TextReasoningAndToolCall(t *testing.T) {
	srv := newRespServer(t, func(w http.ResponseWriter) {
		writeSSE(w,
			`{"type":"response.created","sequence_number":1,"response":{"id":"resp_1","status":"in_progress"}}`,
			`{"type":"response.reasoning_summary_text.delta","sequence_number":2,"delta":"先检索"}`,
			`{"type":"response.output_text.delta","sequence_number":3,"delta":"你最近"}`,
			`{"type":"response.output_text.delta","sequence_number":4,"delta":"在读书"}`,
			`{"type":"response.output_item.done","sequence_number":5,"output_index":0,"item":`+
				`{"id":"fc_1","type":"function_call","call_id":"call_1","name":"search_echos",`+
				`"arguments":"{\"q\":\"书\"}","status":"completed"}}`,
			`{"type":"response.completed","sequence_number":6,"response":{"id":"resp_1","status":"completed",`+
				`"output":[{"id":"fc_1","type":"function_call","call_id":"call_1","name":"search_echos",`+
				`"arguments":"{\"q\":\"书\"}","status":"completed"}]}}`,
		)
	})

	ch, err := srv.provider().Stream(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Content: "我最近在读什么"}},
		Tools:    []ToolDef{{Name: "search_echos", Parameters: []byte(`{"type":"object"}`)}},
	})
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}
	got := drainProviderEvents(t, ch)

	if srv.path != "/v1/responses" {
		t.Fatalf("request path = %q, want /v1/responses", srv.path)
	}
	if srv.body["stream"] != true {
		t.Fatalf("streaming request must set stream=true, got %v", srv.body["stream"])
	}

	var text, reasoning strings.Builder
	var calls []ToolCall
	var done bool
	for _, ev := range got {
		switch ev.Kind {
		case EventTextDelta:
			text.WriteString(ev.Text)
		case EventReasoningDelta:
			reasoning.WriteString(ev.Text)
		case EventToolCall:
			calls = append(calls, ev.ToolCall)
		case EventDone:
			done = true
		case EventError:
			t.Fatalf("unexpected error event: %v", ev.Err)
		}
	}
	if text.String() != "你最近在读书" {
		t.Fatalf("text = %q, want 你最近在读书", text.String())
	}
	if reasoning.String() != "先检索" {
		t.Fatalf("reasoning = %q, want 先检索", reasoning.String())
	}
	if len(calls) != 1 {
		t.Fatalf("got %d tool calls, want 1 (completed event must not duplicate item events): %+v", len(calls), calls)
	}
	if calls[0].ID != "call_1" || calls[0].Name != "search_echos" || string(calls[0].Args) != `{"q":"书"}` {
		t.Fatalf("tool call = %+v", calls[0])
	}
	if !done {
		t.Fatalf("stream should end with EventDone: %+v", got)
	}
}

func TestResponsesStream_ToolCallFromFinalResponseFallback(t *testing.T) {
	srv := newRespServer(t, func(w http.ResponseWriter) {
		writeSSE(w,
			`{"type":"response.completed","sequence_number":1,"response":{"id":"resp_1","status":"completed",`+
				`"output":[{"id":"fc_1","type":"function_call","call_id":"call_9","name":"search_echos",`+
				`"arguments":"","status":"completed"}]}}`,
		)
	})

	ch, _ := srv.provider().Stream(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "q"}}})
	var calls []ToolCall
	for _, ev := range drainProviderEvents(t, ch) {
		if ev.Kind == EventToolCall {
			calls = append(calls, ev.ToolCall)
		}
	}
	if len(calls) != 1 || calls[0].ID != "call_9" || string(calls[0].Args) != "{}" {
		t.Fatalf("fallback tool call = %+v", calls)
	}
}

func TestResponsesStream_ErrorEventAborts(t *testing.T) {
	srv := newRespServer(t, func(w http.ResponseWriter) {
		writeSSE(w,
			`{"type":"response.output_text.delta","sequence_number":1,"delta":"半句"}`,
			`{"type":"error","sequence_number":2,"code":"rate_limit_exceeded","message":"too many requests","param":null}`,
		)
	})

	ch, _ := srv.provider().Stream(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "q"}}})
	got := drainProviderEvents(t, ch)
	last := got[len(got)-1]
	if last.Kind != EventError {
		t.Fatalf("stream should end with EventError, got %+v", got)
	}
	if !strings.Contains(last.Err.Error(), "too many requests") ||
		!strings.Contains(last.Err.Error(), "rate_limit_exceeded") {
		t.Fatalf("error should carry message and code, got %v", last.Err)
	}
}

func TestResponsesStream_FailedResponseAborts(t *testing.T) {
	srv := newRespServer(t, func(w http.ResponseWriter) {
		writeSSE(w,
			`{"type":"response.failed","sequence_number":1,"response":{"id":"resp_1","status":"failed",`+
				`"error":{"code":"server_error","message":"boom"}}}`,
		)
	})

	ch, _ := srv.provider().Stream(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "q"}}})
	got := drainProviderEvents(t, ch)
	last := got[len(got)-1]
	if last.Kind != EventError || !strings.Contains(last.Err.Error(), "boom") {
		t.Fatalf("failed response should surface as EventError with its message, got %+v", got)
	}
}

func TestResponsesComplete_AggregatesOutputText(t *testing.T) {
	srv := newRespServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"resp_1","object":"response","created_at":1,"model":"gpt-5",`+
			`"status":"completed","output":[{"id":"msg_1","type":"message","role":"assistant","status":"completed",`+
			`"content":[{"type":"output_text","text":"最近在读《三体》","annotations":[]}]}]}`)
	})

	resp, err := srv.provider().Complete(context.Background(), Request{
		Messages:  []Message{{Role: RoleUser, Content: "ping"}},
		MaxTokens: 16,
	})
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}
	if resp.Text != "最近在读《三体》" {
		t.Fatalf("text = %q", resp.Text)
	}
	if srv.body["stream"] != nil {
		t.Fatalf("non-streaming request must not set stream, got %v", srv.body["stream"])
	}
}

func TestResponsesComplete_FailedStatusIsError(t *testing.T) {
	srv := newRespServer(t, func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"id":"resp_1","object":"response","created_at":1,"model":"gpt-5",`+
			`"status":"failed","error":{"code":"server_error","message":"boom"},"output":[]}`)
	})

	_, err := srv.provider().Complete(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "ping"}}})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("failed status should error with its message, got %v", err)
	}
}

func TestRespStreamErrorMessage(t *testing.T) {
	cases := []struct{ code, message, want string }{
		{"c", "m", "m（c）"},
		{"", "m", "m"},
		{"c", "", "c"},
		{"", "", "生成失败且未返回错误详情"},
	}
	for _, tc := range cases {
		if got := respStreamErrorMessage(tc.code, tc.message); got != tc.want {
			t.Fatalf("respStreamErrorMessage(%q,%q) = %q, want %q", tc.code, tc.message, got, tc.want)
		}
	}
}

func TestResponsesRun_ToolResultFedBackToModel(t *testing.T) {
	var mu sync.Mutex
	var bodies []map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		mu.Lock()
		bodies = append(bodies, body)
		round := len(bodies)
		mu.Unlock()

		if round == 1 {
			writeSSE(w,
				`{"type":"response.output_item.done","sequence_number":1,"output_index":0,"item":`+
					`{"id":"fc_1","type":"function_call","call_id":"call_1","name":"search_echos",`+
					`"arguments":"{\"query\":\"读书\"}","status":"completed"}}`,
				`{"type":"response.completed","sequence_number":2,"response":{"id":"resp_1","status":"completed","output":[]}}`,
			)
			return
		}
		writeSSE(w,
			`{"type":"response.output_text.delta","sequence_number":1,"delta":"你在读《三体》"}`,
			`{"type":"response.completed","sequence_number":2,"response":{"id":"resp_2","status":"completed","output":[]}}`,
		)
	}))
	t.Cleanup(srv.Close)

	tool, toolCalls := countingTool("search_echos", ToolOutput{Content: "《三体》"}, nil)
	out, err := Run(context.Background(), RunRequest{
		Setting: model.AgentSetting{
			Enable:   true,
			Protocol: string(commonModel.OpenAIResponses),
			Model:    "gpt-5",
			ApiKey:   "sk-test",
			BaseURL:  srv.URL + "/v1",
		},
		Messages: []Message{{Role: RoleUser, Content: "我最近在读什么"}},
		Tools:    []Tool{tool},
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	var text strings.Builder
	var searching, results, done int
	for _, ev := range drain(out) {
		switch ev.Kind {
		case AgentDelta:
			text.WriteString(ev.Text)
		case AgentSearching:
			searching++
		case AgentToolResult:
			results++
		case AgentDone:
			done++
		case AgentError:
			t.Fatalf("unexpected error event: %v", ev.Err)
		}
	}
	if *toolCalls != 1 || searching != 1 || results != 1 || done != 1 {
		t.Fatalf("tool=%d searching=%d results=%d done=%d, want 1/1/1/1", *toolCalls, searching, results, done)
	}
	if text.String() != "你在读《三体》" {
		t.Fatalf("answer = %q", text.String())
	}

	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 {
		t.Fatalf("got %d requests, want 2 (tool round + answer round)", len(bodies))
	}
	items := bodies[1]["input"].([]any)
	if len(items) != 3 {
		t.Fatalf("second round input = %v, want user + function_call + function_call_output", items)
	}
	call := items[1].(map[string]any)
	result := items[2].(map[string]any)
	if call["type"] != "function_call" || call["call_id"] != "call_1" {
		t.Fatalf("second round should replay the function_call, got %v", call)
	}
	if result["type"] != "function_call_output" || result["call_id"] != "call_1" ||
		!strings.Contains(result["output"].(string), "《三体》") {
		t.Fatalf("tool result must be fed back keyed by call_id, got %v", result)
	}
}
