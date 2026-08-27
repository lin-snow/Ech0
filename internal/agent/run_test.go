// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	model "github.com/lin-snow/ech0/internal/model/setting"
)

type fakeProvider struct {
	scripts [][]Event
	calls   int
	gotReqs []Request
}

func (p *fakeProvider) Complete(_ context.Context, _ Request) (Response, error) {
	return Response{}, errors.New("not used")
}

func (p *fakeProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	p.gotReqs = append(p.gotReqs, req)
	var events []Event
	if p.calls < len(p.scripts) {
		events = p.scripts[p.calls]
	}
	p.calls++

	ch := make(chan Event)
	go func() {
		defer close(ch)
		for _, ev := range events {
			select {
			case ch <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func drain(out <-chan AgentEvent) []AgentEvent {
	var evs []AgentEvent
	for ev := range out {
		evs = append(evs, ev)
	}
	return evs
}

func countingTool(name string, output ToolOutput, err error) (Tool, *int) {
	calls := 0
	t := Tool{
		Def: ToolDef{Name: name, Description: "test tool", Parameters: json.RawMessage(`{"type":"object"}`)},
		Execute: func(_ context.Context, _ json.RawMessage) (ToolOutput, error) {
			calls++
			return output, err
		},
	}
	return t, &calls
}

func toolCallEvent(id, name, args string) Event {
	return Event{Kind: EventToolCall, ToolCall: ToolCall{ID: id, Name: name, Args: json.RawMessage(args)}}
}
func textEvent(s string) Event { return Event{Kind: EventTextDelta, Text: s} }
func doneEvent() Event         { return Event{Kind: EventDone} }
func errEvent(err error) Event { return Event{Kind: EventError, Err: err} }

func runLoopSync(ctx context.Context, provider Provider, req RunRequest) []AgentEvent {
	return drain(runChan(ctx, provider, req))
}

func kinds(evs []AgentEvent) []AgentEventKind {
	ks := make([]AgentEventKind, len(evs))
	for i, e := range evs {
		ks[i] = e.Kind
	}
	return ks
}

func countKind(evs []AgentEvent, k AgentEventKind) int {
	n := 0
	for _, e := range evs {
		if e.Kind == k {
			n++
		}
	}
	return n
}

func enabledSetting() model.AgentSetting {
	return model.AgentSetting{Enable: true, Protocol: "openai", Model: "gpt-test", ApiKey: "k"}
}

func TestRunLoop_MultiRoundHappyPath(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{Content: "hit", Meta: "meta"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{
		Setting: enabledSetting(),
		Tools:   []Tool{tool},
	})

	want := []AgentEventKind{AgentSearching, AgentToolResult, AgentDelta, AgentDone}
	got := kinds(evs)
	if len(got) != len(want) {
		t.Fatalf("event kinds = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event[%d] kind = %d, want %d (full: %v)", i, got[i], want[i], got)
		}
	}
	if *execs != 1 {
		t.Fatalf("tool executed %d times, want 1", *execs)
	}
	if fp.calls != 2 {
		t.Fatalf("provider.Stream called %d times, want 2", fp.calls)
	}
	if evs[2].Text != "answer" {
		t.Fatalf("delta text = %q, want %q", evs[2].Text, "answer")
	}
}

func TestRunLoop_ToolDedup(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"same"}`), doneEvent()},
		{toolCallEvent("c2", "search_echos", `{"q":"same"}`), doneEvent()},
		{textEvent("done"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if *execs != 1 {
		t.Fatalf("tool executed %d times, want 1 (dedup)", *execs)
	}
	if n := countKind(evs, AgentSearching); n != 1 {
		t.Fatalf("AgentSearching count = %d, want 1", n)
	}
	if n := countKind(evs, AgentToolResult); n != 1 {
		t.Fatalf("AgentToolResult count = %d, want 1", n)
	}
	if n := countKind(evs, AgentDone); n != 1 {
		t.Fatalf("AgentDone count = %d, want 1", n)
	}
}

func TestRunLoop_MaxRoundsForcesFinalNoToolRound(t *testing.T) {
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("forced answer"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{
		Setting:   enabledSetting(),
		Tools:     []Tool{tool},
		MaxRounds: 1,
	})

	if fp.calls != 2 {
		t.Fatalf("provider.Stream called %d times, want 2 (1 tool round + 1 forced)", fp.calls)
	}
	if len(fp.gotReqs) != 2 {
		t.Fatalf("captured %d requests, want 2", len(fp.gotReqs))
	}
	if len(fp.gotReqs[0].Tools) != 1 {
		t.Fatalf("first round Tools len = %d, want 1", len(fp.gotReqs[0].Tools))
	}
	if fp.gotReqs[1].Tools != nil {
		t.Fatalf("forced final round must pass nil Tools, got %v", fp.gotReqs[1].Tools)
	}
	if evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("last event = %d, want AgentDone", evs[len(evs)-1].Kind)
	}
}

func TestRunLoop_ToolExecErrorFedBack(t *testing.T) {
	tool, execs := countingTool("search_echos", ToolOutput{}, errors.New("boom"))
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("recovered"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if *execs != 1 {
		t.Fatalf("tool executed %d times, want 1", *execs)
	}
	if n := countKind(evs, AgentError); n != 0 {
		t.Fatalf("AgentError count = %d, want 0 (exec error must not abort)", n)
	}
	if n := countKind(evs, AgentSearching); n != 1 {
		t.Fatalf("AgentSearching count = %d, want 1", n)
	}
	if n := countKind(evs, AgentToolResult); n != 0 {
		t.Fatalf("AgentToolResult count = %d, want 0 (failed exec emits no ToolResult)", n)
	}
	if evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("last event = %d, want AgentDone", evs[len(evs)-1].Kind)
	}
}

func TestRunLoop_UnknownTool(t *testing.T) {
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit"}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "nope", `{}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if n := countKind(evs, AgentSearching); n != 0 {
		t.Fatalf("AgentSearching count = %d, want 0 (unknown tool short-circuits)", n)
	}
	if evs[len(evs)-1].Kind != AgentDone {
		t.Fatalf("last event = %d, want AgentDone", evs[len(evs)-1].Kind)
	}
}

func TestRunLoop_TransportError(t *testing.T) {
	fp := &fakeProvider{scripts: [][]Event{
		{errEvent(errors.New("network down"))},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting()})

	if len(evs) != 1 || evs[0].Kind != AgentError {
		t.Fatalf("events = %v, want single AgentError", kinds(evs))
	}
	if evs[0].Err == nil {
		t.Fatalf("AgentError must carry the error")
	}
	if fp.calls != 1 {
		t.Fatalf("provider.Stream called %d times, want 1", fp.calls)
	}
}

func TestRunLoop_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	fp := &fakeProvider{scripts: [][]Event{
		{textEvent("never delivered"), doneEvent()},
	}}

	done := make(chan struct{})
	go func() {
		drain(runChan(ctx, fp, RunRequest{Setting: enabledSetting()}))
		close(done)
	}()

	<-done
}

func runChan(ctx context.Context, provider Provider, req RunRequest) <-chan AgentEvent {
	out := make(chan AgentEvent)
	go runLoop(ctx, provider, req, out)
	return out
}

func TestRun_TimeoutAborts(t *testing.T) {
	out, err := Run(context.Background(), RunRequest{
		Setting: enabledSetting(),
		Timeout: time.Nanosecond,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	drain(out)
}

func TestRunStrings_WithDefaults(t *testing.T) {
	if got := (RunStrings{}).withDefaults(); got != defaultRunStrings {
		t.Fatalf("empty RunStrings should equal defaults, got %+v", got)
	}
	got := RunStrings{DedupNote: "EN dedup", ImageNote: "EN image"}.withDefaults()
	if got.DedupNote != "EN dedup" || got.ImageNote != "EN image" {
		t.Fatalf("provided fields must be preserved, got %+v", got)
	}
	if got.UnknownTool != defaultRunStrings.UnknownTool || got.ToolError != defaultRunStrings.ToolError {
		t.Fatalf("empty fields must fall back to defaults, got %+v", got)
	}
}

func TestRunLoop_CustomImageNote(t *testing.T) {
	const enNote = "(custom english image note)"
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit", Images: []ImagePart{{MediaType: "image/png", Base64: "abc"}}}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{
		Setting: enabledSetting(),
		Tools:   []Tool{tool},
		Strings: RunStrings{ImageNote: enNote},
	})

	found := false
	for _, m := range fp.gotReqs[1].Messages {
		if m.Role == RoleUser && m.Content == enNote {
			found = true
		}
	}
	if !found {
		t.Fatalf("custom ImageNote %q should appear in next round messages", enNote)
	}
}

func TestRunLoop_ParallelToolsSingleRound(t *testing.T) {
	toolA := Tool{
		Def:     ToolDef{Name: "tool_a", Description: "a", Parameters: json.RawMessage(`{"type":"object"}`)},
		Execute: func(_ context.Context, _ json.RawMessage) (ToolOutput, error) { return ToolOutput{Content: "AAA"}, nil },
	}
	toolB := Tool{
		Def:     ToolDef{Name: "tool_b", Description: "b", Parameters: json.RawMessage(`{"type":"object"}`)},
		Execute: func(_ context.Context, _ json.RawMessage) (ToolOutput, error) { return ToolOutput{Content: "BBB"}, nil },
	}
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "tool_a", `{}`), toolCallEvent("c2", "tool_b", `{}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{toolA, toolB}})

	if n := countKind(evs, AgentSearching); n != 2 {
		t.Fatalf("AgentSearching count = %d, want 2", n)
	}
	if n := countKind(evs, AgentToolResult); n != 2 {
		t.Fatalf("AgentToolResult count = %d, want 2", n)
	}

	var toolMsgs []Message
	for _, m := range fp.gotReqs[1].Messages {
		if m.Role == RoleTool {
			toolMsgs = append(toolMsgs, m)
		}
	}
	if len(toolMsgs) != 2 {
		t.Fatalf("next round should carry 2 tool results, got %d", len(toolMsgs))
	}
	if toolMsgs[0].ToolCallID != "c1" || toolMsgs[0].Content != "AAA" {
		t.Fatalf("first tool result = %+v, want c1/AAA", toolMsgs[0])
	}
	if toolMsgs[1].ToolCallID != "c2" || toolMsgs[1].Content != "BBB" {
		t.Fatalf("second tool result = %+v, want c2/BBB", toolMsgs[1])
	}
}

func TestRunLoop_TrimsOldestToolResultOverBudget(t *testing.T) {
	note := defaultRunStrings.ContextTrimNote
	msgs := []Message{
		{Role: RoleSystem, Content: "S"},
		{Role: RoleUser, Content: "Q"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "t1", Name: "search_echos"}}},
		{Role: RoleTool, ToolCallID: "t1", Content: strings.Repeat("a", 100)},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "t2", Name: "search_echos"}}},
		{Role: RoleTool, ToolCallID: "t2", Content: strings.Repeat("b", 20)},
	}
	fp := &fakeProvider{scripts: [][]Event{{textEvent("ok"), doneEvent()}}}

	runLoopSync(context.Background(), fp, RunRequest{
		Setting:          enabledSetting(),
		Messages:         msgs,
		MaxContextTokens: 50,
	})

	got := fp.gotReqs[0].Messages
	if got[3].Content != note {
		t.Fatalf("oldest tool result should be trimmed to note, got %q", got[3].Content)
	}
	if got[5].Content != strings.Repeat("b", 20) {
		t.Fatalf("recent tool result should be preserved, got %q", got[5].Content)
	}
}

func TestRunLoop_ToolImageNoteAppended(t *testing.T) {
	img := ImagePart{MediaType: "image/png", Base64: "abc"}
	tool, _ := countingTool("search_echos", ToolOutput{Content: "hit", Images: []ImagePart{img}}, nil)
	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "search_echos", `{"q":"x"}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if len(fp.gotReqs) != 2 {
		t.Fatalf("captured %d requests, want 2", len(fp.gotReqs))
	}
	var found *Message
	for i := range fp.gotReqs[1].Messages {
		m := fp.gotReqs[1].Messages[i]
		if m.Role == RoleUser && m.Content == toolImageNote {
			found = &fp.gotReqs[1].Messages[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("next round messages should contain the toolImageNote user message")
	} else if len(found.Images) != 1 || found.Images[0].Base64 != "abc" {
		t.Fatalf("toolImageNote message should carry the tool's image, got %+v", found.Images)
	}
}
