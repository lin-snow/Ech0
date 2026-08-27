// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"context"
	"encoding/json"
	"time"

	model "github.com/lin-snow/ech0/internal/model/setting"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role       Role
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
	Images     []ImagePart
}

type ImagePart struct {
	MediaType string
	Base64    string
	URL       string
}

type ToolCall struct {
	ID   string
	Name string
	Args json.RawMessage
}

type ToolDef struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

type Tool struct {
	Def     ToolDef
	Execute func(ctx context.Context, args json.RawMessage) (ToolOutput, error)
}

type ToolOutput struct {
	Content string
	Meta    any
	Images  []ImagePart
}

type Request struct {
	Messages    []Message
	Tools       []ToolDef
	Temperature *float32
	MaxTokens   int
}

type Response struct {
	Text string
}

type EventKind int

const (
	EventTextDelta EventKind = iota
	EventReasoningDelta
	EventToolCall
	EventDone
	EventError
)

type Event struct {
	Kind     EventKind
	Text     string
	ToolCall ToolCall
	Err      error
}

type RunStrings struct {
	DedupNote       string
	UnknownTool     string
	ToolError       string
	ImageNote       string
	ContextTrimNote string
}

type RunRequest struct {
	Setting          model.AgentSetting
	Messages         []Message
	Tools            []Tool
	MaxRounds        int
	Temp             *float32
	Strings          RunStrings
	Timeout          time.Duration
	MaxContextTokens int
}

type AgentEventKind int

const (
	AgentDelta AgentEventKind = iota
	AgentReasoning
	AgentSearching
	AgentToolResult
	AgentDone
	AgentError
)

type AgentEvent struct {
	Kind     AgentEventKind
	Text     string
	ToolName string
	ToolArgs json.RawMessage
	Meta     any
	Err      error
}
