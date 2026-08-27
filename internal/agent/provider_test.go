// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"testing"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	openai "github.com/sashabaranov/go-openai"
)

func TestToolCallAccumulator_CrossChunk(t *testing.T) {
	a := newToolCallAccumulator()
	a.add([]openai.ToolCall{{Index: new(0), ID: "c1", Function: openai.FunctionCall{Name: "search"}}})
	a.add([]openai.ToolCall{{Index: new(0), Function: openai.FunctionCall{Arguments: `{"q"`}}})
	a.add([]openai.ToolCall{{Index: new(0), Function: openai.FunctionCall{Arguments: `:"x"}`}}})

	got := a.finish()
	if len(got) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(got))
	}
	if got[0].ID != "c1" || got[0].Name != "search" {
		t.Fatalf("id/name = %q/%q, want c1/search", got[0].ID, got[0].Name)
	}
	if string(got[0].Args) != `{"q":"x"}` {
		t.Fatalf("args = %s, want {\"q\":\"x\"}", got[0].Args)
	}
}

func TestToolCallAccumulator_MultipleIndicesPreserveOrder(t *testing.T) {
	a := newToolCallAccumulator()
	a.add([]openai.ToolCall{{Index: new(0), ID: "a", Function: openai.FunctionCall{Name: "first", Arguments: "{}"}}})
	a.add([]openai.ToolCall{{Index: new(1), ID: "b", Function: openai.FunctionCall{Name: "second", Arguments: "{}"}}})

	got := a.finish()
	if len(got) != 2 || got[0].Name != "first" || got[1].Name != "second" {
		t.Fatalf("order not preserved: %+v", got)
	}
}

func TestToolCallAccumulator_EmptyArgsFallback(t *testing.T) {
	a := newToolCallAccumulator()
	a.add([]openai.ToolCall{{Index: new(0), ID: "a", Function: openai.FunctionCall{Name: "noargs"}}})

	got := a.finish()
	if len(got) != 1 || string(got[0].Args) != "{}" {
		t.Fatalf("empty args should fall back to {}, got %+v", got)
	}
}

