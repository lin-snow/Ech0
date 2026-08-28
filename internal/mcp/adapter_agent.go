// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func (a *Adapter) registerAgentTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:  "get_recent",
		Title: "Get Recent Summary",
		Description: "Return an AI-generated natural-language summary of the site owner's recent activity. " +
			"The result is cached and may take a few seconds on first call if the cache has expired.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Annotations: readOnlyHints(),
	}, a.getRecent, authModel.ScopeEchoRead)
}

func (a *Adapter) getRecent(ctx context.Context, _ map[string]any) (*ToolCallResult, error) {
	summary, err := a.agentSvc.GetRecent(ctx)
	if err != nil {
		return nil, err
	}
	return textResult(summary), nil
}
