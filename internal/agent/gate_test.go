// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

// The tests in this file pin the mutation gate, which is the one part of the
// loop whose whole value is what it refuses. Every case here is a way a write
// could reach the database without a person having agreed to it, and each must
// stay impossible for reasons the type system or this loop enforces — never for
// reasons a prompt asks the model to remember.

func mutatingProbe(name string, m *Mutation) Tool {
	return Tool{
		Def:      ToolDef{Name: name, Description: "probe", Parameters: json.RawMessage(`{"type":"object"}`)},
		Effect:   EffectMutate,
		Mutation: m,
	}
}

func oneCallThenAnswer(name string) *fakeProvider {
	return &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", name, `{}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}
}

// toolContentOf returns what the model was told for the tool call in the given
// round, which is the only channel a refusal travels on.
func toolContentOf(t *testing.T, fp *fakeProvider, round int) string {
	t.Helper()
	if len(fp.gotReqs) <= round {
		t.Fatalf("provider saw %d rounds, want more than %d", len(fp.gotReqs), round)
	}
	for _, m := range fp.gotReqs[round].Messages {
		if m.Role == RoleTool {
			return m.Content
		}
	}
	t.Fatalf("round %d carried no tool message", round)
	return ""
}

func TestGate_DeclinedDecisionNeverReachesApply(t *testing.T) {
	applied := false
	tool := mutatingProbe("delete_echo", &Mutation{
		Plan: func(_ context.Context, _ json.RawMessage) (Plan, error) {
			return Plan{Prompt: "delete it?", Apply: func(_ context.Context) (ToolOutput, error) {
				applied = true
				return ToolOutput{Content: "deleted"}, nil
			}}, nil
		},
		Confirm: func(_ context.Context, _ Plan) (Decision, error) {
			return Decision{Approved: false, Refusal: "user said no"}, nil
		},
	})

	fp := oneCallThenAnswer("delete_echo")
	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if applied {
		t.Fatal("Apply ran despite a declined decision")
	}
	if got := toolContentOf(t, fp, 1); got != "user said no" {
		t.Fatalf("model was told %q, want the refusal", got)
	}
}

