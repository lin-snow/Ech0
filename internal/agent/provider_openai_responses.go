// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	model "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

const (
	respEventTextDelta             = "response.output_text.delta"
	respEventReasoningTextDelta    = "response.reasoning_text.delta"
	respEventReasoningSummaryDelta = "response.reasoning_summary_text.delta"
	respEventOutputItemDone        = "response.output_item.done"
	respEventCompleted             = "response.completed"
	respEventFailed                = "response.failed"
	respEventError                 = "error"
)

const respItemTypeFunctionCall = "function_call"

type openaiResponsesProvider struct {
	setting model.AgentSetting
}

func (p *openaiResponsesProvider) service() responses.ResponseService {
	opts := []option.RequestOption{option.WithEnvironmentProduction()}
	if p.setting.ApiKey != "" {
		opts = append(opts, option.WithAPIKey(p.setting.ApiKey))
	}
	if p.setting.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(p.setting.BaseURL))
	}
	return responses.NewResponseService(opts...)
}

func (p *openaiResponsesProvider) buildParams(req Request) (responses.ResponseNewParams, error) {
	tools, err := p.buildTools(req.Tools)
	if err != nil {
		return responses.ResponseNewParams{}, err
	}

	params := responses.ResponseNewParams{
		Model: p.setting.Model,
		Input: responses.ResponseNewParamsInputUnion{OfInputItemList: p.buildInput(req.Messages)},
		Tools: tools,
		Store: param.NewOpt(false),
	}
	if req.Temperature != nil {
		params.Temperature = param.NewOpt(float64(*req.Temperature))
	}
	if req.MaxTokens > 0 {
		params.MaxOutputTokens = param.NewOpt(int64(req.MaxTokens))
	}
	return params, nil
}

func (p *openaiResponsesProvider) buildInput(in []Message) responses.ResponseInputParam {
	items := make(responses.ResponseInputParam, 0, len(in))
	for _, m := range in {
		switch m.Role {
		case RoleSystem:
			items = append(items, responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleSystem))
		case RoleAssistant:
			if m.Content != "" {
				items = append(
					items,
					responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleAssistant),
				)
			}
			for _, tc := range m.ToolCalls {
				items = append(
					items,
					responses.ResponseInputItemParamOfFunctionCall(respToolArgs(string(tc.Args)), tc.ID, tc.Name),
				)
			}
		case RoleTool:
			items = append(items, responses.ResponseInputItemParamOfFunctionCallOutput(m.ToolCallID, m.Content))
		default:
			if len(m.Images) > 0 {
				items = append(items, responses.ResponseInputItemParamOfMessage(
					respImageParts(m.Content, m.Images),
					responses.EasyInputMessageRoleUser,
				))
				continue
			}
			items = append(items, responses.ResponseInputItemParamOfMessage(m.Content, responses.EasyInputMessageRoleUser))
		}
	}
	return items
}

func respImageParts(text string, images []ImagePart) responses.ResponseInputMessageContentListParam {
	parts := make(responses.ResponseInputMessageContentListParam, 0, len(images)+1)
	if text != "" {
		parts = append(parts, responses.ResponseInputContentParamOfInputText(text))
	}
	for _, img := range images {
		url := img.URL
		if img.Base64 != "" {
			url = "data:" + img.MediaType + ";base64," + img.Base64
		}
		if url == "" {
			continue
		}
		part := responses.ResponseInputContentParamOfInputImage(responses.ResponseInputImageDetailAuto)
		part.OfInputImage.ImageURL = param.NewOpt(url)
		parts = append(parts, part)
	}
	return parts
}

func (p *openaiResponsesProvider) buildTools(defs []ToolDef) ([]responses.ToolUnionParam, error) {
	if len(defs) == 0 {
		return nil, nil
	}
	tools := make([]responses.ToolUnionParam, 0, len(defs))
	for _, d := range defs {
		var schema map[string]any
		if err := json.Unmarshal(d.Parameters, &schema); err != nil {
			return nil, fmt.Errorf("openai responses: 工具 %s 的 JSON Schema 无法解析: %w", d.Name, err)
		}
		tools = append(tools, responses.ToolUnionParam{OfFunction: &responses.FunctionToolParam{
			Name:        d.Name,
			Description: param.NewOpt(d.Description),
			Parameters:  schema,
			Strict:      param.NewOpt(false),
		}})
	}
	return tools, nil
}

