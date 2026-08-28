// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

func newTestAsker(budget time.Duration) (*asker, chan askEvent) {
	events := make(chan askEvent, 8)
	return &asker{
		registry: newAskRegistry(),
		events:   events,
		userID:   "u1",
		budget:   budget,
		strs:     askStringsFor("zh-CN"),
	}, events
}

// answerWhenAsked plays the browser: it waits for the ask event, then posts the
// reply the way the answer endpoint would.
func answerWhenAsked(t *testing.T, a *asker, events <-chan askEvent, reply func(Ask) []AskAnswer) {
	t.Helper()
	go func() {
		ev := <-events
		if ev.Open == nil {
			t.Errorf("first ask event carried no round")
			return
		}
		if err := a.registry.answer(ev.Open.AskID, "u1", reply(*ev.Open)); err != nil {
			t.Errorf("answer: %v", err)
		}
	}()
}

func TestAsker_AnswerResumesTheParkedRun(t *testing.T) {
	a, events := newTestAsker(2 * time.Second)
	answerWhenAsked(t, a, events, func(ask Ask) []AskAnswer {
		return []AskAnswer{{QuestionID: ask.Questions[0].ID, Selected: []string{"读书"}}}
	})

	answers, err := a.Ask(context.Background(), []AskQuestion{{ID: "tag", Text: "用哪个标签？"}})
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if len(answers) != 1 || answers[0].Selected[0] != "读书" {
		t.Fatalf("answers = %+v, want the reply that was posted", answers)
	}
	if ex := a.exchanges(); len(ex) != 1 || len(ex[0].Answers) != 1 {
		t.Fatalf("exchanges = %+v, want the finished round recorded", ex)
	}
}

