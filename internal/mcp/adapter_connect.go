// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
)

func (a *Adapter) registerConnectTools(reg *Registry) {
	reg.RegisterTool(ToolDefinition{
		Name:        "list_connects",
		Title:       "List Connects",
		Description: "List all saved peer connections (id and URL). These are the remote Ech0 instances the current instance is tracking.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, a.listConnects, authModel.ScopeConnectRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "get_connects_info",
		Title:       "Get Connects Info",
		Description: "Fetch aggregated public info (name, URL, logo, post stats, version) for every saved peer. Data is cached for 30 minutes; the first call after expiry may take a few seconds as the server fans out HTTP requests to each peer.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}, a.getConnectsInfo, authModel.ScopeConnectRead)

	reg.RegisterTool(ToolDefinition{
		Name:        "add_connect",
		Title:       "Add Connect",
		Description: "Add a remote Ech0 instance as a peer connection. The connect_url must point to the /connect endpoint of the remote instance (e.g. https://example.com/api/connect). Returns {id, message}.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"connect_url"},
			"properties": map[string]any{
				"connect_url": map[string]any{"type": "string", "format": "uri", "description": "Full URL to the remote instance's /connect endpoint"},
			},
		},
	}, a.addConnect, authModel.ScopeConnectWrite)

	reg.RegisterTool(ToolDefinition{
		Name:        "delete_connect",
		Title:       "Delete Connect",
		Description: "Remove a saved peer connection by ID. Returns {id, message}. This action cannot be undone.",
		InputSchema: map[string]any{
			"type":     "object",
			"required": []string{"id"},
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "format": "uuid", "description": "Connect UUID"},
			},
		},
	}, a.deleteConnect, authModel.ScopeConnectWrite)
}

func (a *Adapter) registerConnectResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://connect/self",
		Name:        "connect_self",
		Title:       "Current Instance Info",
		Description: "Public card of this Ech0 instance: server name, URL, logo, today/total post counts, owner username, and version.",
		MimeType:    "application/json",
	}, a.resourceConnectSelf, authModel.ScopeConnectRead)
}

func (a *Adapter) listConnects(ctx context.Context, _ map[string]any) (*ToolCallResult, error) {
	connects, err := a.connectSvc.GetConnects()
	if err != nil {
		return nil, err
	}
	return jsonResult(connects)
}

func (a *Adapter) getConnectsInfo(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
	info, err := a.connectSvc.GetConnectsInfo()
	if err != nil {
		return nil, err
	}
	return jsonResult(info)
}

func (a *Adapter) addConnect(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	url := stringArg(args, "connect_url")
	if url == "" {
		return textError("connect_url is required"), nil
	}
	connected := connectModel.Connected{ConnectURL: url}
	if err := a.connectSvc.AddConnect(ctx, connected); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": connected.ID, "message": "connect added successfully"})
}

func (a *Adapter) deleteConnect(ctx context.Context, args map[string]any) (*ToolCallResult, error) {
	id := stringArg(args, "id")
	if id == "" {
		return textError("id is required"), nil
	}
	if err := a.connectSvc.DeleteConnect(ctx, id); err != nil {
		return nil, err
	}
	return jsonResult(map[string]string{"id": id, "message": "connect deleted successfully"})
}

func (a *Adapter) resourceConnectSelf(_ context.Context, _ string) (*ResourceReadResult, error) {
	info, err := a.connectSvc.GetConnect()
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(info)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://connect/self", MimeType: "application/json", Text: string(data)}},
	}, nil
}
