// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	logUtil "github.com/lin-snow/ech0/pkg/log"
	"golang.org/x/sync/errgroup"
)

const defaultMaxRounds = 3

const maxParallelTools = 4

var defaultRunStrings = RunStrings{
	DedupNote:       "（已检索过，结果见上）",
	UnknownTool:     "未知工具：",
	ToolError:       "工具执行失败：",
	ImageNote:       toolImageNote,
	ContextTrimNote: "（早前检索结果已省略以控制长度）",
	Malformed:       "该工具未被正确声明，已拒绝执行：",
}

func (s RunStrings) withDefaults() RunStrings {
	if s.DedupNote == "" {
		s.DedupNote = defaultRunStrings.DedupNote
	}
	if s.UnknownTool == "" {
		s.UnknownTool = defaultRunStrings.UnknownTool
	}
	if s.ToolError == "" {
		s.ToolError = defaultRunStrings.ToolError
	}
	if s.ImageNote == "" {
		s.ImageNote = defaultRunStrings.ImageNote
	}
	if s.ContextTrimNote == "" {
		s.ContextTrimNote = defaultRunStrings.ContextTrimNote
	}
	if s.Malformed == "" {
		s.Malformed = defaultRunStrings.Malformed
	}
	return s
}

// Run starts one agent run and returns the events it produces. The channel
// closes when the run is over.
func Run(ctx context.Context, req RunRequest) (<-chan AgentEvent, error) {
	if err := validate(req.Setting); err != nil {
		return nil, err
	}
	provider, err := providerFor(req.Setting)
	if err != nil {
		return nil, err
	}

	out := make(chan AgentEvent)
	go runLoop(newBudget(ctx, req.Timeout), provider, req, out)
	return out, nil
}

// budget is the run's generation deadline, and the one thing it does that a
// context deadline cannot is stand still.
//
// Time a person spends deciding is not time the run is spending. An interactive
// tool holds its call open for as long as someone takes to answer, and charging
// that to the clock that bounds a hung provider would kill every confirmation
// nobody clicked within the timeout — the mechanism would work only for people
// who happened to be watching. So the deadline lives here, is credited back
// whatever an interactive call waited, and the contexts derived from it are
// per-step rather than one for the whole run.
//
// base is the caller's own context, and it is what every channel send is gated
// on: it dies when the client does, which is the one cancellation a partially
// finished run must still respect.
type budget struct {
	base     context.Context
	deadline time.Time
	unbound  bool
}

func newBudget(ctx context.Context, timeout time.Duration) *budget {
	if timeout <= 0 {
		return &budget{base: ctx, unbound: true}
	}
	return &budget{base: ctx, deadline: time.Now().Add(timeout)}
}

// step derives the context for one piece of work the deadline applies to.
func (b *budget) step() (context.Context, context.CancelFunc) {
	if b.unbound {
		return context.WithCancel(b.base)
	}
	return context.WithDeadline(b.base, b.deadline)
}

// credit hands back the time a person took. Called only for interactive calls,
// so a tool that merely runs slowly cannot buy itself more room.
func (b *budget) credit(waited time.Duration) {
	if !b.unbound {
		b.deadline = b.deadline.Add(waited)
	}
}

func runLoop(b *budget, provider Provider, req RunRequest, out chan<- AgentEvent) {
	defer close(out)

	maxRounds := req.MaxRounds
	if maxRounds <= 0 {
		maxRounds = defaultMaxRounds
	}

	toolDefs := make([]ToolDef, 0, len(req.Tools))
	toolByName := make(map[string]Tool, len(req.Tools))
	for _, t := range req.Tools {
		toolDefs = append(toolDefs, t.Def)
		toolByName[t.Def.Name] = t
	}

	messages := req.Messages
	seen := make(map[string]bool)
	strs := req.Strings.withDefaults()

	for round := 0; round < maxRounds; round++ {
		trimContext(messages, req.MaxContextTokens, strs.ContextTrimNote)
		o := streamRound(b, provider, out, messages, toolDefs, req.Temp)
		if o.aborted {
			return
		}
		if o.err != nil {
			emit(b.base, out, AgentEvent{Kind: AgentError, Err: o.err})
			return
		}
		if len(o.calls) == 0 {
			emit(b.base, out, AgentEvent{Kind: AgentDone})
			return
		}

		messages = append(messages, Message{Role: RoleAssistant, Content: o.assistant, ToolCalls: o.calls})
		if !execTools(b, out, o.calls, toolByName, seen, &messages, strs) {
			return
		}
	}

	trimContext(messages, req.MaxContextTokens, strs.ContextTrimNote)
	o := streamRound(b, provider, out, messages, nil, req.Temp)
	if o.aborted {
		return
	}
	if o.err != nil {
		emit(b.base, out, AgentEvent{Kind: AgentError, Err: o.err})
		return
	}
	emit(b.base, out, AgentEvent{Kind: AgentDone})
}