// The round has to be registered before it is shown. Reversed, a client fast
// enough to answer before the slot existed would have its click refused and
// would never be asked again.
func TestAsker_RoundIsAnswerableWhenTheEventArrives(t *testing.T) {
	a, events := newTestAsker(2 * time.Second)

	done := make(chan error, 1)
	go func() {
		_, err := a.Ask(context.Background(), []AskQuestion{{ID: "q", Text: "问题"}})
		done <- err
	}()

	ev := <-events
	if ev.Open == nil {
		t.Fatal("no round on the ask event")
	}
	// Answered the instant the event lands: the slot must already be there.
	if err := a.registry.answer(ev.Open.AskID, "u1", []AskAnswer{{QuestionID: "q", Custom: "好"}}); err != nil {
		t.Fatalf("answer immediately after the event: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("Ask: %v", err)
	}
}

func TestAsker_ExpiryIsNotAnAnswer(t *testing.T) {
	a, events := newTestAsker(30 * time.Millisecond)

	q := AskQuestion{ID: "confirm", Text: "删除？", Options: []AskOption{{Label: "删除"}, {Label: "取消"}}}
	_, err := a.Ask(context.Background(), []AskQuestion{q})
	if err == nil {
		t.Fatal("Ask returned no error after the budget expired")
	}
	if len(a.exchanges()) != 0 {
		t.Fatal("an unanswered round was recorded as an exchange")
	}

	var opened, closed string
	for len(events) > 0 {
		ev := <-events
		if ev.Open != nil {
			opened = ev.Open.AskID
		}
		if ev.Closed != "" {
			closed = ev.Closed
		}
	}
	if opened == "" || closed != opened {
		t.Fatalf("expected ask_closed for %q, got %q", opened, closed)
	}
}

func TestAsker_CancelledContextEndsTheWait(t *testing.T) {
	a, _ := newTestAsker(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := a.Ask(ctx, []AskQuestion{{ID: "q", Text: "问题"}}); err == nil {
		t.Fatal("Ask returned no error on a cancelled context")
	}
}

func TestAsker_EmptyRoundIsRefused(t *testing.T) {
	a, _ := newTestAsker(time.Second)
	if _, err := a.Ask(context.Background(), []AskQuestion{{ID: "q", Text: "   "}}); err == nil {
		t.Fatal("a round with nothing answerable in it should be refused")
	}
}

func TestAskRegistry_AnswerableExactlyOnce(t *testing.T) {
	r := newAskRegistry()
	p := r.open("ask-1", "u1")

	if err := r.answer("ask-1", "u1", []AskAnswer{{QuestionID: "q", Custom: "first"}}); err != nil {
		t.Fatalf("first answer: %v", err)
	}
	if err := r.answer("ask-1", "u1", []AskAnswer{{QuestionID: "q", Custom: "second"}}); err == nil {
		t.Fatal("a second answer to the same round must be refused")
	}

	got := <-p.answers
	if got[0].Custom != "first" {
		t.Fatalf("delivered %q, want the first answer", got[0].Custom)
	}
}

func TestAskRegistry_RefusesUnknownAndForeignRounds(t *testing.T) {
	r := newAskRegistry()
	r.open("ask-1", "u1")

	if err := r.answer("ask-2", "u1", []AskAnswer{{QuestionID: "q"}}); err == nil {
		t.Fatal("an unknown round must be refused")
	}
	if err := r.answer("ask-1", "u2", []AskAnswer{{QuestionID: "q"}}); err == nil {
		t.Fatal("another user's round must be refused")
	}
	if err := r.answer("ask-1", "u1", []AskAnswer{{QuestionID: "q"}}); err != nil {
		t.Fatalf("the owner's answer should still work: %v", err)
	}
}

func TestAskRegistry_ClosedRoundIsNoLongerAnswerable(t *testing.T) {
	r := newAskRegistry()
	r.open("ask-1", "u1")
	r.close("ask-1")

	if err := r.answer("ask-1", "u1", []AskAnswer{{QuestionID: "q"}}); err == nil {
		t.Fatal("a closed round must be refused")
	}
}

// consented is the one place a label is compared, and everything that is not an
// exact pick of the affirmative must read as a no.
func TestConsented(t *testing.T) {
	q := AskQuestion{
		ID:      "delete_echo",
		Options: []AskOption{{Label: "删除"}, {Label: "取消"}},
	}

	cases := []struct {
		name     string
		answers  []AskAnswer
		approved bool
		note     string
	}{
		{"affirmative pick", []AskAnswer{{QuestionID: "delete_echo", Selected: []string{"删除"}}}, true, ""},
		{"cancel pick", []AskAnswer{{QuestionID: "delete_echo", Selected: []string{"取消"}}}, false, ""},
		{"nothing picked", []AskAnswer{{QuestionID: "delete_echo"}}, false, ""},
		{"no answer at all", nil, false, ""},
		{"answer to another question", []AskAnswer{{QuestionID: "other", Selected: []string{"删除"}}}, false, ""},
		{"both options picked", []AskAnswer{{QuestionID: "delete_echo", Selected: []string{"删除", "取消"}}}, false, ""},
		{
			"typed agreement is not consent",
			[]AskAnswer{{QuestionID: "delete_echo", Custom: "删除"}},
			false, "删除",
		},
		{
			"typed instruction comes back as the reason",
			[]AskAnswer{{QuestionID: "delete_echo", Selected: []string{"删除"}, Custom: "别删，改成 #读书"}},
			false, "别删，改成 #读书",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			approved, note := consented(q, c.answers)
			if approved != c.approved {
				t.Fatalf("approved = %v, want %v", approved, c.approved)
			}
			if note != c.note {
				t.Fatalf("note = %q, want %q", note, c.note)
			}
		})
	}
}

// A question with no options cannot be consent — there is no affirmative label
// to pick, so a write tool that forgot its options refuses rather than accepting
// whatever came back.
func TestConsented_OptionlessQuestionIsNeverApproved(t *testing.T) {
	q := AskQuestion{ID: "confirm"}
	if approved, _ := consented(q, []AskAnswer{{QuestionID: "confirm", Selected: []string{"yes"}}}); approved {
		t.Fatal("an option-less confirmation was approved")
	}
}

func TestClampAskQuestions(t *testing.T) {
	four := 4
	qs := clampAskQuestions([]AskQuestion{
		{Text: "第一个问题没有 id"},
		{ID: "b", Text: "有效", Options: []AskOption{
			{Label: "a"},
			{Label: "a"},
			{Label: "b"},
			{Label: "c"},
			{Label: "d"},
			{Label: "e"},
			{Label: "f"},
			{Label: "g"},
		}, Recommended: &four},
		{ID: "c", Text: "  "},
		{ID: "d", Text: "d"},
		{ID: "e", Text: "e"},
		{ID: "f", Text: "f"},
		{ID: "g", Text: "g"},
	})

	if len(qs) != maxAskQuestions {
		t.Fatalf("kept %d questions, want %d", len(qs), maxAskQuestions)
	}
	if qs[0].ID != "q1" {
		t.Fatalf("missing id fell back to %q, want q1", qs[0].ID)
	}
	if len(qs[1].Options) != maxAskOptions {
		t.Fatalf("kept %d options, want %d", len(qs[1].Options), maxAskOptions)
	}
	if qs[1].Options[0].Label == qs[1].Options[1].Label {
		t.Fatal("duplicate option labels survived clamping")
	}
	if qs[1].Recommended == nil || *qs[1].Recommended != 4 {
		t.Fatalf("recommended = %v, want 4 (in range after clamping)", qs[1].Recommended)
	}
}

func TestClampRecommended_DropsOutOfRange(t *testing.T) {
	nine := 9
	neg := -1
	if got := clampRecommended(&nine, 2); got != nil {
		t.Fatalf("recommended 9 over 2 options = %v, want nil", got)
	}
	if got := clampRecommended(&neg, 2); got != nil {
		t.Fatalf("negative recommended = %v, want nil", got)
	}
}

// A write tool end to end: the confirmation is what stands between the model's
// call and the deletion, and the loop is what invokes it.
type recordingEchoSvc struct {
	stubEchoSvc
	deleted []string
	posted  []echoModel.Echo
	updated []echoModel.Echo
}

func (f *recordingEchoSvc) DeleteEchoById(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

func (f *recordingEchoSvc) PostEcho(_ context.Context, echo *echoModel.Echo) error {
	echo.ID = "new-echo"
	f.posted = append(f.posted, *echo)
	return nil
}

func (f *recordingEchoSvc) UpdateEcho(_ context.Context, echo *echoModel.Echo) error {
	f.updated = append(f.updated, *echo)
	return nil
}

// testEchoID is a real UUID because the write tools require one: an id that is
// not a UUID is refused before anything is asked, which is the behaviour that
// keeps a positional 【1】 from ever reaching the database.
const testEchoID = "019ce0ea-82dd-774f-ae2d-5445512d42ad"

// echoArgs builds a write tool's arguments around that id.
func echoArgs(id string, extra ...string) string {
	if len(extra) == 0 {
		return fmt.Sprintf(`{"id":%q}`, id)
	}
	return fmt.Sprintf(`{"id":%q,%s}`, id, strings.Join(extra, ","))
}

func existingEcho() *echoModel.Echo {
	return &echoModel.Echo{
		ID:        testEchoID,
		Content:   "今天天气不错",
		Private:   false,
		Tags:      []echoModel.Tag{{ID: "t1", Name: "日常"}},
		CreatedAt: ts("2026-01-02"),
		EchoFiles: []echoModel.EchoFile{{ID: "f1"}},
		Layout:    echoModel.LayoutGrid,
		Extension: &echoModel.EchoExtension{
			ID:      "x1",
			Type:    echoModel.Extension_MUSIC,
			Payload: map[string]any{"url": "https://example.com/a.mp3"},
		},
	}
}

// runMutation drives a mutating tool the way the loop does, so these tests
// exercise the same ordering the gate enforces rather than a shortcut past it.
func runMutation(t *testing.T, tool agent.Tool, args string) (agent.ToolOutput, error) {
	t.Helper()
	if tool.Effect != agent.EffectMutate {
		t.Fatalf("tool %q declared effect %d, want EffectMutate", tool.Def.Name, tool.Effect)
	}
	m := tool.Mutation
	if m == nil || m.Plan == nil || m.Confirm == nil {
		t.Fatalf("tool %q is not a complete mutation", tool.Def.Name)
	}
	plan, err := m.Plan(context.Background(), []byte(args))
	if err != nil {
		return agent.ToolOutput{}, err
	}
	if plan.Apply == nil {
		t.Fatalf("tool %q planned nothing to apply", tool.Def.Name)
	}
	decision, err := m.Confirm(context.Background(), plan)
	if err != nil {
		return agent.ToolOutput{}, err
	}
	if !decision.Approved {
		return agent.ToolOutput{Content: decision.Refusal}, nil
	}
	return plan.Apply(context.Background())
}

func TestDeleteEchoTool_DeletesOnlyAfterApproval(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}

	a, events := newTestAsker(2 * time.Second)
	a.registry = s.asks
	ws := writeStringsFor("zh-CN")
	answerWhenAsked(t, a, events, func(ask Ask) []AskAnswer {
		q := ask.Questions[0]
		if q.Detail == "" {
			t.Errorf("the delete confirmation carried no detail block")
		}
		if q.Options[0].Label != ws.DeleteYes {
			t.Errorf("affirmative label = %q, want %q", q.Options[0].Label, ws.DeleteYes)
		}
		return []AskAnswer{{QuestionID: q.ID, Selected: []string{ws.DeleteYes}}}
	})

	out, err := runMutation(t, s.deleteEchoTool(a, "zh-CN", time.UTC), echoArgs(testEchoID))
	if err != nil {
		t.Fatalf("delete_echo: %v", err)
	}
	if len(echoSvc.deleted) != 1 || echoSvc.deleted[0] != testEchoID {
		t.Fatalf("deleted = %v, want [echo-1]", echoSvc.deleted)
	}
	if out.Content == "" {
		t.Fatal("the model was told nothing after the deletion")
	}
}

func TestDeleteEchoTool_CancelLeavesTheEchoAlone(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}

	a, events := newTestAsker(2 * time.Second)
	a.registry = s.asks
	ws := writeStringsFor("zh-CN")
	answerWhenAsked(t, a, events, func(ask Ask) []AskAnswer {
		return []AskAnswer{{QuestionID: ask.Questions[0].ID, Selected: []string{ws.Cancel}}}
	})

	out, err := runMutation(t, s.deleteEchoTool(a, "zh-CN", time.UTC), echoArgs(testEchoID))
	if err != nil {
		t.Fatalf("delete_echo: %v", err)
	}
	if len(echoSvc.deleted) != 0 {
		t.Fatalf("deleted = %v, want nothing", echoSvc.deleted)
	}
	if out.Content != askStringsFor("zh-CN").declined("") {
		t.Fatalf("model was told %q, want the declined note", out.Content)
	}
}

