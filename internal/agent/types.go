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

// ToolEffect is how much a tool changes, and therefore what the loop has to do
// before running it. It is the tool's own declaration; a caller cannot name one
// and cannot widen the one a tool declares.
//
// EffectUnset is the zero value, and that ordering is the whole point. A tool
// added without a declared effect lands in the strictest branch, not the
// widest: it does not run. The mechanism is not what protects anyone here — the
// default is.
type ToolEffect uint8

const (
	EffectUnset ToolEffect = iota
	EffectRead
	EffectMutate
)

// Tool is one callable the model may reach.
//
// Exactly one body is set, and which one is decided by Effect: a read fills
// Run, a change fills Mutation. That is a deliberate asymmetry rather than one
// function with a flag — a mutation's body is unreachable except through the
// confirmation, and the type is what says so.
type Tool struct {
	Def    ToolDef
	Effect ToolEffect

	Run func(ctx context.Context, args json.RawMessage) (ToolOutput, error)

	Mutation *Mutation

	Interactive bool
}

// Mutation is a change split at the only place it can safely be split: after it
// is known and described, before it is made.
//
// Plan and Confirm are both supplied by the caller and both invoked by the
// loop, in that order, with Apply reachable only from between them. The tool
// never calls Confirm, which is why it cannot forget to — there is no code path
// through this type that reaches a change without a decision, and no prompt,
// context length or model behaviour can produce one.
type Mutation struct {
	Plan func(ctx context.Context, args json.RawMessage) (Plan, error)

	Confirm func(ctx context.Context, p Plan) (Decision, error)
}

// Plan is one worked-out change: what to show, and what to do if it is allowed.
type Plan struct {
	Prompt any

	Apply func(ctx context.Context) (ToolOutput, error)
}

// Decision is what a person said. Anything other than Approved is a no,
// including a Decision nobody filled in.
type Decision struct {
	Approved bool
	Refusal  string
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
	Malformed       string
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
