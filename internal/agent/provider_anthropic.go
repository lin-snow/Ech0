// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

const anthropicDefaultMaxTokens = 4096

type anthropicProvider struct {
	setting model.AgentSetting
}

func (p *anthropicProvider) newClient() anthropic.Client {
	opts := []option.RequestOption{option.WithAPIKey(p.setting.ApiKey)}
	if p.setting.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(p.setting.BaseURL))
	}
	return anthropic.NewClient(opts...)
}

func (p *anthropicProvider) buildParams(req Request) anthropic.MessageNewParams {
	systemBlocks, msgs := p.buildMessages(req.Messages)

	maxTokens := int64(anthropicDefaultMaxTokens)
	if req.MaxTokens > 0 {
		maxTokens = int64(req.MaxTokens)
	}
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.setting.Model),
		MaxTokens: maxTokens,
		Messages:  msgs,
	}
	if len(systemBlocks) > 0 {
		systemBlocks[len(systemBlocks)-1].CacheControl = anthropic.NewCacheControlEphemeralParam()
		params.System = systemBlocks
	}
	if tools := p.buildTools(req.Tools); len(tools) > 0 {
		params.Tools = tools
	}
	if req.Temperature != nil {
		params.Temperature = param.NewOpt(float64(*req.Temperature))
	}
	return params
}

func (p *anthropicProvider) generate(ctx context.Context, req Request) (string, []ToolCall, error) {
	client := p.newClient()
	resp, err := client.Messages.New(ctx, p.buildParams(req))
	if err != nil {
		return "", nil, err
	}

	var text strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	return text.String(), toolCallsFromContent(resp.Content), nil
}

func toolCallsFromContent(content []anthropic.ContentBlockUnion) []ToolCall {
	var calls []ToolCall
	for _, block := range content {
		if block.Type != "tool_use" {
			continue
		}
		args := json.RawMessage(block.Input)
		if len(args) == 0 {
			args = json.RawMessage("{}")
		}
		calls = append(calls, ToolCall{
			ID:   block.ID,
			Name: block.Name,
			Args: args,
		})
	}
	return calls
}

func (p *anthropicProvider) buildMessages(in []Message) ([]anthropic.TextBlockParam, []anthropic.MessageParam) {
	var (
		systemBlocks []anthropic.TextBlockParam
		msgs         []anthropic.MessageParam
		pendingTools []anthropic.ContentBlockParamUnion
	)

	flush := func() {
		if len(pendingTools) > 0 {
			msgs = append(msgs, anthropic.NewUserMessage(pendingTools...))
			pendingTools = nil
		}
	}

	for _, m := range in {
		switch m.Role {
		case RoleSystem:
			flush()
			systemBlocks = append(systemBlocks, anthropic.TextBlockParam{Text: m.Content})
		case RoleAssistant:
			flush()
			var blocks []anthropic.ContentBlockParamUnion
			if m.Content != "" {
				blocks = append(blocks, anthropic.NewTextBlock(m.Content))
			}
			for _, tc := range m.ToolCalls {
				blocks = append(blocks, anthropic.NewToolUseBlock(tc.ID, json.RawMessage(tc.Args), tc.Name))
			}
			if len(blocks) > 0 {
				msgs = append(msgs, anthropic.NewAssistantMessage(blocks...))
			}
		case RoleTool:
			pendingTools = append(pendingTools, anthropic.NewToolResultBlock(m.ToolCallID, m.Content, false))
		default:
			flush()
			msgs = append(msgs, anthropic.NewUserMessage(userBlocks(m)...))
		}
	}
	flush()
	return systemBlocks, msgs
}

func userBlocks(m Message) []anthropic.ContentBlockParamUnion {
	var blocks []anthropic.ContentBlockParamUnion
	if m.Content != "" {
		blocks = append(blocks, anthropic.NewTextBlock(m.Content))
	}
	for _, img := range m.Images {
		switch {
		case img.Base64 != "":
			blocks = append(blocks, anthropic.NewImageBlockBase64(img.MediaType, img.Base64))
		case img.URL != "":
			blocks = append(blocks, anthropic.NewImageBlock(anthropic.URLImageSourceParam{URL: img.URL}))
		}
	}
	if len(blocks) == 0 {
		blocks = append(blocks, anthropic.NewTextBlock(m.Content))
	}
	return blocks
}

func (p *anthropicProvider) buildTools(defs []ToolDef) []anthropic.ToolUnionParam {
	if len(defs) == 0 {
		return nil
	}
	tools := make([]anthropic.ToolUnionParam, 0, len(defs))
	for _, d := range defs {
		var parsed struct {
			Properties any      `json:"properties"`
			Required   []string `json:"required"`
		}
		_ = json.Unmarshal(d.Parameters, &parsed)

		t := anthropic.ToolUnionParamOfTool(
			anthropic.ToolInputSchemaParam{
				Properties: parsed.Properties,
				Required:   parsed.Required,
			},
			d.Name,
		)
		if t.OfTool != nil && d.Description != "" {
			t.OfTool.Description = param.NewOpt(d.Description)
		}
		tools = append(tools, t)
	}
	return tools
}

func (p *anthropicProvider) Complete(ctx context.Context, req Request) (Response, error) {
	text, _, err := p.generate(ctx, req)
	if err != nil {
		return Response{}, err
	}
	if text == "" {
		return Response{}, errors.New("anthropic: empty text response")
	}
	return Response{Text: text}, nil
}

func (p *anthropicProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event)
	go p.stream(ctx, req, ch)
	return ch, nil
}

func (p *anthropicProvider) stream(ctx context.Context, req Request, ch chan<- Event) {
	defer close(ch)

	client := p.newClient()
	stream := client.Messages.NewStreaming(ctx, p.buildParams(req))

	var acc anthropic.Message
	for stream.Next() {
		event := stream.Current()
		if err := acc.Accumulate(event); err != nil {
			send(ctx, ch, Event{Kind: EventError, Err: err})
			return
		}

		if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
			if td, ok := delta.Delta.AsAny().(anthropic.TextDelta); ok && td.Text != "" {
				if !send(ctx, ch, Event{Kind: EventTextDelta, Text: td.Text}) {
					return
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		send(ctx, ch, Event{Kind: EventError, Err: err})
		return
	}
	for _, tc := range toolCallsFromContent(acc.Content) {
		if !send(ctx, ch, Event{Kind: EventToolCall, ToolCall: tc}) {
			return
		}
	}
	send(ctx, ch, Event{Kind: EventDone})
}
