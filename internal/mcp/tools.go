// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

type ToolDefinition struct {
	Name        string         `json:"name"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ToolsListResult struct {
	ResultEnvelope
	CacheInfo
	Tools []ToolDefinition `json:"tools"`
}

type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type ToolCallResult struct {
	ResultEnvelope
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError"`
}