// A zero Decision is what a Confirm implementation returns when it forgets to
// fill anything in. It must read as a no.
func TestGate_ZeroDecisionIsRefusal(t *testing.T) {
	applied := false
	tool := mutatingProbe("update_echo", &Mutation{
		Plan: func(_ context.Context, _ json.RawMessage) (Plan, error) {
			return Plan{Apply: func(_ context.Context) (ToolOutput, error) {
				applied = true
				return ToolOutput{}, nil
			}}, nil
		},
		Confirm: func(_ context.Context, _ Plan) (Decision, error) { return Decision{}, nil },
	})

	runLoopSync(context.Background(), oneCallThenAnswer("update_echo"),
		RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if applied {
		t.Fatal("Apply ran on a zero-value Decision")
	}
}

// A failure to ask is not consent. This is the timeout path: nobody answered,
// Confirm reports why, and the write must not happen anyway.
func TestGate_ConfirmErrorIsNotConsent(t *testing.T) {
	applied := false
	tool := mutatingProbe("create_echo", &Mutation{
		Plan: func(_ context.Context, _ json.RawMessage) (Plan, error) {
			return Plan{Apply: func(_ context.Context) (ToolOutput, error) {
				applied = true
				return ToolOutput{}, nil
			}}, nil
		},
		Confirm: func(_ context.Context, _ Plan) (Decision, error) {
			return Decision{Approved: true}, errors.New("nobody answered in time")
		},
	})

	fp := oneCallThenAnswer("create_echo")
	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if applied {
		t.Fatal("Apply ran even though Confirm failed")
	}
	if got := toolContentOf(t, fp, 1); got == "" {
		t.Fatal("the model was told nothing about the failed confirmation")
	}
}

// The reason EffectUnset is the zero value: a write tool whose author forgot to
// declare anything must not execute. This is the case a prompt cannot cover,
// because there is no prompt involved.
func TestGate_UndeclaredEffectRefusesToRun(t *testing.T) {
	ran := false
	tool := Tool{
		Def: ToolDef{Name: "sloppy_tool", Description: "forgot its effect", Parameters: json.RawMessage(`{"type":"object"}`)},
		Run: func(_ context.Context, _ json.RawMessage) (ToolOutput, error) {
			ran = true
			return ToolOutput{Content: "changed it"}, nil
		},
	}

	fp := oneCallThenAnswer("sloppy_tool")
	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if ran {
		t.Fatal("a tool with no declared effect executed")
	}
	if got := toolContentOf(t, fp, 1); got != defaultRunStrings.Malformed+"sloppy_tool" {
		t.Fatalf("model was told %q, want the malformed refusal", got)
	}
}

// An effect value nobody taught the switch about lands on the strict side too.
func TestGate_UnknownEffectRefusesToRun(t *testing.T) {
	ran := false
	tool := Tool{
		Def:    ToolDef{Name: "future_tool", Description: "from a later edit", Parameters: json.RawMessage(`{"type":"object"}`)},
		Effect: ToolEffect(99),
		Run: func(_ context.Context, _ json.RawMessage) (ToolOutput, error) {
			ran = true
			return ToolOutput{}, nil
		},
	}

	runLoopSync(context.Background(), oneCallThenAnswer("future_tool"),
		RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if ran {
		t.Fatal("a tool with an unknown effect executed")
	}
}

// A mutating tool that never got its Mutation filled in — or got half of one —
// is a bug here, and the safe reading of a bug in the gate is to refuse.
func TestGate_IncompleteMutationRefusesToRun(t *testing.T) {
	applied := func(flag *bool) applyProbe {
		return applyProbe{flag: flag}
	}

	cases := []struct {
		name     string
		mutation func(*bool) *Mutation
	}{
		{"no mutation at all", func(*bool) *Mutation { return nil }},
		{"no plan", func(flag *bool) *Mutation {
			return &Mutation{Confirm: approve}
		}},
		{"no confirm", func(flag *bool) *Mutation {
			return &Mutation{Plan: applied(flag).plan}
		}},
		{"plan without apply", func(*bool) *Mutation {
			return &Mutation{
				Plan:    func(_ context.Context, _ json.RawMessage) (Plan, error) { return Plan{}, nil },
				Confirm: approve,
			}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ran := false
			tool := mutatingProbe("half_declared", tc.mutation(&ran))

			fp := oneCallThenAnswer("half_declared")
			runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

			if ran {
				t.Fatal("Apply ran through an incomplete Mutation")
			}
			if got := toolContentOf(t, fp, 1); got != defaultRunStrings.Malformed+"half_declared" {
				t.Fatalf("model was told %q, want the malformed refusal", got)
			}
		})
	}
}

type applyProbe struct{ flag *bool }

func (p applyProbe) plan(_ context.Context, _ json.RawMessage) (Plan, error) {
	return Plan{Apply: func(_ context.Context) (ToolOutput, error) {
		*p.flag = true
		return ToolOutput{}, nil
	}}, nil
}

func approve(_ context.Context, _ Plan) (Decision, error) { return Decision{Approved: true}, nil }

// The approved path, so the refusals above are not merely a gate that refuses
// everything.
func TestGate_ApprovedDecisionReachesApply(t *testing.T) {
	applied := false
	tool := mutatingProbe("create_echo", &Mutation{
		Plan: func(_ context.Context, _ json.RawMessage) (Plan, error) {
			return Plan{Apply: func(_ context.Context) (ToolOutput, error) {
				applied = true
				return ToolOutput{Content: "posted"}, nil
			}}, nil
		},
		Confirm: approve,
	})

	fp := oneCallThenAnswer("create_echo")
	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if !applied {
		t.Fatal("Apply did not run for an approved decision")
	}
	if got := toolContentOf(t, fp, 1); got != "posted" {
		t.Fatalf("model was told %q, want the applied result", got)
	}
}

// Ordering: Plan runs before Confirm, Confirm before Apply, and each exactly
// once. Plan running after the question would mean the person approved a
// description of something else.
func TestGate_PlanThenConfirmThenApply(t *testing.T) {
	var order []string
	tool := mutatingProbe("update_echo", &Mutation{
		Plan: func(_ context.Context, _ json.RawMessage) (Plan, error) {
			order = append(order, "plan")
			return Plan{Apply: func(_ context.Context) (ToolOutput, error) {
				order = append(order, "apply")
				return ToolOutput{Content: "updated"}, nil
			}}, nil
		},
		Confirm: func(_ context.Context, _ Plan) (Decision, error) {
			order = append(order, "confirm")
			return Decision{Approved: true}, nil
		},
	})

	runLoopSync(context.Background(), oneCallThenAnswer("update_echo"),
		RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	want := []string{"plan", "confirm", "apply"}
	if len(order) != len(want) {
		t.Fatalf("call order = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("call order = %v, want %v", order, want)
		}
	}
}

// A mutation is interactive whether or not anyone said so, which is what keeps
// two confirmations in one round from racing the same stream — and what keeps
// the "running" chip off a question the reader is supposed to be looking at.
func TestGate_MutationsRunSeriallyAndEmitNoSearchingEvent(t *testing.T) {
	live := 0
	peak := 0
	confirm := func(_ context.Context, _ Plan) (Decision, error) {
		live++
		if live > peak {
			peak = live
		}
		time.Sleep(10 * time.Millisecond)
		live--
		return Decision{Approved: true}, nil
	}
	newTool := func(name string) Tool {
		return mutatingProbe(name, &Mutation{
			Plan: func(_ context.Context, _ json.RawMessage) (Plan, error) {
				return Plan{Apply: func(_ context.Context) (ToolOutput, error) {
					return ToolOutput{Content: "done " + name}, nil
				}}, nil
			},
			Confirm: confirm,
		})
	}

	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "delete_a", `{}`), toolCallEvent("c2", "delete_b", `{}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	evs := runLoopSync(context.Background(), fp, RunRequest{
		Setting: enabledSetting(),
		Tools:   []Tool{newTool("delete_a"), newTool("delete_b")},
	})

	if peak != 1 {
		t.Fatalf("peak concurrent confirmations = %d, want 1", peak)
	}
	if n := countKind(evs, AgentSearching); n != 0 {
		t.Fatalf("AgentSearching count = %d, want 0 for interactive calls", n)
	}
	if n := countKind(evs, AgentToolResult); n != 2 {
		t.Fatalf("AgentToolResult count = %d, want 2", n)
	}
}

// The same arguments twice are two decisions, not one cached answer. Reusing
// the dedup note here would hand the model consent nobody gave the second time.
func TestGate_RepeatedMutationIsAskedAgain(t *testing.T) {
	confirms := 0
	applies := 0
	tool := mutatingProbe("delete_echo", &Mutation{
		Plan: func(_ context.Context, _ json.RawMessage) (Plan, error) {
			return Plan{Apply: func(_ context.Context) (ToolOutput, error) {
				applies++
				return ToolOutput{Content: "deleted"}, nil
			}}, nil
		},
		Confirm: func(_ context.Context, _ Plan) (Decision, error) {
			confirms++
			return Decision{Approved: true}, nil
		},
	})

	fp := &fakeProvider{scripts: [][]Event{
		{toolCallEvent("c1", "delete_echo", `{"id":"same"}`), doneEvent()},
		{toolCallEvent("c2", "delete_echo", `{"id":"same"}`), doneEvent()},
		{textEvent("answer"), doneEvent()},
	}}

	runLoopSync(context.Background(), fp, RunRequest{
		Setting:   enabledSetting(),
		Tools:     []Tool{tool},
		MaxRounds: 3,
	})

	if confirms != 2 {
		t.Fatalf("confirmations = %d, want 2 — a repeated write must be asked again", confirms)
	}
	if applies != 2 {
		t.Fatalf("applies = %d, want 2", applies)
	}
}

// A read whose body is missing is the same class of bug as a half-declared
// mutation, and gets the same refusal rather than a nil call.
func TestGate_ReadWithoutBodyRefusesToRun(t *testing.T) {
	tool := Tool{
		Def:    ToolDef{Name: "empty_read", Description: "no body", Parameters: json.RawMessage(`{"type":"object"}`)},
		Effect: EffectRead,
	}

	fp := oneCallThenAnswer("empty_read")
	runLoopSync(context.Background(), fp, RunRequest{Setting: enabledSetting(), Tools: []Tool{tool}})

	if got := toolContentOf(t, fp, 1); got != defaultRunStrings.Malformed+"empty_read" {
		t.Fatalf("model was told %q, want the malformed refusal", got)
	}
}
