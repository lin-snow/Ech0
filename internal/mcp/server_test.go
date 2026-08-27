// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func testViewer() viewer.Context {
	return viewer.NewUserViewerWithToken("test-user", "access", []string{"echo:read", "echo:write", "profile:read"}, []string{"mcp-remote"}, "test-jti")
}

func testRequest(t *testing.T, method, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, "/mcp", bytes.NewBufferString(body))
	req = req.WithContext(viewer.WithContext(req.Context(), testViewer()))
	return req
}

func setupTestServer() *Server {
	reg := NewRegistry()
	reg.RegisterTool(ToolDefinition{
		Name:        "echo_tool",
		Description: "test tool",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		return &ToolCallResult{Content: []ContentItem{{Type: "text", Text: "hello"}}}, nil
	}, "echo:read")

	reg.RegisterTool(ToolDefinition{
		Name:        "write_tool",
		Description: "test write tool",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		return &ToolCallResult{Content: []ContentItem{{Type: "text", Text: "written"}}}, nil
	}, "admin:settings")

	reg.RegisterResource(ResourceDefinition{
		URI:      "ech0://test",
		Name:     "test",
		MimeType: "text/plain",
	}, func(_ context.Context, _ string) (*ResourceReadResult, error) {
		return &ResourceReadResult{
			Contents: []ResourceContent{{URI: "ech0://test", MimeType: "text/plain", Text: "test data"}},
		}, nil
	}, "echo:read")

	reg.RegisterResource(ResourceDefinition{
		URI:      "ech0://items/{id}",
		Name:     "item",
		MimeType: "application/json",
	}, func(_ context.Context, uri string) (*ResourceReadResult, error) {
		return &ResourceReadResult{
			Contents: []ResourceContent{{URI: uri, MimeType: "application/json", Text: `{"uri":"` + uri + `"}`}},
		}, nil
	}, "echo:read")

	return NewServer(reg)
}

func withMeta(params map[string]any, version string) map[string]any {
	if params == nil {
		params = map[string]any{}
	}
	params["_meta"] = map[string]any{
		metaKeyProtocolVersion:                       version,
		"io.modelcontextprotocol/clientInfo":         map[string]any{"name": "test-client", "version": "0.0.0"},
		"io.modelcontextprotocol/clientCapabilities": map[string]any{},
	}
	return params
}

func doRaw(t *testing.T, srv *Server, headers map[string]string, body string) (*httptest.ResponseRecorder, Response) {
	t.Helper()
	req := testRequest(t, http.MethodPost, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp Response
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
	}
	return rec, resp
}

func doModern(t *testing.T, srv *Server, method string, params map[string]any) (*httptest.ResponseRecorder, Response) {
	t.Helper()
	params = withMeta(params, ProtocolVersion)
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: method, Params: paramsJSON})

	headers := map[string]string{
		"MCP-Protocol-Version": ProtocolVersion,
		"Mcp-Method":           method,
	}
	if name, ok := params["name"].(string); ok {
		headers["Mcp-Name"] = name
	}
	if uri, ok := params["uri"].(string); ok {
		headers["Mcp-Name"] = uri
	}
	return doRaw(t, srv, headers, string(body))
}

func unmarshalResult[T any](t *testing.T, resp Response) T {
	t.Helper()
	b, _ := json.Marshal(resp.Result)
	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	return result
}

func TestDiscover(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doModern(t, srv, "server/discover", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	result := unmarshalResult[DiscoverResult](t, resp)
	if result.ResultType != resultTypeComplete {
		t.Errorf("resultType = %q, want %q", result.ResultType, resultTypeComplete)
	}
	if len(result.SupportedVersions) != 1 || result.SupportedVersions[0] != ProtocolVersion {
		t.Errorf("supportedVersions = %v, want [%s]", result.SupportedVersions, ProtocolVersion)
	}
	if result.TTLMs <= 0 || result.CacheScope != cacheScopePublic {
		t.Errorf("cache hints = (%d, %q), want positive ttl and %q", result.TTLMs, result.CacheScope, cacheScopePublic)
	}
	info, ok := result.Meta[metaKeyServerInfo].(map[string]any)
	if !ok || info["name"] != ServerName {
		t.Errorf("_meta serverInfo = %v, want name %q", result.Meta[metaKeyServerInfo], ServerName)
	}
}

func TestToolsList(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "tools/list", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ToolsListResult](t, resp)
	if len(result.Tools) != 2 {
		t.Errorf("tools count = %d, want 2", len(result.Tools))
	}
	if result.ResultType != resultTypeComplete {
		t.Errorf("resultType = %q, want %q", result.ResultType, resultTypeComplete)
	}
	if result.TTLMs <= 0 || result.CacheScope != cacheScopePublic {
		t.Errorf("cache hints = (%d, %q), want positive ttl and %q", result.TTLMs, result.CacheScope, cacheScopePublic)
	}
}