type roundOutcome struct {
	calls     []ToolCall
	assistant string
	aborted   bool
	err       error
}

func streamRound(
	b *budget,
	provider Provider,
	out chan<- AgentEvent,
	messages []Message,
	toolDefs []ToolDef,
	temp *float32,
) roundOutcome {
	ctx, cancel := b.step()
	defer cancel()

	evCh, err := provider.Stream(ctx, Request{
		Messages:    messages,
		Tools:       toolDefs,
		Temperature: temp,
	})
	if err != nil {
		return roundOutcome{err: err}
	}

	var (
		o    roundOutcome
		text strings.Builder
	)
	for ev := range evCh {
		switch ev.Kind {
		case EventTextDelta:
			text.WriteString(ev.Text)
			if !emit(b.base, out, AgentEvent{Kind: AgentDelta, Text: ev.Text}) {
				o.aborted = true
			}
		case EventReasoningDelta:
			if !emit(b.base, out, AgentEvent{Kind: AgentReasoning, Text: ev.Text}) {
				o.aborted = true
			}
		case EventToolCall:
			o.calls = append(o.calls, ev.ToolCall)
		case EventError:
			o.err = ev.Err
		case EventDone:
		}
		if o.aborted {
			o.assistant = text.String()
			return o
		}
	}
	o.assistant = text.String()
	return o
}

// execTools runs one round's tool calls and appends their results to messages.
//
// The calls are split rather than merged into one group: machine work goes into
// the parallel group as before, and interactive work runs afterwards, one call
// at a time. Two questions racing the same event stream would reach the client
// interleaved with no way to tell which picker a click belongs to, and the
// second one would be answered by the first one's reply.
func execTools(
	b *budget,
	out chan<- AgentEvent,
	calls []ToolCall,
	toolByName map[string]Tool,
	seen map[string]bool,
	messages *[]Message,
	strs RunStrings,
) bool {
	n := len(calls)
	toolMsgs := make([]Message, n)
	imageMsgs := make([]*Message, n)
	outputs := make([]ToolOutput, n)
	execErrs := make([]error, n)

	var runnable, interactive []int
	for i, tc := range calls {
		key := tc.Name + ":" + string(tc.Args)
		if seen[key] {
			toolMsgs[i] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.DedupNote}
			continue
		}
		tool, ok := toolByName[tc.Name]
		if !ok {
			toolMsgs[i] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.UnknownTool + tc.Name}
			continue
		}
		if tool.blocksOnPerson() {
			interactive = append(interactive, i)
			continue
		}
		seen[key] = true
		runnable = append(runnable, i)
	}

	if len(runnable) > 0 {
		ctx, cancel := b.step()
		var g errgroup.Group
		g.SetLimit(maxParallelTools)
		for _, idx := range runnable {
			idx, tc, tool := idx, calls[idx], toolByName[calls[idx].Name]
			g.Go(func() error {
				if !emit(b.base, out, AgentEvent{Kind: AgentSearching, ToolName: tc.Name, ToolArgs: tc.Args}) {
					return b.base.Err()
				}
				outputs[idx], execErrs[idx] = dispatchTool(ctx, tool, tc.Args, strs)
				return nil
			})
		}
		err := g.Wait()
		cancel()
		if err != nil {
			return false
		}
	}

	for _, idx := range interactive {
		tc, tool := calls[idx], toolByName[calls[idx].Name]
		started := time.Now()
		outputs[idx], execErrs[idx] = dispatchTool(b.base, tool, tc.Args, strs)
		b.credit(time.Since(started))
	}

	for _, idx := range append(runnable, interactive...) {
		tc := calls[idx]
		if execErrs[idx] != nil {
			logUtil.GetLogger().Warn("agent tool execute failed",
				slog.String("module", "agent"),
				slog.String("tool", tc.Name),
				logUtil.Err(execErrs[idx]))
			toolMsgs[idx] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.ToolError + execErrs[idx].Error()}
			continue
		}
		if !emit(b.base, out, AgentEvent{Kind: AgentToolResult, ToolName: tc.Name, Meta: outputs[idx].Meta}) {
			return false
		}
		toolMsgs[idx] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: outputs[idx].Content}
		if len(outputs[idx].Images) > 0 {
			imageMsgs[idx] = &Message{Role: RoleUser, Content: strs.ImageNote, Images: outputs[idx].Images}
		}
	}

	*messages = append(*messages, toolMsgs...)
	for i := range imageMsgs {
		if imageMsgs[i] != nil {
			*messages = append(*messages, *imageMsgs[i])
		}
	}
	return true
}