// The bug this guards: search results are numbered 【1】【2】 for the model to
// name in prose, and a model told to pass "the id" reaches for that number.
// Handed straight to GetEchoById it resolves to nothing and the Echo reads as
// missing — a wrong answer, when the truth is that the model quoted the wrong
// field. Nothing is read, nothing is asked, and the refusal says what to do.
func TestWriteTools_RejectNonUUIDIDs(t *testing.T) {
	ws := writeStringsFor("zh-CN")
	bad := []string{`{"id":"1"}`, `{"id":"【1】"}`, `{"id":"echo-1"}`, `{"id":"  "}`, `{}`}

	for _, args := range bad {
		t.Run(args, func(t *testing.T) {
			echoSvc := &recordingEchoSvc{}
			reads := 0
			echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) {
				reads++
				return existingEcho(), nil
			}
			s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}
			a, events := newTestAsker(time.Second)
			a.registry = s.asks

			for _, tool := range []agent.Tool{
				s.deleteEchoTool(a, "zh-CN", time.UTC),
				s.updateEchoTool(a, "zh-CN", time.UTC),
			} {
				_, err := runMutation(t, tool, args)
				if err == nil {
					t.Fatalf("%s accepted %s", tool.Def.Name, args)
				}
				if err.Error() != ws.BadEchoID {
					t.Fatalf("%s said %q, want the instruction that names the id= field", tool.Def.Name, err)
				}
			}
			if reads != 0 {
				t.Fatalf("the Echo was read %d times for an id that is not one", reads)
			}
			if len(events) != 0 {
				t.Fatal("a confirmation was shown for an id that is not one")
			}
			if len(echoSvc.deleted) != 0 || len(echoSvc.updated) != 0 {
				t.Fatal("a write happened for an id that is not one")
			}
		})
	}
}