func TestToolsCallSuccess(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "tools/call", map[string]any{"name": "echo_tool", "arguments": map[string]any{}})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ToolCallResult](t, resp)
	if result.IsError {
		t.Error("expected success but got isError=true")
	}
	if len(result.Content) == 0 || result.Content[0].Text != "hello" {
		t.Errorf("unexpected content: %v", result.Content)
	}
	if result.ResultType != resultTypeComplete {
		t.Errorf("resultType = %q, want %q", result.ResultType, resultTypeComplete)
	}
}

func TestToolsCallInsufficientScopes(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "tools/call", map[string]any{"name": "write_tool", "arguments": map[string]any{}})
	if resp.Error != nil {
		t.Fatalf("unexpected rpc error: %v", resp.Error)
	}
	result := unmarshalResult[ToolCallResult](t, resp)
	if !result.IsError {
		t.Error("expected isError=true for insufficient scopes")
	}
}

func TestToolsCallNotFound(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doModern(t, srv, "tools/call", map[string]any{"name": "nonexistent", "arguments": map[string]any{}})
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent tool")
	}
	if resp.Error.Code != ErrCodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, ErrCodeInvalidParams)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d (application-level error)", rec.Code, http.StatusOK)
	}
}

func TestResourcesList(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/list", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ResourcesListResult](t, resp)
	if len(result.Resources) != 2 {
		t.Errorf("resources count = %d, want 2", len(result.Resources))
	}
}

func TestResourcesReadSuccess(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/read", map[string]any{"uri": "ech0://test"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ResourceReadResult](t, resp)
	if len(result.Contents) == 0 || result.Contents[0].Text != "test data" {
		t.Errorf("unexpected content: %v", result.Contents)
	}
	if result.CacheScope != cacheScopePrivate {
		t.Errorf("cacheScope = %q, want %q (authorization-scoped data)", result.CacheScope, cacheScopePrivate)
	}
}

func TestResourcesReadPrefixMatch(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/read", map[string]any{"uri": "ech0://items/abc-123"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ResourceReadResult](t, resp)
	if len(result.Contents) == 0 {
		t.Fatal("expected content from prefix-matched resource")
	}
	if result.Contents[0].URI != "ech0://items/abc-123" {
		t.Errorf("URI = %q, want %q", result.Contents[0].URI, "ech0://items/abc-123")
	}
}

func TestResourcesReadNotFound(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/read", map[string]any{"uri": "ech0://missing"})
	if resp.Error == nil {
		t.Fatal("expected error for nonexistent resource")
	}
}

