// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"log/slog"
	"strings"
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
	return s
}

func Run(ctx context.Context, req RunRequest) (<-chan AgentEvent, error) {
	if err := validate(req.Setting); err != nil {
		return nil, err
	}
	provider, err := providerFor(req.Setting)
	if err != nil {
		return nil, err
	}

	out := make(chan AgentEvent)
	go func() {
		runCtx := ctx
		if req.Timeout > 0 {
			var cancel context.CancelFunc
			runCtx, cancel = context.WithTimeout(ctx, req.Timeout)
			defer cancel()
		}
		runLoop(runCtx, provider, req, out)
	}()
	return out, nil
}

func runLoop(ctx context.Context, provider Provider, req RunRequest, out chan<- AgentEvent) {
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
		o := streamRound(ctx, provider, out, messages, toolDefs, req.Temp)
		if o.aborted {
			return
		}
		if o.err != nil {
			emit(ctx, out, AgentEvent{Kind: AgentError, Err: o.err})
			return
		}
		if len(o.calls) == 0 {
			emit(ctx, out, AgentEvent{Kind: AgentDone})
			return
		}

		messages = append(messages, Message{Role: RoleAssistant, Content: o.assistant, ToolCalls: o.calls})
		if !execTools(ctx, out, o.calls, toolByName, seen, &messages, strs) {
			return
		}
	}

	trimContext(messages, req.MaxContextTokens, strs.ContextTrimNote)
	o := streamRound(ctx, provider, out, messages, nil, req.Temp)
	if o.aborted {
		return
	}
	if o.err != nil {
		emit(ctx, out, AgentEvent{Kind: AgentError, Err: o.err})
		return
	}
	emit(ctx, out, AgentEvent{Kind: AgentDone})
}

type roundOutcome struct {
	calls     []ToolCall
	assistant string
	aborted   bool
	err       error
}

func streamRound(
	ctx context.Context,
	provider Provider,
	out chan<- AgentEvent,
	messages []Message,
	toolDefs []ToolDef,
	temp *float32,
) roundOutcome {
	evCh, err := provider.Stream(ctx, Request{
		Messages:    messages,
		Tools:       toolDefs,
		Temperature: temp,
	})
	if err != nil {
		return roundOutcome{err: err}
	}

	var (
		o roundOutcome
		b strings.Builder
	)
	for ev := range evCh {
		switch ev.Kind {
		case EventTextDelta:
			b.WriteString(ev.Text)
			if !emit(ctx, out, AgentEvent{Kind: AgentDelta, Text: ev.Text}) {
				o.aborted = true
			}
		case EventReasoningDelta:
			if !emit(ctx, out, AgentEvent{Kind: AgentReasoning, Text: ev.Text}) {
				o.aborted = true
			}
		case EventToolCall:
			o.calls = append(o.calls, ev.ToolCall)
		case EventError:
			o.err = ev.Err
		case EventDone:
		}
		if o.aborted {
			o.assistant = b.String()
			return o
		}
	}
	o.assistant = b.String()
	return o
}

func execTools(
	ctx context.Context,
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

	var runnable []int
	for i, tc := range calls {
		key := tc.Name + ":" + string(tc.Args)
		if seen[key] {
			toolMsgs[i] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.DedupNote}
			continue
		}
		seen[key] = true
		if _, ok := toolByName[tc.Name]; !ok {
			toolMsgs[i] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.UnknownTool + tc.Name}
			continue
		}
		runnable = append(runnable, i)
	}

	var g errgroup.Group
	g.SetLimit(maxParallelTools)
	for _, idx := range runnable {
		idx, tc, tool := idx, calls[idx], toolByName[calls[idx].Name]
		g.Go(func() error {
			if !emit(ctx, out, AgentEvent{Kind: AgentSearching, ToolName: tc.Name, ToolArgs: tc.Args}) {
				return ctx.Err()
			}
			outputs[idx], execErrs[idx] = tool.Execute(ctx, tc.Args)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return false
	}

	for _, idx := range runnable {
		tc := calls[idx]
		if execErrs[idx] != nil {
			logUtil.GetLogger().Warn("agent tool execute failed",
				slog.String("module", "agent"),
				slog.String("tool", tc.Name),
				logUtil.Err(execErrs[idx]))
			toolMsgs[idx] = Message{Role: RoleTool, ToolCallID: tc.ID, Content: strs.ToolError + execErrs[idx].Error()}
			continue
		}
		if !emit(ctx, out, AgentEvent{Kind: AgentToolResult, ToolName: tc.Name, Meta: outputs[idx].Meta}) {
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

func trimContext(messages []Message, budget int, note string) {
	if budget <= 0 {
		return
	}
	for contextTokens(messages) > budget {
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