// Nobody answering must leave the Echo alone as surely as a cancel does.
func TestDeleteEchoTool_ExpiryLeavesTheEchoAlone(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}

	a, _ := newTestAsker(20 * time.Millisecond)
	a.registry = s.asks

	if _, err := runMutation(t, s.deleteEchoTool(a, "zh-CN", time.UTC), echoArgs(testEchoID)); err == nil {
		t.Fatal("delete_echo returned no error when nobody answered")
	}
	if len(echoSvc.deleted) != 0 {
		t.Fatalf("deleted = %v, want nothing", echoSvc.deleted)
	}
}

func TestDeleteEchoTool_UnknownEchoNeverAsks(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) {
		return nil, errors.New(commonModel.ECHO_NOT_FOUND)
	}
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}
	a, events := newTestAsker(time.Second)
	a.registry = s.asks

	if _, err := runMutation(t, s.deleteEchoTool(a, "zh-CN", time.UTC), `{"id":"nope"}`); err == nil {
		t.Fatal("delete_echo should fail when the Echo cannot be read")
	}
	if len(events) != 0 {
		t.Fatal("a question was shown for an Echo that could not be read")
	}
}

func TestCreateEchoTool_PostsTheConfirmedContent(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}

	a, events := newTestAsker(2 * time.Second)
	a.registry = s.asks
	ws := writeStringsFor("zh-CN")
	answerWhenAsked(t, a, events, func(ask Ask) []AskAnswer {
		return []AskAnswer{{QuestionID: ask.Questions[0].ID, Selected: []string{ws.CreateYes}}}
	})

	_, err := runMutation(t, s.createEchoTool(a, "zh-CN", time.UTC),
		`{"content":"  写点什么  ","tags":["#读书"," 日常 ","读书"],"private":true}`)
	if err != nil {
		t.Fatalf("create_echo: %v", err)
	}
	if len(echoSvc.posted) != 1 {
		t.Fatalf("posted %d echos, want 1", len(echoSvc.posted))
	}
	got := echoSvc.posted[0]
	if got.Content != "写点什么" {
		t.Fatalf("content = %q, want the trimmed text", got.Content)
	}
	if !got.Private {
		t.Fatal("private was not carried through")
	}
	if len(got.Tags) != 2 || got.Tags[0].Name != "读书" || got.Tags[1].Name != "日常" {
		t.Fatalf("tags = %+v, want the hash stripped and the duplicate dropped", got.Tags)
	}
}