func TestResourcesReadBase64SentinelName(t *testing.T) {
	srv := setupTestServer()
	uri := "ech0://test"
	params := withMeta(map[string]any{"uri": uri}, ProtocolVersion)
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "resources/read", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": ProtocolVersion,
		"Mcp-Method":           "resources/read",
		"Mcp-Name":             "=?base64?" + base64.StdEncoding.EncodeToString([]byte(uri)) + "?=",
	}, string(body))
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMethodNotFound(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doModern(t, srv, "unknown/method", nil)
	if resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, ErrCodeMethodNotFound)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestInitializeRemoved(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doRaw(t, srv, nil, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	if resp.Error == nil {
		t.Fatal("expected error for legacy initialize")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, ErrCodeMethodNotFound)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	data, ok := resp.Error.Data.(map[string]any)
	if !ok {
		t.Fatalf("error data = %v, want map with supported versions", resp.Error.Data)
	}
	supported, _ := data["supported"].([]any)
	if len(supported) != 1 || supported[0] != ProtocolVersion {
		t.Errorf("supported = %v, want [%s]", supported, ProtocolVersion)
	}
}

func TestMissingProtocolVersionHeader(t *testing.T) {
	srv := setupTestServer()
	params := withMeta(nil, ProtocolVersion)
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{"Mcp-Method": "tools/list"}, string(body))
	if resp.Error == nil || resp.Error.Code != ErrCodeHeaderMismatch {
		t.Fatalf("expected header mismatch error, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestVersionHeaderBodyMismatch(t *testing.T) {
	srv := setupTestServer()
	params := withMeta(nil, ProtocolVersion)
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": "2025-11-25",
		"Mcp-Method":           "tools/list",
	}, string(body))
	if resp.Error == nil || resp.Error.Code != ErrCodeHeaderMismatch {
		t.Fatalf("expected header mismatch error, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUnsupportedProtocolVersion(t *testing.T) {
	srv := setupTestServer()
	params := withMeta(nil, "2025-11-25")
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": "2025-11-25",
		"Mcp-Method":           "tools/list",
	}, string(body))
	if resp.Error == nil || resp.Error.Code != ErrCodeUnsupportedProtocolVersion {
		t.Fatalf("expected unsupported version error, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	data, ok := resp.Error.Data.(map[string]any)
	if !ok {
		t.Fatalf("error data = %v, want map", resp.Error.Data)
	}
	if data["requested"] != "2025-11-25" {
		t.Errorf("requested = %v, want 2025-11-25", data["requested"])
	}
	supported, _ := data["supported"].([]any)
	if len(supported) != 1 || supported[0] != ProtocolVersion {
		t.Errorf("supported = %v, want [%s]", supported, ProtocolVersion)
	}
}

func TestMcpMethodHeaderMismatch(t *testing.T) {
	srv := setupTestServer()
	params := withMeta(nil, ProtocolVersion)
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": ProtocolVersion,
		"Mcp-Method":           "resources/list",
	}, string(body))
	if resp.Error == nil || resp.Error.Code != ErrCodeHeaderMismatch {
		t.Fatalf("expected header mismatch error, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMissingMcpNameOnToolsCall(t *testing.T) {
	srv := setupTestServer()
	params := withMeta(map[string]any{"name": "echo_tool", "arguments": map[string]any{}}, ProtocolVersion)
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/call", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": ProtocolVersion,
		"Mcp-Method":           "tools/call",
	}, string(body))
	if resp.Error == nil || resp.Error.Code != ErrCodeHeaderMismatch {
		t.Fatalf("expected header mismatch error, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNotificationAccepted(t *testing.T) {
	srv := setupTestServer()
	rec, _ := doRaw(t, srv, nil, `{"jsonrpc":"2.0","method":"notifications/cancelled","params":{}}`)
	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected empty body for notification, got %q", rec.Body.String())
	}
}

func TestNullRequestID(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doRaw(t, srv, nil, `{"jsonrpc":"2.0","id":null,"method":"tools/list","params":{}}`)
	if resp.Error == nil || resp.Error.Code != ErrCodeInvalidRequest {
		t.Fatalf("expected invalid request error, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetMethodNotAllowed(t *testing.T) {
	srv := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestAdapterRegistersDiscoveryCapabilities(t *testing.T) {
	reg := NewRegistry()
	NewAdapter(nil, nil, nil, nil, nil, nil, nil, nil, nil).RegisterAll(reg)

	for _, name := range []string{"get_hot_posts", "get_random_post", "get_on_this_day_posts"} {
		_, scopes, ok := reg.LookupTool(name)
		if !ok {
			t.Errorf("tool %q not registered", name)
			continue
		}
		if len(scopes) != 1 || scopes[0] != authModel.ScopeEchoRead {
			t.Errorf("tool %q scopes = %v, want [%s]", name, scopes, authModel.ScopeEchoRead)
		}
	}

	_, scopes, ok := reg.LookupResource("ech0://stats/visitors")
	if !ok {
		t.Fatal("resource ech0://stats/visitors not registered")
	}
	if len(scopes) != 1 || scopes[0] != authModel.ScopeAdminSettings {
		t.Errorf("visitor stats scopes = %v, want [%s]", scopes, authModel.ScopeAdminSettings)
	}
}

func TestInvalidJSON(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doRaw(t, srv, nil, "not json")
	if resp.Error == nil || resp.Error.Code != ErrCodeParse {
		t.Errorf("expected parse error, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
