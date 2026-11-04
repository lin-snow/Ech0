// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func elicitationCaps() map[string]any {
	return map[string]any{"elicitation": map[string]any{"form": map[string]any{}}}
}

// countingServer registers one destructive tool and reports how many times it
// actually executed, so confirmation tests can prove the first round-trip has
// no side effect.
func countingServer(calls *int) *Server {
	reg := NewRegistry()
	reg.RegisterTool(ToolDefinition{
		Name:        "delete_thing",
		Description: "delete a thing",
		InputSchema: map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string"}}},
		Annotations: destructiveHints(),
	}, func(_ context.Context, _ map[string]any) (*ToolCallResult, error) {
		*calls = *calls + 1
		return &ToolCallResult{Content: []ContentItem{{Type: "text", Text: "deleted"}}}, nil
	}, authz.ScopeEchoWrite)
	return NewServer(reg)
}

func destructiveViewer() viewer.Context {
	return viewer.NewUserViewerWithToken(
		"test-user", "access", []string{authz.ScopeEchoWrite}, []string{authz.AudienceMCPRemote}, "test-jti")
}

func TestDestructiveToolAsksForConfirmation(t *testing.T) {
	calls := 0
	srv := countingServer(&calls)

	_, resp := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", map[string]any{
		"name":      "delete_thing",
		"arguments": map[string]any{"id": "abc"},
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if calls != 0 {
		t.Fatalf("tool executed before confirmation")
	}

	result := unmarshalResult[InputRequiredResult](t, resp)
	if result.ResultType != resultTypeInputRequired {
		t.Errorf("resultType = %q, want %q", result.ResultType, resultTypeInputRequired)
	}
	confirm, ok := result.InputRequests[confirmationKey]
	if !ok || confirm.Method != confirmationMethod {
		t.Fatalf("confirm input request = %+v", result.InputRequests)
	}
	if result.RequestState == "" {
		t.Error("requestState is empty")
	}
	if _, ok := result.Meta[metaKeyServerInfo]; !ok {
		t.Error("input_required result is missing serverInfo metadata")
	}
	fields := resultFields(t, resp)
	if _, ok := fields["ttlMs"]; ok {
		t.Error("input_required result must not carry cache hints")
	}
}

func TestConfirmationAcceptExecutesOnce(t *testing.T) {
	calls := 0
	srv := countingServer(&calls)

	_, first := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", map[string]any{
		"name":      "delete_thing",
		"arguments": map[string]any{"id": "abc"},
	})
	pending := unmarshalResult[InputRequiredResult](t, first)

	_, second := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", map[string]any{
		"name":         "delete_thing",
		"arguments":    map[string]any{"id": "abc"},
		"requestState": pending.RequestState,
		"inputResponses": map[string]any{
			confirmationKey: map[string]any{
				"action":  "accept",
				"content": map[string]any{confirmationKey: true},
			},
		},
	})
	if second.Error != nil {
		t.Fatalf("unexpected error: %v", second.Error)
	}
	if calls != 1 {
		t.Fatalf("tool calls = %d, want 1", calls)
	}
	result := unmarshalResult[ToolCallResult](t, second)
	if result.IsError {
		t.Errorf("tool result is an error: %+v", result)
	}
}

func TestConfirmationDeclineDoesNotExecute(t *testing.T) {
	calls := 0
	srv := countingServer(&calls)

	_, first := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", map[string]any{
		"name":      "delete_thing",
		"arguments": map[string]any{"id": "abc"},
	})
	pending := unmarshalResult[InputRequiredResult](t, first)

	_, second := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", map[string]any{
		"name":         "delete_thing",
		"arguments":    map[string]any{"id": "abc"},
		"requestState": pending.RequestState,
		"inputResponses": map[string]any{
			confirmationKey: map[string]any{"action": "decline"},
		},
	})
	if second.Error != nil {
		t.Fatalf("unexpected error: %v", second.Error)
	}
	if calls != 0 {
		t.Fatal("declined tool still executed")
	}
	result := unmarshalResult[ToolCallResult](t, second)
	if !result.IsError {
		t.Errorf("declined result should be flagged as an error for the model: %+v", result)
	}
}

func TestConfirmationStateRejectsTampering(t *testing.T) {
	calls := 0
	srv := countingServer(&calls)

	_, first := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", map[string]any{
		"name":      "delete_thing",
		"arguments": map[string]any{"id": "abc"},
	})
	pending := unmarshalResult[InputRequiredResult](t, first)

	accept := map[string]any{
		confirmationKey: map[string]any{
			"action":  "accept",
			"content": map[string]any{confirmationKey: true},
		},
	}
	for name, params := range map[string]map[string]any{
		"swapped target": {
			"name":           "delete_thing",
			"arguments":      map[string]any{"id": "other"},
			"requestState":   pending.RequestState,
			"inputResponses": accept,
		},
		"tampered state": {
			"name":           "delete_thing",
			"arguments":      map[string]any{"id": "abc"},
			"requestState":   pending.RequestState + "x",
			"inputResponses": accept,
		},
	} {
		t.Run(name, func(t *testing.T) {
			rec, resp := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", params)
			if resp.Error == nil || resp.Error.Code != ErrCodeInvalidParams {
				t.Fatalf("error = %+v, want invalid params", resp.Error)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}
		})
	}
	if calls != 0 {
		t.Fatalf("tampered requests executed the tool %d times", calls)
	}
}

func TestConfirmationStateIsSingleUse(t *testing.T) {
	calls := 0
	srv := countingServer(&calls)

	_, first := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", map[string]any{
		"name":      "delete_thing",
		"arguments": map[string]any{"id": "abc"},
	})
	pending := unmarshalResult[InputRequiredResult](t, first)

	retry := map[string]any{
		"name":         "delete_thing",
		"arguments":    map[string]any{"id": "abc"},
		"requestState": pending.RequestState,
		"inputResponses": map[string]any{
			confirmationKey: map[string]any{
				"action":  "accept",
				"content": map[string]any{confirmationKey: true},
			},
		},
	}
	if _, resp := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", retry); resp.Error != nil {
		t.Fatalf("first retry failed: %v", resp.Error)
	}
	if calls != 1 {
		t.Fatalf("tool calls = %d, want 1", calls)
	}

	_, resp := doModernAs(t, srv, destructiveViewer(), elicitationCaps(), "tools/call", retry)
	if resp.Error == nil || resp.Error.Code != ErrCodeInvalidParams {
		t.Fatalf("replayed confirmation = %+v, want invalid params", resp.Error)
	}
	if calls != 1 {
		t.Fatalf("replayed confirmation executed the tool again")
	}
}

func TestDestructiveToolWithoutElicitationExecutes(t *testing.T) {
	calls := 0
	srv := countingServer(&calls)

	_, resp := doModernAs(t, srv, destructiveViewer(), nil, "tools/call", map[string]any{
		"name":      "delete_thing",
		"arguments": map[string]any{"id": "abc"},
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if calls != 1 {
		t.Fatalf("tool calls = %d, want 1", calls)
	}
}

func TestDestructiveToolLegacySkipsConfirmation(t *testing.T) {
	calls := 0
	srv := countingServer(&calls)

	req := testRequest(t, http.MethodPost,
		`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"delete_thing","arguments":{"id":"abc"}}}`)
	req = req.WithContext(viewer.WithContext(req.Context(), destructiveViewer()))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var resp Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	if calls != 1 {
		t.Fatalf("tool calls = %d, want 1", calls)
	}
}
