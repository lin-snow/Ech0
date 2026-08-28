// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func TestManifestMirrorsRegistry(t *testing.T) {
	srv := setupTestServer()
	manifest := srv.registry.Manifest()

	if manifest.Path != MCPEndpointPath {
		t.Errorf("path = %q, want %q", manifest.Path, MCPEndpointPath)
	}
	if manifest.Transport != TransportStreamableHTTP {
		t.Errorf("transport = %q, want %q", manifest.Transport, TransportStreamableHTTP)
	}
	if manifest.Audience != authModel.AudienceMCPRemote {
		t.Errorf("audience = %q, want %q", manifest.Audience, authModel.AudienceMCPRemote)
	}
	if len(manifest.ProtocolVersions) != len(SupportedVersions) {
		t.Errorf("protocol versions = %v, want %v", manifest.ProtocolVersions, SupportedVersions)
	}

	if len(manifest.Tools) != len(srv.registry.ToolDefinitions()) {
		t.Errorf("tools = %d, want %d", len(manifest.Tools), len(srv.registry.ToolDefinitions()))
	}
	if len(manifest.Resources) != len(srv.registry.ResourceDefinitions()) {
		t.Errorf("resources = %d, want %d", len(manifest.Resources), len(srv.registry.ResourceDefinitions()))
	}
	if len(manifest.ResourceTemplates) != len(srv.registry.ResourceTemplateDefinitions()) {
		t.Errorf(
			"templates = %d, want %d",
			len(manifest.ResourceTemplates),
			len(srv.registry.ResourceTemplateDefinitions()),
		)
	}
}

func TestManifestCarriesScopesAndHints(t *testing.T) {
	manifest := setupTestServer().registry.Manifest()

	tools := make(map[string]ManifestTool, len(manifest.Tools))
	for _, tool := range manifest.Tools {
		tools[tool.Name] = tool
	}

	readTool, ok := tools["echo_tool"]
	if !ok {
		t.Fatalf("echo_tool missing from manifest")
	}
	if !readTool.ReadOnly || readTool.Destructive {
		t.Errorf("echo_tool hints = (readOnly %v, destructive %v), want (true, false)", readTool.ReadOnly, readTool.Destructive)
	}
	if len(readTool.Scopes) != 1 || readTool.Scopes[0] != authModel.ScopeEchoRead {
		t.Errorf("echo_tool scopes = %v, want [%s]", readTool.Scopes, authModel.ScopeEchoRead)
	}

	writeTool, ok := tools["write_tool"]
	if !ok {
		t.Fatalf("write_tool missing from manifest")
	}
	if writeTool.ReadOnly || !writeTool.Destructive {
		t.Errorf("write_tool hints = (readOnly %v, destructive %v), want (false, true)", writeTool.ReadOnly, writeTool.Destructive)
	}
	if len(writeTool.Scopes) != 1 || writeTool.Scopes[0] != authModel.ScopeAdminSettings {
		t.Errorf("write_tool scopes = %v, want [%s]", writeTool.Scopes, authModel.ScopeAdminSettings)
	}

	if len(manifest.ResourceTemplates) != 1 || manifest.ResourceTemplates[0].URITemplate != "ech0://items/{id}" {
		t.Errorf("templates = %v, want the ech0://items/{id} template", manifest.ResourceTemplates)
	}
	if len(manifest.ResourceTemplates[0].Scopes) != 1 {
		t.Errorf("template scopes = %v, want exactly one", manifest.ResourceTemplates[0].Scopes)
	}
}

func TestGetManifestHandler(t *testing.T) {
	h := &Handler{server: setupTestServer()}

	result, err := h.GetManifest(context.Background(), &GetManifestInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Data.Tools) == 0 {
		t.Errorf("handler returned no tools")
	}
	if result.Data.Path != MCPEndpointPath {
		t.Errorf("path = %q, want %q", result.Data.Path, MCPEndpointPath)
	}
}
