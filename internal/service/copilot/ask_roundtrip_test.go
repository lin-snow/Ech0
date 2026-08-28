// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/kvstore"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/test/helpers"
)

// This file drives the whole confirmation round-trip the way the browser does:
// a real model endpoint (an httptest server speaking the OpenAI wire format)
// asks to delete an Echo, AskStream parks the run and writes a real `ask` SSE
// frame, the frame is read off the wire, and the decision goes back through the
// real AnswerAsk — the same call the HTTP handler makes.
//
// Nothing here is a stub of the mechanism under test. The only fake is the model
// itself, because the whole point is what happens between its tool call and the
// database.

// sseWriter is an http.ResponseWriter that publishes each flushed frame, so a
// test can answer a question while the run is still parked on it. A
// ResponseRecorder cannot: its buffer is only readable once the handler returns,
// and this handler does not return until the question is answered.
type sseWriter struct {
	mu     sync.Mutex
	header http.Header
	buf    strings.Builder
	frames chan string
	cursor int
}

func newSSEWriter() *sseWriter {
	return &sseWriter{header: http.Header{}, frames: make(chan string, 32)}
}

func (w *sseWriter) Header() http.Header { return w.header }
func (w *sseWriter) WriteHeader(int)     {}

func (w *sseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf.Write(b)
	return len(b), nil
}

// Flush publishes whatever complete frames have accumulated. Frames are split on
// the blank line that terminates an SSE event, exactly as the browser splits
// them.
func (w *sseWriter) Flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	all := w.buf.String()
	for {
		idx := strings.Index(all[w.cursor:], "\n\n")
		if idx < 0 {
			return
		}
		frame := all[w.cursor : w.cursor+idx]
		w.cursor += idx + 2
		select {
		case w.frames <- frame:
		default:
		}
	}
}

func (w *sseWriter) body() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// awaitAsk reads frames until the `ask` event arrives and returns its payload.
func awaitAsk(t *testing.T, w *sseWriter, within time.Duration) Ask {
	t.Helper()
	deadline := time.After(within)
	for {
		select {
		case frame := <-w.frames:
			name, data := parseSSEFrame(frame)
			if name != "ask" {
				continue
			}
			var ask Ask
			if err := json.Unmarshal([]byte(data), &ask); err != nil {
				t.Fatalf("ask payload is not valid JSON: %v (%q)", err, data)
			}
			return ask
		case <-deadline:
			t.Fatalf("no ask frame within %s; stream so far:\n%s", within, w.body())
		}
	}
}

func parseSSEFrame(frame string) (name, data string) {
	for _, line := range strings.Split(frame, "\n") {
		switch {
		case strings.HasPrefix(line, "event:"):
			name = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}
	return name, data
}

// modelServer speaks just enough of the OpenAI streaming wire format to emit one
// tool call and then one sentence. Each request consumes the next script entry,
// so the run's two rounds are scripted independently.
func modelServer(t *testing.T, scripts ...[]string) *httptest.Server {
	t.Helper()
	var (
		mu    sync.Mutex
		round int
	)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		current := round
		round++
		mu.Unlock()

		// Drain the body so a client-side write error cannot be mistaken for a
		// protocol problem.
		_, _ = bufio.NewReader(r.Body).ReadString(0)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Errorf("httptest writer does not flush")
			return
		}

		chunks := []string{`{"choices":[{"delta":{"content":"（脚本用尽）"},"index":0}]}`}
		if current < len(scripts) {
			chunks = scripts[current]
		}
		for _, chunk := range chunks {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", chunk)
			flusher.Flush()
		}
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
}

func toolCallChunk(id, name, args string) string {
	payload, _ := json.Marshal(args)
	return fmt.Sprintf(
		`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":%q,"type":"function","function":{"name":%q,"arguments":%s}}]},"index":0}]}`,
		id, name, payload)
}

func textChunk(text string) string {
	payload, _ := json.Marshal(text)
	return fmt.Sprintf(`{"choices":[{"delta":{"content":%s},"index":0}]}`, payload)
}

func newRoundTripService(t *testing.T, echoSvc EchoService, baseURL string) *CopilotService {
	t.Helper()
	kv := kvstore.NewMemory()
	seedAgentSetting(t, kv, settingModel.AgentSetting{
		Enable:   true,
		Protocol: "openai",
		Model:    "test-model",
		ApiKey:   "test-key",
		BaseURL:  baseURL + "/v1",
	})
	return &CopilotService{
		durableKV:   kv,
		userReader:  &stubUserReader{user: userModel.User{ID: "u1", Username: "alice"}},
		echoService: echoSvc,
		embedding:   &stubEmbeddingSvc{},
		asks:        newAskRegistry(),
	}
}