func TestToOpenAIRole(t *testing.T) {
	cases := map[Role]string{
		RoleSystem:    openai.ChatMessageRoleSystem,
		RoleAssistant: openai.ChatMessageRoleAssistant,
		RoleTool:      openai.ChatMessageRoleTool,
		RoleUser:      openai.ChatMessageRoleUser,
		Role("weird"): openai.ChatMessageRoleUser,
	}
	for in, want := range cases {
		if got := toOpenAIRole(in); got != want {
			t.Fatalf("toOpenAIRole(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOpenAIBuildMessages(t *testing.T) {
	p := &openaiProvider{}
	in := []Message{
		{Role: RoleAssistant, Content: "calling", ToolCalls: []ToolCall{{ID: "c1", Name: "search", Args: []byte(`{"q":"x"}`)}}},
		{Role: RoleTool, ToolCallID: "c1", Content: "result"},
		{Role: RoleUser, Content: "see image", Images: []ImagePart{{MediaType: "image/png", Base64: "abc"}}},
	}
	msgs := p.buildMessages(in)
	if len(msgs) != 3 {
		t.Fatalf("got %d messages, want 3", len(msgs))
	}
	if len(msgs[0].ToolCalls) != 1 || msgs[0].ToolCalls[0].Function.Arguments != `{"q":"x"}` {
		t.Fatalf("assistant tool_calls not mapped: %+v", msgs[0].ToolCalls)
	}
	if msgs[1].ToolCallID != "c1" {
		t.Fatalf("tool message ToolCallID = %q, want c1", msgs[1].ToolCallID)
	}
	if msgs[2].Content != "" || len(msgs[2].MultiContent) == 0 {
		t.Fatalf("image message should use MultiContent, got Content=%q MultiContent=%+v", msgs[2].Content, msgs[2].MultiContent)
	}
}

func TestOpenAIImageParts(t *testing.T) {
	parts := openAIImageParts("hello", []ImagePart{
		{MediaType: "image/png", Base64: "abc"},
		{URL: "https://example.com/x.jpg"},
		{},
	})
	if len(parts) != 3 {
		t.Fatalf("got %d parts, want 3 (1 text + 2 image)", len(parts))
	}
	if parts[0].Type != openai.ChatMessagePartTypeText || parts[0].Text != "hello" {
		t.Fatalf("first part should be text 'hello', got %+v", parts[0])
	}
	if parts[1].ImageURL == nil || parts[1].ImageURL.URL != "data:image/png;base64,abc" {
		t.Fatalf("base64 image should become data URL, got %+v", parts[1].ImageURL)
	}
	if parts[2].ImageURL == nil || parts[2].ImageURL.URL != "https://example.com/x.jpg" {
		t.Fatalf("url image should pass through, got %+v", parts[2].ImageURL)
	}

	noText := openAIImageParts("", []ImagePart{{Base64: "abc", MediaType: "image/png"}})
	if len(noText) != 1 || noText[0].Type != openai.ChatMessagePartTypeImageURL {
		t.Fatalf("empty text should yield only the image part, got %+v", noText)
	}
}

func TestAnthropicBuildMessages_ConsecutiveToolsMerge(t *testing.T) {
	p := &anthropicProvider{}
	in := []Message{
		{Role: RoleSystem, Content: "you are a bot"},
		{Role: RoleUser, Content: "q"},
		{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "c1", Name: "search", Args: []byte(`{}`)}}},
		{Role: RoleTool, ToolCallID: "c1", Content: "r1"},
		{Role: RoleTool, ToolCallID: "c2", Content: "r2"},
		{Role: RoleUser, Content: "follow up"},
	}

	systemBlocks, msgs := p.buildMessages(in)

	if len(systemBlocks) != 1 || systemBlocks[0].Text != "you are a bot" {
		t.Fatalf("system blocks = %+v, want single 'you are a bot'", systemBlocks)
	}
	if len(msgs) != 4 {
		t.Fatalf("got %d messages, want 4: %+v", len(msgs), msgs)
	}
	if msgs[0].Role != anthropic.MessageParamRoleUser || msgs[1].Role != anthropic.MessageParamRoleAssistant {
		t.Fatalf("unexpected leading roles: %v / %v", msgs[0].Role, msgs[1].Role)
	}
	merged := msgs[2]
	if merged.Role != anthropic.MessageParamRoleUser {
		t.Fatalf("merged tool results should be a user message, got role %v", merged.Role)
	}
	if len(merged.Content) != 2 {
		t.Fatalf("merged message should have 2 tool_result blocks, got %d", len(merged.Content))
	}
	for i, b := range merged.Content {
		if b.OfToolResult == nil {
			t.Fatalf("merged block %d is not a tool_result: %+v", i, b)
		}
	}
}

func TestAnthropicUserBlocks(t *testing.T) {
	empty := userBlocks(Message{Role: RoleUser})
	if len(empty) != 1 || empty[0].OfText == nil {
		t.Fatalf("empty user message should fall back to a single text block, got %+v", empty)
	}

	blocks := userBlocks(Message{
		Role:    RoleUser,
		Content: "txt",
		Images: []ImagePart{
			{MediaType: "image/png", Base64: "abc"},
			{URL: "https://example.com/y.png"},
		},
	})
	if len(blocks) != 3 {
		t.Fatalf("got %d blocks, want 3 (text + 2 images)", len(blocks))
	}
	if blocks[0].OfText == nil || blocks[0].OfText.Text != "txt" {
		t.Fatalf("first block should be text 'txt', got %+v", blocks[0])
	}
	if blocks[1].OfImage == nil || blocks[2].OfImage == nil {
		t.Fatalf("image blocks not produced: %+v", blocks)
	}
}
