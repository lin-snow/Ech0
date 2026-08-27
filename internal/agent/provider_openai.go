// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	model "github.com/lin-snow/ech0/internal/model/setting"
	openai "github.com/sashabaranov/go-openai"
)

type openaiProvider struct {
	setting model.AgentSetting
}

func (p *openaiProvider) client() *openai.Client {
	cfg := openai.DefaultConfig(p.setting.ApiKey)
	if p.setting.BaseURL != "" {
		cfg.BaseURL = p.setting.BaseURL
	}
	return openai.NewClientWithConfig(cfg)
}

func (p *openaiProvider) buildMessages(in []Message) []openai.ChatCompletionMessage {
	msgs := make([]openai.ChatCompletionMessage, 0, len(in))
	for _, m := range in {
		msg := openai.ChatCompletionMessage{
			Role:       toOpenAIRole(m.Role),
			ToolCallID: m.ToolCallID,
		}
		if len(m.Images) > 0 {
			msg.MultiContent = openAIImageParts(m.Content, m.Images)
		} else {
			msg.Content = m.Content
		}
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openai.ToolCall{
				ID:   tc.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      tc.Name,
					Arguments: string(tc.Args),
				},
			})
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

func openAIImageParts(text string, images []ImagePart) []openai.ChatMessagePart {
	parts := make([]openai.ChatMessagePart, 0, len(images)+1)
	if text != "" {
		parts = append(parts, openai.ChatMessagePart{Type: openai.ChatMessagePartTypeText, Text: text})
	}
	for _, img := range images {
		url := img.URL
		if img.Base64 != "" {
			url = "data:" + img.MediaType + ";base64," + img.Base64
		}
		if url == "" {
			continue
		}
		parts = append(parts, openai.ChatMessagePart{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: url},
		})
	}
	return parts
}

func (p *openaiProvider) buildTools(defs []ToolDef) []openai.Tool {
	if len(defs) == 0 {
		return nil
	}
	tools := make([]openai.Tool, 0, len(defs))
	for _, d := range defs {
		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        d.Name,
				Description: d.Description,
				Parameters:  json.RawMessage(d.Parameters),
			},
		})
	}
	return tools
}

func (p *openaiProvider) Complete(ctx context.Context, req Request) (Response, error) {
	chatReq := openai.ChatCompletionRequest{
		Model:    p.setting.Model,
		Messages: p.buildMessages(req.Messages),
		Tools:    p.buildTools(req.Tools),
	}
	if req.Temperature != nil {
		chatReq.Temperature = *req.Temperature
	}
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}

	resp, err := p.client().CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return Response{}, err
	}
	if len(resp.Choices) == 0 {
		return Response{}, errors.New("openai: empty response")
	}
	return Response{Text: resp.Choices[0].Message.Content}, nil
}

func (p *openaiProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event)
	go p.stream(ctx, req, ch)
	return ch, nil
}

func (p *openaiProvider) stream(ctx context.Context, req Request, ch chan<- Event) {
	defer close(ch)

	chatReq := openai.ChatCompletionRequest{
		Model:    p.setting.Model,
		Messages: p.buildMessages(req.Messages),
		Tools:    p.buildTools(req.Tools),
		Stream:   true,
	}
	if req.Temperature != nil {
		chatReq.Temperature = *req.Temperature
	}
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}

	stream, err := p.client().CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		send(ctx, ch, Event{Kind: EventError, Err: err})
		return
	}
	defer func() { _ = stream.Close() }()

	acc := newToolCallAccumulator()
	guard := &toolCallLeakGuard{}
	splitter := &reasoningSplitter{}

	for {
		resp, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			send(ctx, ch, Event{Kind: EventError, Err: recvErr})
			return
		}
		if len(resp.Choices) == 0 {
			continue
		}
		delta := resp.Choices[0].Delta

		if delta.ReasoningContent != "" {
			if !send(ctx, ch, Event{Kind: EventReasoningDelta, Text: delta.ReasoningContent}) {
				return
			}
		}
		if delta.Content != "" {
			answer, reasoning := splitter.feed(delta.Content)
			if reasoning != "" && !send(ctx, ch, Event{Kind: EventReasoningDelta, Text: reasoning}) {
				return
			}
			if answer != "" {
				safe, tripped := guard.feed(answer)
				if tripped {
					send(ctx, ch, Event{Kind: EventError, Err: errTextToolCallLeak})
					return
				}
				if safe != "" && !send(ctx, ch, Event{Kind: EventTextDelta, Text: safe}) {
					return
				}
			}
		}
		acc.add(delta.ToolCalls)
	}

	ansRest, reaRest := splitter.flush()
	if reaRest != "" && !send(ctx, ch, Event{Kind: EventReasoningDelta, Text: reaRest}) {
		return
	}
	if ansRest != "" {
		safe, tripped := guard.feed(ansRest)
		if tripped {
			send(ctx, ch, Event{Kind: EventError, Err: errTextToolCallLeak})
			return
		}
		if safe != "" && !send(ctx, ch, Event{Kind: EventTextDelta, Text: safe}) {
			return
		}
	}
	if rest := guard.flush(); rest != "" {
		if !send(ctx, ch, Event{Kind: EventTextDelta, Text: rest}) {
			return
		}
	}
	for _, tc := range acc.finish() {
		if !send(ctx, ch, Event{Kind: EventToolCall, ToolCall: tc}) {
			return
		}
	}
	send(ctx, ch, Event{Kind: EventDone})
}