// The approving path, end to end: the model asks to delete, the browser sees the
// question, clicks the affirmative, and the Echo is deleted — all inside one
// streaming request.
func TestRoundTrip_ApprovedDeleteRunsAndStreamResumes(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }

	srv := modelServer(t,
		[]string{toolCallChunk("call-1", "delete_echo", echoArgs(testEchoID))},
		[]string{textChunk("已经删掉了。")},
	)
	defer srv.Close()

	s := newRoundTripService(t, echoSvc, srv.URL)
	w := newSSEWriter()

	streamDone := make(chan error, 1)
	go func() {
		streamDone <- s.AskStream(helpers.CtxAsUser("u1"), "把那条天气的删掉", "zh-CN", "Asia/Shanghai", w)
	}()

	ask := awaitAsk(t, w, 10*time.Second)

	// The question has to name the Echo concretely; "delete Echo 019a…" is not
	// something a person can consent to.
	q := ask.Questions[0]
	if !strings.Contains(q.Detail, testEchoID) || !strings.Contains(q.Detail, "今天天气不错") {
		t.Fatalf("detail did not describe the Echo: %q", q.Detail)
	}
	if len(echoSvc.deleted) != 0 {
		t.Fatalf("the Echo was deleted before anyone answered: %v", echoSvc.deleted)
	}

	ws := writeStringsFor("zh-CN")
	if err := s.AnswerAsk(helpers.CtxAsUser("u1"), ask.AskID, []AskAnswer{
		{QuestionID: q.ID, Selected: []string{ws.DeleteYes}},
	}); err != nil {
		t.Fatalf("AnswerAsk: %v", err)
	}

	if err := <-streamDone; err != nil {
		t.Fatalf("AskStream: %v", err)
	}

	if len(echoSvc.deleted) != 1 || echoSvc.deleted[0] != testEchoID {
		t.Fatalf("deleted = %v, want [echo-1]", echoSvc.deleted)
	}

	body := w.body()
	// The turn continued in the same request: the model's sentence arrived after
	// the answer went back, on the same stream.
	if !strings.Contains(body, "已经删掉了。") {
		t.Fatalf("the run did not resume on the same stream:\n%s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Fatalf("stream did not end with done:\n%s", body)
	}

	// The exchange is on the turn, so reopening the conversation shows who
	// approved the deletion rather than only that it happened.
	session, err := s.GetSession(helpers.CtxAsUser("u1"))
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	assistant := session[len(session)-1]
	if len(assistant.Asks) != 1 {
		t.Fatalf("persisted asks = %+v, want the one exchange", assistant.Asks)
	}
	if got := assistant.Asks[0].Answers[0].Selected; len(got) != 1 || got[0] != ws.DeleteYes {
		t.Fatalf("persisted answer = %v, want [%s]", got, ws.DeleteYes)
	}
}

// Cancelling refuses that one write and lets the turn carry on, which is the
// difference between "don't delete this" and "stop answering me".
func TestRoundTrip_CancelledDeleteChangesNothing(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }

	srv := modelServer(t,
		[]string{toolCallChunk("call-1", "delete_echo", echoArgs(testEchoID))},
		[]string{textChunk("好，没有删。")},
	)
	defer srv.Close()

	s := newRoundTripService(t, echoSvc, srv.URL)
	w := newSSEWriter()

	streamDone := make(chan error, 1)
	go func() {
		streamDone <- s.AskStream(helpers.CtxAsUser("u1"), "删掉那条", "zh-CN", "Asia/Shanghai", w)
	}()

	ask := awaitAsk(t, w, 10*time.Second)
	ws := writeStringsFor("zh-CN")
	if err := s.AnswerAsk(helpers.CtxAsUser("u1"), ask.AskID, []AskAnswer{
		{QuestionID: ask.Questions[0].ID, Selected: []string{ws.Cancel}},
	}); err != nil {
		t.Fatalf("AnswerAsk: %v", err)
	}

	if err := <-streamDone; err != nil {
		t.Fatalf("AskStream: %v", err)
	}

	if len(echoSvc.deleted) != 0 {
		t.Fatalf("deleted = %v, want nothing", echoSvc.deleted)
	}
	if body := w.body(); !strings.Contains(body, "好，没有删。") {
		t.Fatalf("the run did not carry on after the refusal:\n%s", body)
	}
}

