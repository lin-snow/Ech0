// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	model "github.com/lin-snow/ech0/internal/model/comment"
)

var integrationCommentInputSchema = map[string]any{
	"type":     "object",
	"required": []string{"echo_id", "content"},
	"properties": map[string]any{
		"echo_id": map[string]any{
			"type":        "string",
			"format":      "uuid",
			"description": "Post UUID to comment on",
		},
		"content": map[string]any{
			"type":        "string",
			"description": "Comment text (max 200 characters; server validates UTF-8 rune count)",
		},
		"nickname": map[string]any{
			"type":        "string",
			"description": "Display name shown on the comment (default: Integration)",
		},
		"metadata": map[string]any{
			"type":        "string",
			"description": "Optional opaque note for operators (e.g. model id); recorded in server logs, not shown in the public comment UI",
		},
	},
}

func (a *Adapter) registerCommentTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:        "list_comments",
		Title:       "List Comments",
		Description: "List approved public comments for a post (equivalent to GET /api/comments with echo_id). Returns comment objects including nickname, content, created_at, and status when applicable.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"echo_id"},
			"properties": map[string]any{
				"echo_id": map[string]any{"type": "string", "format": "uuid", "description": "Post UUID whose comments to retrieve"},
			},
		},
		Annotations: readOnlyHints(),
	}, a.listComments, authModel.ScopeCommentRead)

	reg.RegisterTool(ToolDefinition{
		Name:  "create_integration_comment",
		Title: "Create Integration Comment",
		Description: "Create a trusted integration/AI comment using the same backend as POST /api/comments/integration: " +
			"no captcha or form_token; source is marked integration; subject to integration rate limits and duplicate checks. " +
			"Requires comment:write. See resource ech0://guide/integration-comment for the REST equivalent (curl) and audience rules.",
		InputSchema: integrationCommentInputSchema,
		Annotations: createHints(),
	}, a.createIntegrationComment, authModel.ScopeCommentWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "create_comment",
		Title:       "Create Comment (Integration)",
		Description: "Alias of create_integration_comment. Prefer this name in agent workflows; behavior and requirements are identical.",
		InputSchema: integrationCommentInputSchema,
		Annotations: createHints(),
	}, a.createIntegrationComment, authModel.ScopeCommentWrite)
}

func (a *Adapter) registerCommentResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://comments/recent",
		Name:        "recent_comments",
		Title:       "Recent Comments",
		Description: "The 20 most recent approved public comments across all posts, newest first.",
		MimeType:    "application/json",
		Cache:       privateCache(liveTTLMs),
	}, a.resourceRecentComments, authModel.ScopeCommentRead)

	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://guide/integration-comment",
		Name:        "integration_comment_guide",
		Title:       "Integration Comment Guide",
		Description: "REST and MCP usage for posting integration/AI comments without captcha (same as POST /api/comments/integration).",
		MimeType:    "text/markdown",
		Cache:       publicCache(staticTTLMs),
	}, a.resourceIntegrationCommentGuide, authModel.ScopeCommentRead)
}

func (a *Adapter) listComments(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	echoID := stringArg(args, "echo_id")
	if echoID == "" {
		return textError("echo_id is required"), nil
	}
	comments, err := a.commentSvc.ListPublicByEchoID(ctx, echoID)
	if err != nil {
		return nil, err
	}
	return jsonResult(comments)
}

func (a *Adapter) createIntegrationComment(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	echoID := stringArg(args, "echo_id")
	content := stringArg(args, "content")
	if echoID == "" || content == "" {
		return textError("echo_id and content are required"), nil
	}
	dto := &model.CreateIntegrationCommentDto{
		EchoID:   echoID,
		Content:  content,
		Nickname: stringArg(args, "nickname"),
		Metadata: stringArg(args, "metadata"),
	}
	result, err := a.commentSvc.CreateIntegrationComment(ctx, "", "MCP", dto)
	if err != nil {
		return nil, err
	}
	return jsonResult(result)
}

func (a *Adapter) resourceRecentComments(ctx context.Context, _ string) (*ResourceReadResult, error) {
	comments, err := a.commentSvc.ListPublicComments(ctx, 20)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(comments)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://comments/recent", MimeType: "application/json", Text: string(data)}},
	}, nil
}

func (a *Adapter) resourceIntegrationCommentGuide(ctx context.Context, _ string) (*ResourceReadResult, error) {
	token := RawTokenFromContext(ctx)
	baseURL := BaseURLFromContext(ctx)
	if baseURL == "" {
		baseURL = "http://localhost:6277"
	}

	md := fmt.Sprintf(`# Ech0 Integration Comment Guide

Integration comments use the **same** server logic whether you call MCP tools or the REST API: trusted JWT, no captcha / form_token, `+"`source: integration`"+` on the stored comment.

## MCP tools

- **`+"`create_comment`"+`** or **`+"`create_integration_comment`"+`**: JSON body fields match the REST DTO below. Requires **`+"`comment:write`"+`** scope.

## REST API (non-MCP clients)

`+"```"+`
POST %s/api/comments/integration
Content-Type: application/json
Authorization: Bearer <access-token>
`+"```"+`

### Token requirements

- **Scope**: `+"`comment:write`"+`
- **Audience**: `+"`mcp-remote`"+` (MCP / AI Agent tokens) **or** `+"`integration`"+` (dedicated integration tokens). Both are accepted by this route.

### Request body (JSON)

| Field | Required | Description |
|-------|----------|-------------|
| echo_id | Yes | Target post UUID |
| content | Yes | Comment text (max 200 characters) |
| nickname | No | Display name (default: Integration) |
| metadata | No | Optional operator/debug string (logged server-side) |

### Example (curl)

`+"```bash"+`
curl -X POST %s/api/comments/integration \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer %s" \
  -d '{"echo_id":"<post-uuid>","content":"Hello from automation","nickname":"AI Bot","metadata":"gpt-4"}'
`+"```"+`

### Response

`+"```json"+`
{
  "code": 1,
  "msg": "...",
  "data": {
    "id": "<comment-uuid>",
    "status": "approved"
  }
}
`+"```"+`

`+"`status`"+` is `+"`pending`"+` when the instance requires comment approval; otherwise `+"`approved`"+`.

## Your credentials (this MCP session)

- **Base URL**: `+"`%s`"+`
- **Bearer Token**: `+"`%s`"+`

## Behaviour notes

- Integration path has **its own rate limits** (per IP and per user when the token is bound to a user).
- Duplicate submissions of the same content in a short window are rejected.
- Comments appear in the admin panel with integration source for moderation.

## Listing comments

- MCP: `+"`list_comments`"+` with `+"`echo_id`"+`
- REST: `+"`GET %s/api/comments?echo_id=<post-uuid>`"+` (public approved comments)
`, baseURL, baseURL, token, baseURL, token, baseURL)

	return &ResourceReadResult{
		Contents: []ResourceContent{{
			URI:      "ech0://guide/integration-comment",
			MimeType: "text/markdown",
			Text:     md,
		}},
	}, nil
}