func TestCreateEchoTool_EmptyContentNeverAsks(t *testing.T) {
	s := &CopilotService{echoService: &recordingEchoSvc{}, asks: newAskRegistry()}
	a, events := newTestAsker(time.Second)
	a.registry = s.asks

	if _, err := runMutation(t, s.createEchoTool(a, "zh-CN", time.UTC), `{"content":"   "}`); err == nil {
		t.Fatal("create_echo should refuse empty content")
	}
	if len(events) != 0 {
		t.Fatal("a question was shown for an empty Echo")
	}
}

// UpdateEcho replaces files wholesale, so anything the confirmation did not show
// the person has to survive the write. A chat that silently dropped an Echo's
// images would be a chat that destroyed them.
func TestUpdateEchoTool_CarriesUnshownFieldsThrough(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}

	a, events := newTestAsker(2 * time.Second)
	a.registry = s.asks
	ws := writeStringsFor("zh-CN")
	answerWhenAsked(t, a, events, func(ask Ask) []AskAnswer {
		return []AskAnswer{{QuestionID: ask.Questions[0].ID, Selected: []string{ws.UpdateYes}}}
	})

	if _, err := runMutation(t, s.updateEchoTool(a, "zh-CN", time.UTC),
		echoArgs(testEchoID, `"content":"改过的内容"`)); err != nil {
		t.Fatalf("update_echo: %v", err)
	}
	if len(echoSvc.updated) != 1 {
		t.Fatalf("updated %d echos, want 1", len(echoSvc.updated))
	}
	got := echoSvc.updated[0]
	if got.Content != "改过的内容" {
		t.Fatalf("content = %q, want the new text", got.Content)
	}
	if len(got.EchoFiles) != 1 || got.EchoFiles[0].ID != "f1" {
		t.Fatalf("files = %+v, want the existing file carried through", got.EchoFiles)
	}
	// UpdateEcho deletes and re-creates the extension row from what it is handed,
	// so an extension this tool never showed the person has to arrive back intact
	// or editing a caption would destroy the Echo's music card.
	if got.Extension == nil || got.Extension.Type != echoModel.Extension_MUSIC {
		t.Fatalf("extension = %+v, want the existing music card carried through", got.Extension)
	}
	if got.Layout != echoModel.LayoutGrid {
		t.Fatalf("layout = %q, want the existing layout carried through", got.Layout)
	}
	if len(got.Tags) != 1 || got.Tags[0].Name != "日常" {
		t.Fatalf("tags = %+v, want the existing tags untouched", got.Tags)
	}
}