// A second click on the same round — a stale tab, a double-tap — is refused
// rather than allowed to overwrite the answer that was given.
func TestRoundTrip_SecondAnswerIsRefused(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }

	srv := modelServer(t,
		[]string{toolCallChunk("call-1", "delete_echo", echoArgs(testEchoID))},
		[]string{textChunk("已经删掉了。")},
	)
	defer srv.Close()

	s := newRoundTripService(t, echoSvc, srv.URL)
	w := newSSEWriter()

	streamDone := make(chan error, 1)
	go func() {
		streamDone <- s.AskStream(helpers.CtxAsUser("u1"), "删掉那条", "zh-CN", "Asia/Shanghai", w)
	}()

	ask := awaitAsk(t, w, 10*time.Second)
	ws := writeStringsFor("zh-CN")
	first := []AskAnswer{{QuestionID: ask.Questions[0].ID, Selected: []string{ws.DeleteYes}}}
	if err := s.AnswerAsk(helpers.CtxAsUser("u1"), ask.AskID, first); err != nil {
		t.Fatalf("first AnswerAsk: %v", err)
	}
	if err := s.AnswerAsk(helpers.CtxAsUser("u1"), ask.AskID, first); err == nil {
		t.Fatal("a second answer to the same round was accepted")
	}

	if err := <-streamDone; err != nil {
		t.Fatalf("AskStream: %v", err)
	}
	if len(echoSvc.deleted) != 1 {
		t.Fatalf("deleted = %v, want exactly one deletion", echoSvc.deleted)
	}
}

// Another account cannot answer this conversation's question, even knowing its
// id.
func TestRoundTrip_ForeignUserCannotAnswer(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }

	srv := modelServer(t,
		[]string{toolCallChunk("call-1", "delete_echo", echoArgs(testEchoID))},
		[]string{textChunk("好。")},
	)
	defer srv.Close()

	s := newRoundTripService(t, echoSvc, srv.URL)
	w := newSSEWriter()

	ctx, cancel := context.WithCancel(helpers.CtxAsUser("u1"))
	defer cancel()

	streamDone := make(chan error, 1)
	go func() { streamDone <- s.AskStream(ctx, "删掉那条", "zh-CN", "Asia/Shanghai", w) }()

	ask := awaitAsk(t, w, 10*time.Second)
	ws := writeStringsFor("zh-CN")
	if err := s.AnswerAsk(helpers.CtxAsUser("intruder"), ask.AskID, []AskAnswer{
		{QuestionID: ask.Questions[0].ID, Selected: []string{ws.DeleteYes}},
	}); err == nil {
		t.Fatal("another user answered this conversation's question")
	}
	if len(echoSvc.deleted) != 0 {
		t.Fatalf("deleted = %v, want nothing", echoSvc.deleted)
	}

	// Let the parked run go rather than waiting out its ask budget.
	cancel()
	<-streamDone
}

// ask_user over the wire: a plain question, answered with typed text, resumes the
// same turn with those words as the tool result.
func TestRoundTrip_AskUserAnsweredWithTypedText(t *testing.T) {
	srv := modelServer(t,
		[]string{toolCallChunk("call-1", "ask_user",
			`{"questions":[{"id":"tag","question":"用哪个标签？","options":[{"label":"读书"},{"label":"日常"}]}]}`)},
		[]string{textChunk("好，用「工程」。")},
	)
	defer srv.Close()

	s := newRoundTripService(t, &recordingEchoSvc{}, srv.URL)
	w := newSSEWriter()

	streamDone := make(chan error, 1)
	go func() {
		streamDone <- s.AskStream(helpers.CtxAsUser("u1"), "帮我想个标签", "zh-CN", "Asia/Shanghai", w)
	}()

	ask := awaitAsk(t, w, 10*time.Second)
	if len(ask.Questions[0].Options) != 2 {
		t.Fatalf("options = %+v, want the two the model offered", ask.Questions[0].Options)
	}
	if err := s.AnswerAsk(helpers.CtxAsUser("u1"), ask.AskID, []AskAnswer{
		{QuestionID: ask.Questions[0].ID, Custom: "工程"},
	}); err != nil {
		t.Fatalf("AnswerAsk: %v", err)
	}

	if err := <-streamDone; err != nil {
		t.Fatalf("AskStream: %v", err)
	}
	if body := w.body(); !strings.Contains(body, "好，用「工程」。") {
		t.Fatalf("the run did not resume with the typed answer:\n%s", body)
	}
}