type toolCallAccumulator struct {
	order []int
	byIdx map[int]*ToolCall
	args  map[int][]byte
}

func newToolCallAccumulator() *toolCallAccumulator {
	return &toolCallAccumulator{
		byIdx: make(map[int]*ToolCall),
		args:  make(map[int][]byte),
	}
}

func (a *toolCallAccumulator) add(deltas []openai.ToolCall) {
	for _, d := range deltas {
		idx := 0
		if d.Index != nil {
			idx = *d.Index
		}
		tc, ok := a.byIdx[idx]
		if !ok {
			tc = &ToolCall{}
			a.byIdx[idx] = tc
			a.order = append(a.order, idx)
		}
		if d.ID != "" {
			tc.ID = d.ID
		}
		if d.Function.Name != "" {
			tc.Name = d.Function.Name
		}
		if d.Function.Arguments != "" {
			a.args[idx] = append(a.args[idx], d.Function.Arguments...)
		}
	}
}

func (a *toolCallAccumulator) finish() []ToolCall {
	out := make([]ToolCall, 0, len(a.order))
	for _, idx := range a.order {
		tc := a.byIdx[idx]
		args := a.args[idx]
		if len(args) == 0 {
			args = []byte("{}")
		}
		out = append(out, ToolCall{ID: tc.ID, Name: tc.Name, Args: json.RawMessage(args)})
	}
	return out
}

func toOpenAIRole(r Role) string {
	switch r {
	case RoleSystem:
		return openai.ChatMessageRoleSystem
	case RoleAssistant:
		return openai.ChatMessageRoleAssistant
	case RoleTool:
		return openai.ChatMessageRoleTool
	default:
		return openai.ChatMessageRoleUser
	}
}

func send(ctx context.Context, ch chan<- Event, ev Event) bool {
	select {
	case ch <- ev:
		return true
	case <-ctx.Done():
		return false
	}
}

var errTextToolCallLeak = errors.New(
	"检测到模型以文本形式返回工具调用：端点未归一化为结构化 tool_calls。请在推理服务启用对应的 " +
		"tool-call parser（如 vLLM 的 --enable-auto-tool-choice --tool-call-parser）后重试")

var textToolCallMarkers = []string{"<tool_call>", "<function="}

type toolCallLeakGuard struct {
	pending string
}

func (g *toolCallLeakGuard) feed(text string) (safe string, tripped bool) {
	g.pending += text
	for _, m := range textToolCallMarkers {
		if strings.Contains(g.pending, m) {
			g.pending = ""
			return "", true
		}
	}
	hold := markerPrefixHold(g.pending)
	safe, g.pending = g.pending[:len(g.pending)-hold], g.pending[len(g.pending)-hold:]
	return safe, false
}

func (g *toolCallLeakGuard) flush() string {
	s := g.pending
	g.pending = ""
	return s
}

func markerPrefixHold(s string) int {
	hold := 0
	for _, m := range textToolCallMarkers {
		n := min(len(m), len(s))
		for k := n; k > hold; k-- {
			if strings.HasPrefix(m, s[len(s)-k:]) {
				hold = k
				break
			}
		}
	}
	return hold
}