func (p *openaiResponsesProvider) Complete(ctx context.Context, req Request) (Response, error) {
	params, err := p.buildParams(req)
	if err != nil {
		return Response{}, err
	}

	svc := p.service()
	resp, err := svc.New(ctx, params)
	if err != nil {
		return Response{}, err
	}
	if resp.Status == responses.ResponseStatusFailed {
		return Response{}, fmt.Errorf("openai responses: %s", respFailureMessage(resp.Error))
	}
	return Response{Text: resp.OutputText()}, nil
}

func (p *openaiResponsesProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event)
	go p.stream(ctx, req, ch)
	return ch, nil
}

func (p *openaiResponsesProvider) stream(ctx context.Context, req Request, ch chan<- Event) {
	defer close(ch)

	params, err := p.buildParams(req)
	if err != nil {
		send(ctx, ch, Event{Kind: EventError, Err: err})
		return
	}

	svc := p.service()
	stream := svc.NewStreaming(ctx, params)
	defer func() { _ = stream.Close() }()

	pipe := &textPipeline{}
	calls := &respToolCalls{}

	for stream.Next() {
		ev := stream.Current()
		switch ev.Type {
		case respEventTextDelta:
			if !pipe.feed(ctx, ch, ev.Delta) {
				return
			}
		case respEventReasoningTextDelta, respEventReasoningSummaryDelta:
			if ev.Delta != "" && !send(ctx, ch, Event{Kind: EventReasoningDelta, Text: ev.Delta}) {
				return
			}
		case respEventOutputItemDone:
			calls.addItem(ev.Item)
		case respEventCompleted:
			calls.addFinal(ev.Response.Output)
		case respEventFailed:
			send(ctx, ch, Event{Kind: EventError, Err: fmt.Errorf(
				"openai responses: %s", respFailureMessage(ev.Response.Error),
			)})
			return
		case respEventError:
			send(ctx, ch, Event{Kind: EventError, Err: fmt.Errorf(
				"openai responses: %s", respStreamErrorMessage(ev.Code, ev.Message),
			)})
			return
		}
	}
	if err := stream.Err(); err != nil {
		send(ctx, ch, Event{Kind: EventError, Err: err})
		return
	}

	if !pipe.flush(ctx, ch) {
		return
	}
	for _, tc := range calls.finish() {
		if !send(ctx, ch, Event{Kind: EventToolCall, ToolCall: tc}) {
			return
		}
	}
	send(ctx, ch, Event{Kind: EventDone})
}

type respToolCalls struct {
	fromItems []ToolCall
	fromFinal []ToolCall
}

func (c *respToolCalls) addItem(item responses.ResponseOutputItemUnion) {
	if tc, ok := respToolCallFromItem(item); ok {
		c.fromItems = append(c.fromItems, tc)
	}
}

func (c *respToolCalls) addFinal(output []responses.ResponseOutputItemUnion) {
	for _, item := range output {
		if tc, ok := respToolCallFromItem(item); ok {
			c.fromFinal = append(c.fromFinal, tc)
		}
	}
}

func (c *respToolCalls) finish() []ToolCall {
	if len(c.fromItems) > 0 {
		return c.fromItems
	}
	return c.fromFinal
}

func respToolCallFromItem(item responses.ResponseOutputItemUnion) (ToolCall, bool) {
	if item.Type != respItemTypeFunctionCall {
		return ToolCall{}, false
	}
	call := item.AsFunctionCall()
	return ToolCall{
		ID:   call.CallID,
		Name: call.Name,
		Args: json.RawMessage(respToolArgs(call.Arguments)),
	}, true
}

func respToolArgs(args string) string {
	if strings.TrimSpace(args) == "" {
		return "{}"
	}
	return args
}

func respFailureMessage(respErr responses.ResponseError) string {
	return respStreamErrorMessage(string(respErr.Code), respErr.Message)
}

func respStreamErrorMessage(code, message string) string {
	switch {
	case message != "" && code != "":
		return message + "（" + code + "）"
	case message != "":
		return message
	case code != "":
		return code
	default:
		return "生成失败且未返回错误详情"
	}
}
