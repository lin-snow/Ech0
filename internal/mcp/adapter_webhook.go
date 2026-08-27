// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

func (a *Adapter) registerWebhookTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:        "list_webhooks",
		Title:       "List Webhooks",
		Description: "List all configured webhooks. Returns an array of webhook objects (id, name, url, is_active, last_status, last_trigger, timestamps). Secrets are never exposed.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, a.listWebhooks, authModel.ScopeAdminSettings)

	reg.RegisterTool(ToolDefinition{
		Name:        "create_webhook",
		Title:       "Create Webhook",
		Description: "Create a new webhook endpoint. Returns {message} on success.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"name", "url"},
			"properties": map[string]any{
				"name":      map[string]any{"type": "string", "description": "Webhook display name"},
				"url":       map[string]any{"type": "string", "format": "uri", "description": "Endpoint URL that will receive POST requests"},
				"secret":    map[string]any{"type": "string", "description": "Optional HMAC signing secret for request verification"},
				"is_active": map[string]any{"type": "boolean", "description": "Enable or disable the webhook", "default": true},
			},
		},
	}, a.createWebhook, authModel.ScopeAdminSettings)

	reg.RegisterTool(ToolDefinition{
		Name:        "update_webhook",
		Title:       "Update Webhook",
		Description: "Update an existing webhook by ID. All fields in the body replace the current values. Returns {id, message}.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id", "name", "url"},
			"properties": map[string]any{
				"id":        map[string]any{"type": "string", "format": "uuid", "description": "Webhook UUID"},
				"name":      map[string]any{"type": "string", "description": "Webhook display name"},
				"url":       map[string]any{"type": "string", "format": "uri", "description": "Endpoint URL"},
				"secret":    map[string]any{"type": "string", "description": "HMAC signing secret (leave empty to clear)"},
				"is_active": map[string]any{"type": "boolean", "description": "Enable or disable the webhook"},
			},
		},
	}, a.updateWebhook, authModel.ScopeAdminSettings)

	reg.RegisterTool(ToolDefinition{
		Name:        "delete_webhook",
		Title:       "Delete Webhook",
		Description: "Delete a webhook by ID. Returns {id, message}. This action cannot be undone.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "Webhook UUID"},
			},
		},
	}, a.deleteWebhook, authModel.ScopeAdminSettings)

	reg.RegisterTool(ToolDefinition{
		Name:        "test_webhook",
		Title:       "Test Webhook",
		Description: "Send a test POST request to the webhook endpoint. Returns {id, message} indicating whether the test was dispatched.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "Webhook UUID"},
			},
		},
	}, a.testWebhook, authModel.ScopeAdminSettings)
}

func (a *Adapter) listWebhooks(ctx context.Context, _ map[string]any) (*ToolCallResult, error) {
	webhooks, err := a.settingSvc.GetAllWebhooks(ctx)
	if err != nil {
		return nil, err
	}
	return jsonResult(webhooks)
}

func (a *Adapter) createWebhook(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	name := stringArg(args, "name")
	if name == "" {
		return textError("name is required"), nil
	}
	url := stringArg(args, "url")
	if url == "" {
		return textError("url is required"), nil
	}
	dto := &settingModel.WebhookDto{
		Name:     name,
		URL:      url,
		Secret:   stringArg(args, "secret"),
		IsActive: true,
	}
	if v, ok := args["is_active"]; ok {
		if b, ok := v.(bool); ok {
			dto.IsActive = b
		}
	}
	if err := a.settingSvc.CreateWebhook(ctx, dto); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"message": "webhook created successfully"})
}

func (a *Adapter) updateWebhook(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	name := stringArg(args, "name")
	if name == "" {
		return textError("name is required"), nil
	}
	url := stringArg(args, "url")
	if url == "" {
		return textError("url is required"), nil
	}
	dto := &settingModel.WebhookDto{
		Name:     name,
		URL:      url,
		Secret:   stringArg(args, "secret"),
		IsActive: boolArg(args, "is_active"),
	}
	if err := a.settingSvc.UpdateWebhook(ctx, id, dto); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "webhook updated successfully"})
}

func (a *Adapter) deleteWebhook(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.settingSvc.DeleteWebhook(ctx, id); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "webhook deleted successfully"})
}

func (a *Adapter) testWebhook(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.settingSvc.TestWebhook(ctx, id); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "webhook test dispatched"})
}