// The fields an Echo really has but these tools cannot write. encoding/json
// drops unknown fields without a word, which here would end the run with the
// tool reporting a posted Echo and the person reading that their images went
// with it. Neither happened, so the call has to fail.
func TestWriteTools_RefuseFieldsTheyCannotWrite(t *testing.T) {
	ws := writeStringsFor("zh-CN")
	unsupported := map[string]string{
		"files":      `"files":[{"id":"f9"}]`,
		"echo_files": `"echo_files":[{"id":"f9"}]`,
		"images":     `"images":["/uploads/a.png"]`,
		"extension":  `"extension":{"type":"MUSIC","payload":{}}`,
		"layout":     `"layout":"grid"`,
	}

	for name, field := range unsupported {
		t.Run(name, func(t *testing.T) {
			echoSvc := &recordingEchoSvc{}
			echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }
			s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}
			a, events := newTestAsker(time.Second)
			a.registry = s.asks

			args := map[string]string{
				"create_echo": `{"content":"写点什么",` + field + `}`,
				"update_echo": echoArgs(testEchoID, `"content":"改过的内容"`, field),
			}
			for _, tool := range []agent.Tool{
				s.createEchoTool(a, "zh-CN", time.UTC),
				s.updateEchoTool(a, "zh-CN", time.UTC),
			} {
				_, err := runMutation(t, tool, args[tool.Def.Name])
				if err == nil {
					t.Fatalf("%s accepted %s and would have dropped it silently", tool.Def.Name, name)
				}
				if !strings.Contains(err.Error(), ws.UnsupportedArg) {
					t.Fatalf("%s said %q, want the refusal that says attachments cannot be written here", tool.Def.Name, err)
				}
				if !strings.Contains(err.Error(), name) {
					t.Fatalf("%s refused without naming the field: %q", tool.Def.Name, err)
				}
			}
			if len(events) != 0 {
				t.Fatal("a confirmation was shown for a call the tool cannot honour")
			}
			if len(echoSvc.posted) != 0 || len(echoSvc.updated) != 0 {
				t.Fatal("a write happened for a call the tool cannot honour")
			}
		})
	}
}

// Malformed JSON and a wrong type are the model's own mistakes and must read as
// themselves: dressed up as the attachment refusal they would send it hunting
// for a field it never passed.
func TestWriteTools_MalformedArgsAreNotBlamedOnAttachments(t *testing.T) {
	ws := writeStringsFor("zh-CN")
	s := &CopilotService{echoService: &recordingEchoSvc{}, asks: newAskRegistry()}
	a, _ := newTestAsker(time.Second)
	a.registry = s.asks

	for _, args := range []string{`{"content":123}`, `{"content":`, `not json at all`} {
		_, err := runMutation(t, s.createEchoTool(a, "zh-CN", time.UTC), args)
		if err == nil {
			t.Fatalf("create_echo accepted %s", args)
		}
		if strings.Contains(err.Error(), ws.UnsupportedArg) {
			t.Fatalf("%s was reported as an unsupported field: %q", args, err)
		}
	}
}

// A write tool's schema and its decoder have to agree. The decoder refuses
// unknown fields, so a schema that quietly permits them hands the model a
// contract the tool will not honour — it would pass files, be told no, and have
// no way to see that the schema misled it.
//
// This also catches a hand-edited schema string that stopped being JSON. Those
// consts are opaque to the compiler, and the next thing to notice would be a
// provider rejecting the entire chat request.
func TestWriteTools_SchemasMatchTheDecoder(t *testing.T) {
	s := &CopilotService{echoService: &recordingEchoSvc{}, asks: newAskRegistry()}
	a, _ := newTestAsker(time.Second)
	a.registry = s.asks

	for _, tool := range []agent.Tool{
		s.createEchoTool(a, "zh-CN", time.UTC),
		s.updateEchoTool(a, "zh-CN", time.UTC),
		s.deleteEchoTool(a, "zh-CN", time.UTC),
	} {
		t.Run(tool.Def.Name, func(t *testing.T) {
			var schema struct {
				Type       string         `json:"type"`
				Properties map[string]any `json:"properties"`
				Additional *bool          `json:"additionalProperties"`
			}
			if err := json.Unmarshal(tool.Def.Parameters, &schema); err != nil {
				t.Fatalf("schema is not valid JSON: %v", err)
			}
			if schema.Type != "object" {
				t.Fatalf(`type = %q, want "object"`, schema.Type)
			}
			if len(schema.Properties) == 0 {
				t.Fatal("schema declares no properties")
			}
			if schema.Additional == nil || *schema.Additional {
				t.Fatal(`schema must set "additionalProperties": false — the decoder refuses unknown fields, and a schema that allows them promises the model something the tool will not do`)
			}
		})
	}
}

