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
		Annotations: readOnlyHints(),
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		return &ToolCallResult{Content: []ContentItem{{Type: "text", Text: "hello"}}}, nil
	}, "echo:read")

	reg.RegisterTool(ToolDefinition{
		Name:        "json_tool",
		Description: "test structured tool",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Annotations: readOnlyHints(),
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		return jsonResult(map[string]any{"count": 2})
	}, "echo:read")

	reg.RegisterTool(ToolDefinition{
		Name:        "array_tool",
		Description: "test array-returning tool",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Annotations: readOnlyHints(),
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		return jsonResult([]map[string]any{{"id": "1"}, {"id": "2"}})
	}, "echo:read")

	reg.RegisterTool(ToolDefinition{
		Name:        "write_tool",
		Description: "test write tool",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{}},
		Annotations: destructiveHints(),
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
		URI:      "ech0://guide",
		Name:     "guide",
		MimeType: "text/markdown",
		Cache:    publicCache(staticTTLMs),
	}, func(_ context.Context, _ string) (*ResourceReadResult, error) {
		return &ResourceReadResult{
			Contents: []ResourceContent{{URI: "ech0://guide", MimeType: "text/markdown", Text: "# guide"}},
		}, nil
	}, "echo:read")

	reg.RegisterResourceTemplate(ResourceTemplateDefinition{
		URITemplate: "ech0://items/{id}",
		Name:        "item",
		MimeType:    "application/json",
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
		metaKeyProtocolVersion:               version,
		"io.modelcontextprotocol/clientInfo": map[string]any{"name": "test-client", "version": "0.0.0"},
		metaKeyClientCapabilities:            map[string]any{},
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

func doLegacy(
	t *testing.T,
	srv *Server,
	version, method string,
	params map[string]any,
) (*httptest.ResponseRecorder, Response) {
	t.Helper()
	if params == nil {
		params = map[string]any{}
	}
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: method, Params: paramsJSON})

	headers := map[string]string{}
	if version != "" {
		headers["MCP-Protocol-Version"] = version
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

func resultFields(t *testing.T, resp Response) map[string]any {
	t.Helper()
	b, _ := json.Marshal(resp.Result)
	var fields map[string]any
	if err := json.Unmarshal(b, &fields); err != nil {
		t.Fatalf("unmarshal result fields: %v", err)
	}
	return fields
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
	if len(result.SupportedVersions) != len(SupportedVersions) || result.SupportedVersions[0] != ProtocolVersion {
		t.Errorf("supportedVersions = %v, want %v", result.SupportedVersions, SupportedVersions)
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
	if len(result.Tools) != 4 {
		t.Errorf("tools count = %d, want 4", len(result.Tools))
	}
	if result.Tools[0].Name != "echo_tool" || result.Tools[3].Name != "write_tool" {
		t.Errorf("tools are not in registration order: %v", result.Tools)
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

func TestToolsCallReturnsStructuredContent(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "tools/call", map[string]any{"name": "json_tool", "arguments": map[string]any{}})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	fields := resultFields(t, resp)
	structured, ok := fields["structuredContent"].(map[string]any)
	if !ok {
		t.Fatalf("structuredContent = %v, want object", fields["structuredContent"])
	}
	if structured["count"] != float64(2) {
		t.Errorf("structuredContent.count = %v, want 2", structured["count"])
	}
	content, _ := fields["content"].([]any)
	if len(content) == 0 {
		t.Error("structured results must still carry the serialized JSON as text content")
	}
}

func TestToolsCallOmitsArrayStructuredContent(t *testing.T) {
	srv := setupTestServer()
	for _, tc := range []struct {
		name string
		call func() (*httptest.ResponseRecorder, Response)
	}{
		{"modern", func() (*httptest.ResponseRecorder, Response) {
			return doModern(t, srv, "tools/call", map[string]any{"name": "array_tool", "arguments": map[string]any{}})
		}},
		{"legacy", func() (*httptest.ResponseRecorder, Response) {
			return doLegacy(t, srv, protocolVersion20251125, "tools/call", map[string]any{"name": "array_tool", "arguments": map[string]any{}})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, resp := tc.call()
			if resp.Error != nil {
				t.Fatalf("unexpected error: %v", resp.Error)
			}
			fields := resultFields(t, resp)
			if _, present := fields["structuredContent"]; present {
				t.Errorf("structuredContent = %v, want omitted for an array payload", fields["structuredContent"])
			}
			content, _ := fields["content"].([]any)
			if len(content) == 0 {
				t.Error("array payload must still be delivered in the text content block")
			}
		})
	}
}

func TestToolsCallInsufficientScopes(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doModern(t, srv, "tools/call", map[string]any{"name": "write_tool", "arguments": map[string]any{}})
	if resp.Error == nil || resp.Error.Code != ErrCodeInsufficientScope {
		t.Fatalf("expected insufficient scope error, got %v", resp.Error)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	challenge := rec.Header().Get("WWW-Authenticate")
	if challenge != `Bearer error="insufficient_scope", scope="admin:settings"` {
		t.Errorf("WWW-Authenticate = %q, want an insufficient_scope challenge naming admin:settings", challenge)
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

func TestResourcesListExcludesTemplates(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/list", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ResourcesListResult](t, resp)
	if len(result.Resources) != 2 {
		t.Fatalf("resources count = %d, want 2 concrete resources", len(result.Resources))
	}
	for _, res := range result.Resources {
		if res.URI == "ech0://items/{id}" {
			t.Error("resources/list must not advertise parameterised URIs")
		}
	}
}

func TestResourceTemplatesList(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/templates/list", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ResourceTemplatesListResult](t, resp)
	if len(result.ResourceTemplates) != 1 {
		t.Fatalf("templates count = %d, want 1", len(result.ResourceTemplates))
	}
	if result.ResourceTemplates[0].URITemplate != "ech0://items/{id}" {
		t.Errorf("uriTemplate = %q, want %q", result.ResourceTemplates[0].URITemplate, "ech0://items/{id}")
	}
	if result.TTLMs <= 0 || result.CacheScope != cacheScopePublic {
		t.Errorf("cache hints = (%d, %q), want positive ttl and %q", result.TTLMs, result.CacheScope, cacheScopePublic)
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
		t.Errorf("cacheScope = %q, want %q (unset policy must not be shareable)", result.CacheScope, cacheScopePrivate)
	}
}

func TestResourcesReadUsesRegisteredCachePolicy(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/read", map[string]any{"uri": "ech0://guide"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ResourceReadResult](t, resp)
	if result.CacheScope != cacheScopePublic || result.TTLMs != staticTTLMs {
		t.Errorf("cache hints = (%d, %q), want (%d, %q)", result.TTLMs, result.CacheScope, staticTTLMs, cacheScopePublic)
	}
}

func TestResourcesReadTemplateMatch(t *testing.T) {
	srv := setupTestServer()
	_, resp := doModern(t, srv, "resources/read", map[string]any{"uri": "ech0://items/abc-123"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[ResourceReadResult](t, resp)
	if len(result.Contents) == 0 {
		t.Fatal("expected content from template-matched resource")
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
	if resp.Error.Code != ErrCodeInvalidParams {
		t.Errorf("error code = %d, want %d", resp.Error.Code, ErrCodeInvalidParams)
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

func TestLegacyInitializeHandshake(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doRaw(t, srv, nil,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"legacy","version":"1"}}}`)
	if resp.Error != nil {
		t.Fatalf("legacy initialize must be answered, got %v", resp.Error)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	result := unmarshalResult[InitializeResult](t, resp)
	if result.ProtocolVersion != protocolVersion20250618 {
		t.Errorf("protocolVersion = %q, want %q (echo the client's supported revision)", result.ProtocolVersion, protocolVersion20250618)
	}
	if result.ServerInfo.Name != ServerName {
		t.Errorf("serverInfo.name = %q, want %q", result.ServerInfo.Name, ServerName)
	}
	if result.Capabilities.Tools == nil || result.Capabilities.Resources == nil {
		t.Error("legacy handshake must advertise tools and resources capabilities")
	}
	if rec.Header().Get("Mcp-Session-Id") != "" {
		t.Error("no session may be minted: the server is stateless in both eras")
	}
}

func TestLegacyInitializeUnknownVersionNegotiatesDown(t *testing.T) {
	srv := setupTestServer()
	_, resp := doRaw(t, srv, nil,
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := unmarshalResult[InitializeResult](t, resp)
	if result.ProtocolVersion != preferredLegacyVersion {
		t.Errorf("protocolVersion = %q, want %q", result.ProtocolVersion, preferredLegacyVersion)
	}
}

func TestLegacyInitializedNotificationAccepted(t *testing.T) {
	srv := setupTestServer()
	rec, _ := doRaw(t, srv, nil, `{"jsonrpc":"2.0","method":"notifications/initialized"}`)
	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
}

func TestLegacyPing(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doLegacy(t, srv, protocolVersion20251125, "ping", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestModernPingRemoved(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doModern(t, srv, "ping", nil)
	if resp.Error == nil || resp.Error.Code != ErrCodeMethodNotFound {
		t.Fatalf("expected method not found, got %v", resp.Error)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestModernInitializeRejected(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doModern(t, srv, "initialize", nil)
	if resp.Error == nil || resp.Error.Code != ErrCodeMethodNotFound {
		t.Fatalf("expected method not found, got %v", resp.Error)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	data, ok := resp.Error.Data.(map[string]any)
	if !ok {
		t.Fatalf("error data = %v, want map with supported versions", resp.Error.Data)
	}
	if supported, _ := data["supported"].([]any); len(supported) != len(SupportedVersions) {
		t.Errorf("supported = %v, want %v", data["supported"], SupportedVersions)
	}
}

func TestLegacyToolsCallWithoutModernMetadata(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doLegacy(t, srv, protocolVersion20251125, "tools/call", map[string]any{
		"name":      "echo_tool",
		"arguments": map[string]any{},
	})
	if resp.Error != nil {
		t.Fatalf("legacy tools/call must not require modern headers, got %v", resp.Error)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	fields := resultFields(t, resp)
	for _, key := range []string{"resultType", "ttlMs", "cacheScope", "_meta"} {
		if _, present := fields[key]; present {
			t.Errorf("legacy result must omit %q, got %v", key, fields[key])
		}
	}
}

func TestLegacyToolsListWithoutVersionHeader(t *testing.T) {
	srv := setupTestServer()
	rec, resp := doLegacy(t, srv, "", "tools/list", nil)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	fields := resultFields(t, resp)
	if _, present := fields["ttlMs"]; present {
		t.Error("legacy tools/list must omit cache hints")
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

func TestMissingClientCapabilities(t *testing.T) {
	srv := setupTestServer()
	paramsJSON, _ := json.Marshal(map[string]any{
		"_meta": map[string]any{metaKeyProtocolVersion: ProtocolVersion},
	})
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": ProtocolVersion,
		"Mcp-Method":           "tools/list",
	}, string(body))
	if resp.Error == nil || resp.Error.Code != ErrCodeInvalidParams {
		t.Fatalf("expected invalid params for missing clientCapabilities, got %v", resp.Error)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMissingBodyProtocolVersion(t *testing.T) {
	srv := setupTestServer()
	paramsJSON, _ := json.Marshal(map[string]any{
		"_meta": map[string]any{metaKeyClientCapabilities: map[string]any{}},
	})
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": ProtocolVersion,
		"Mcp-Method":           "tools/list",
	}, string(body))
	if resp.Error == nil || resp.Error.Code != ErrCodeInvalidParams {
		t.Fatalf("expected invalid params for missing _meta protocolVersion, got %v", resp.Error)
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
		"MCP-Protocol-Version": protocolVersion20251125,
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
	params := withMeta(nil, "1900-01-01")
	paramsJSON, _ := json.Marshal(params)
	body, _ := json.Marshal(Request{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "tools/list", Params: paramsJSON})
	rec, resp := doRaw(t, srv, map[string]string{
		"MCP-Protocol-Version": "1900-01-01",
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
	if data["requested"] != "1900-01-01" {
		t.Errorf("requested = %v, want 1900-01-01", data["requested"])
	}
	supported, _ := data["supported"].([]any)
	if len(supported) != len(SupportedVersions) || supported[0] != ProtocolVersion {
		t.Errorf("supported = %v, want %v", supported, SupportedVersions)
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
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		req := httptest.NewRequest(method, "/mcp", nil)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s status = %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}

func TestAdapterRegistersDiscoveryCapabilities(t *testing.T) {
	reg := NewRegistry()
	NewAdapter(nil, nil, nil, nil, nil, nil, nil, nil, nil).RegisterAll(reg)

	for _, name := range []string{"get_hot_posts", "get_random_post", "get_on_this_day_posts"} {
		binding, ok := reg.LookupTool(name)
		if !ok {
			t.Errorf("tool %q not registered", name)
			continue
		}
		if len(binding.Scopes) != 1 || binding.Scopes[0] != authModel.ScopeEchoRead {
			t.Errorf("tool %q scopes = %v, want [%s]", name, binding.Scopes, authModel.ScopeEchoRead)
		}
	}

	binding, ok := reg.LookupResource("ech0://stats/visitors")
	if !ok {
		t.Fatal("resource ech0://stats/visitors not registered")
	}
	if len(binding.Scopes) != 1 || binding.Scopes[0] != authModel.ScopeAdminSettings {
		t.Errorf("visitor stats scopes = %v, want [%s]", binding.Scopes, authModel.ScopeAdminSettings)
	}
}

func TestAdapterAnnotatesDestructiveTools(t *testing.T) {
	reg := NewRegistry()
	NewAdapter(nil, nil, nil, nil, nil, nil, nil, nil, nil).RegisterAll(reg)

	byName := make(map[string]ToolDefinition)
	for _, def := range reg.ToolDefinitions() {
		if def.Annotations == nil {
			t.Fatalf("tool %q has no behaviour annotations", def.Name)
		}
		byName[def.Name] = def
	}

	for _, name := range []string{"delete_post", "delete_tag", "delete_file", "delete_webhook", "update_post"} {
		anno := byName[name].Annotations
		if anno == nil || anno.DestructiveHint == nil || !*anno.DestructiveHint {
			t.Errorf("tool %q must be marked destructive", name)
		}
	}
	for _, name := range []string{"search_posts", "get_post", "list_files", "list_webhooks"} {
		anno := byName[name].Annotations
		if anno == nil || anno.ReadOnlyHint == nil || !*anno.ReadOnlyHint {
			t.Errorf("tool %q must be marked read-only", name)
		}
	}
	if anno := byName["create_post"].Annotations; anno == nil || anno.DestructiveHint == nil || *anno.DestructiveHint {
		t.Error("create_post must not be marked destructive")
	}
}

func TestAdapterRegistersPostTemplate(t *testing.T) {
	reg := NewRegistry()
	NewAdapter(nil, nil, nil, nil, nil, nil, nil, nil, nil).RegisterAll(reg)

	for _, def := range reg.ResourceDefinitions() {
		if def.URI == "ech0://posts/{id}" {
			t.Fatal("ech0://posts/{id} must be advertised as a template, not a concrete resource")
		}
	}
	var found bool
	for _, tpl := range reg.ResourceTemplateDefinitions() {
		if tpl.URITemplate == "ech0://posts/{id}" {
			found = true
		}
	}
	if !found {
		t.Error("ech0://posts/{id} missing from resources/templates/list")
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