// blocksOnPerson reports whether this call waits on a human. Derived rather
// than declared for mutations: a write tool whose author forgot the flag would
// otherwise run inside the parallel group, and its confirmation would race
// whatever else that round asked.
func (t Tool) blocksOnPerson() bool {
	return t.Interactive || t.Effect == EffectMutate
}

// dispatchTool is the gate. Every tool call in the loop goes through it, and
// what it enforces is an ordering no tool implementation can opt out of:
// a change is planned, shown, and only then made.
//
// The gate is here rather than in the prompt because a prompt is not a gate.
// An instruction to confirm before writing is remembered until the context
// grows, the conversation turns, or the model is talked out of it — and the one
// time it is forgotten is a write nobody agreed to. Nothing the model can emit
// reaches Apply except through Confirm, so forgetting is not among the things
// that can go wrong.
//
// Every branch that cannot establish that ordering refuses. That includes the
// ones that mean this repository has a bug — an undeclared effect, a mutation
// missing Plan, Confirm or Apply — because a tool nobody finished declaring is
// exactly the tool that must not run.
func dispatchTool(ctx context.Context, tool Tool, args json.RawMessage, strs RunStrings) (ToolOutput, error) {
	switch tool.Effect {
	case EffectRead:
		if tool.Run == nil {
			return malformedTool(tool, strs)
		}
		return tool.Run(ctx, args)
	case EffectMutate:
		return applyMutation(ctx, tool, args, strs)
	default:
		return malformedTool(tool, strs)
	}
}

func applyMutation(ctx context.Context, tool Tool, args json.RawMessage, strs RunStrings) (ToolOutput, error) {
	m := tool.Mutation
	if m == nil || m.Plan == nil || m.Confirm == nil {
		return malformedTool(tool, strs)
	}

	plan, err := m.Plan(ctx, args)
	if err != nil {
		return ToolOutput{}, err
	}
	if plan.Apply == nil {
		return malformedTool(tool, strs)
	}

	decision, err := m.Confirm(ctx, plan)
	if err != nil {
		return ToolOutput{}, err
	}
	if !decision.Approved {
		return ToolOutput{Content: decision.Refusal}, nil
	}

	return plan.Apply(ctx)
}

// malformedTool refuses a call the loop cannot run safely and says so twice: to
// the operator at error level, because it is a bug here rather than anything the
// model did, and to the model as an ordinary refusal, because the turn still has
// to end in a sentence somebody wrote.
func malformedTool(tool Tool, strs RunStrings) (ToolOutput, error) {
	logUtil.GetLogger().Error("agent tool is not declared correctly",
		slog.String("module", "agent"),
		slog.String("tool", tool.Def.Name),
		slog.Int("effect", int(tool.Effect)))
	return ToolOutput{Content: strs.Malformed + tool.Def.Name}, nil
}

func trimContext(messages []Message, limit int, note string) {
	if limit <= 0 {
		return
	}
	for contextTokens(messages) > limit {
		idx := -1
		for i := range messages {
			if messages[i].Role == RoleTool && messages[i].Content != note {
				idx = i
				break
			}
		}
		if idx < 0 {
			return
		}
		messages[idx].Content = note
	}
}

func contextTokens(messages []Message) int {
	total := 0
	for i := range messages {
		total += utf8.RuneCountInString(messages[i].Content)
	}
	return total
}

const toolImageNote = "（以下是上一步检索命中的 Echo 的配图，供你结合图片内容作答）"

func emit(ctx context.Context, out chan<- AgentEvent, ev AgentEvent) bool {
	select {
	case out <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}