func TestUpdateEchoTool_NoRealChangeNeverAsks(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	echoSvc.getByIDFn = func(string) (*echoModel.Echo, error) { return existingEcho(), nil }
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}

	a, events := newTestAsker(time.Second)
	a.registry = s.asks

	// Same content, the same tag in a different spelling, and the visibility it
	// already has: nothing to approve.
	if _, err := runMutation(t, s.updateEchoTool(a, "zh-CN", time.UTC),
		echoArgs(testEchoID, `"content":"今天天气不错","tags":["#日常"],"private":false`)); err == nil {
		t.Fatal("update_echo should refuse a change that changes nothing")
	}
	if len(events) != 0 {
		t.Fatal("a question was shown for a no-op update")
	}
	if len(echoSvc.updated) != 0 {
		t.Fatalf("updated = %+v, want nothing", echoSvc.updated)
	}
}

// Every write tool must arrive at the loop as a complete mutation. This is the
// test that catches a fourth one added later without one.
func TestWriteTools_AreCompleteMutations(t *testing.T) {
	echoSvc := &recordingEchoSvc{}
	s := &CopilotService{echoService: echoSvc, asks: newAskRegistry()}
	a, _ := newTestAsker(time.Second)
	a.registry = s.asks

	tools := []agent.Tool{
		s.createEchoTool(a, "zh-CN", time.UTC),
		s.updateEchoTool(a, "zh-CN", time.UTC),
		s.deleteEchoTool(a, "zh-CN", time.UTC),
	}
	for _, tool := range tools {
		t.Run(tool.Def.Name, func(t *testing.T) {
			if tool.Effect != agent.EffectMutate {
				t.Fatalf("effect = %d, want EffectMutate", tool.Effect)
			}
			if tool.Run != nil {
				t.Fatal("a mutating tool must not carry a read body — the loop would never call it, but its presence hides the mistake")
			}
			if tool.Mutation == nil || tool.Mutation.Plan == nil || tool.Mutation.Confirm == nil {
				t.Fatal("mutation is incomplete")
			}
		})
	}
}

// ask_user is a read and must stay one: if it ever declared EffectMutate the
// loop would put it through a confirmation, and a question would need a question
// to be asked.
func TestAskUserTool_IsAnInteractiveRead(t *testing.T) {
	s := &CopilotService{asks: newAskRegistry()}
	a, _ := newTestAsker(time.Second)
	tool := s.askUserTool(a, "zh-CN")

	if tool.Effect != agent.EffectRead {
		t.Fatalf("effect = %d, want EffectRead", tool.Effect)
	}
	if !tool.Interactive {
		t.Fatal("ask_user must be interactive")
	}
	if tool.Run == nil || tool.Mutation != nil {
		t.Fatal("ask_user must be a plain read body")
	}
}

func TestAnswerText_RendersARound(t *testing.T) {
	strs := askStringsFor("zh-CN")
	questions := []AskQuestion{{ID: "a", Text: "问题一"}, {ID: "b", Text: "问题二"}}
	answers := []AskAnswer{
		{QuestionID: "b", Custom: "打字的回答"},
		{QuestionID: "a", Selected: []string{"选项一", "选项二"}},
	}

	got := answerText(strs, questions, answers)
	for _, want := range []string{"问题一", "选项一、选项二", "问题二", "打字的回答"} {
		if !strings.Contains(got, want) {
			t.Fatalf("answerText = %q, want it to contain %q", got, want)
		}
	}

	single := answerText(strs, questions[:1], answers[1:])
	if single != "选项一、选项二" {
		t.Fatalf("a single question answered as %q, want just the reply", single)
	}
}

func TestReplyOf_UnansweredIsSaidPlainly(t *testing.T) {
	strs := askStringsFor("zh-CN")
	if got := replyOf(strs, nil, "missing"); got != strs.NoAnswer {
		t.Fatalf("replyOf = %q, want %q", got, strs.NoAnswer)
	}
	if got := replyOf(strs, []AskAnswer{{QuestionID: "q"}}, "q"); got != strs.NoAnswer {
		t.Fatalf("an empty reply rendered as %q, want %q", got, strs.NoAnswer)
	}
}
