// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/danielgtaylor/huma/v2"
)

// stubRegistry serves prebuilt component schemas without a real type registry.
type stubRegistry struct {
	schemas map[string]*huma.Schema
}

func (registry *stubRegistry) Schema(reflect.Type, bool, string) *huma.Schema { return nil }
func (registry *stubRegistry) SchemaFromRef(string) *huma.Schema              { return nil }
func (registry *stubRegistry) TypeFromRef(string) reflect.Type                { return nil }
func (registry *stubRegistry) RegisterTypeAlias(reflect.Type, reflect.Type)   {}
func (registry *stubRegistry) Map() map[string]*huma.Schema                   { return registry.schemas }

func testDocument() *huma.OpenAPI {
	components := &stubRegistry{schemas: map[string]*huma.Schema{
		"EchoUpsertDto": {
			Type: "object",
			Properties: map[string]*huma.Schema{
				"content": {Type: "string"},
				"private": {Type: "boolean"},
				"tags":    {Type: "array", Items: &huma.Schema{Ref: "#/components/schemas/Tag"}},
			},
			Required: []string{"content"},
		},
		"Tag": {
			Type:       "object",
			Properties: map[string]*huma.Schema{"name": {Type: "string"}},
		},
	}}

	return &huma.OpenAPI{
		Components: &huma.Components{Schemas: components},
		Paths: map[string]*huma.PathItem{
			"/echo": {
				Post: &huma.Operation{
					OperationID:   "echo-create",
					Summary:       "创建新的 Echo",
					DefaultStatus: 200,
					Security:      []map[string][]string{{bearerScheme: {"echo:write"}}},
					RequestBody: &huma.RequestBody{
						Required: true,
						Content: map[string]*huma.MediaType{
							"application/json": {Schema: &huma.Schema{Ref: "#/components/schemas/EchoUpsertDto"}},
						},
					},
					Responses: map[string]*huma.Response{
						"200": {Content: map[string]*huma.MediaType{"application/json": {}}},
					},
				},
			},
			"/echo/{id}": {
				Get: &huma.Operation{
					OperationID:   "echo-get",
					Summary:       "获取指定 ID 的 Echo",
					DefaultStatus: 200,
					Security:      []map[string][]string{{bearerScheme: {"echo:read"}}},
					Parameters: []*huma.Param{
						{Name: "id", In: "path", Required: true, Schema: &huma.Schema{Type: "string", Format: "uuid"}},
					},
					Responses: map[string]*huma.Response{
						"200": {Content: map[string]*huma.MediaType{"application/json": {}}},
					},
				},
				Delete: &huma.Operation{
					OperationID:   "echo-delete",
					DefaultStatus: 204,
					Security:      []map[string][]string{{bearerScheme: {"echo:write"}}},
				},
			},
			"/chat/answer": {
				Post: &huma.Operation{
					OperationID:   "chat-answer",
					DefaultStatus: 200,
					Responses: map[string]*huma.Response{
						"200": {Content: map[string]*huma.MediaType{"text/event-stream": {}}},
					},
				},
			},
		},
	}
}

func TestBuild_Operations(t *testing.T) {
	doc, body, fingerprint, err := Build(testDocument())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if doc.Ops["echo.create"].Method != "POST" || doc.Ops["echo.create"].Path != "/echo" {
		t.Fatalf("unexpected create op: %+v", doc.Ops["echo.create"])
	}
	if doc.Ops["echo.delete"].Effect != EffectDestructive {
		t.Fatalf("delete effect = %q", doc.Ops["echo.delete"].Effect)
	}
	if doc.Ops["chat.answer"].Kind != KindStream {
		t.Fatalf("stream kind = %q", doc.Ops["chat.answer"].Kind)
	}
	if got := doc.Ops["echo.create"].Scopes; len(got) != 1 || got[0] != "echo:write" {
		t.Fatalf("scopes = %v", got)
	}

	create := doc.Ops["echo.create"]
	if create.Bind["content"] != BindBody || create.Bind["private"] != BindBody {
		t.Fatalf("body bind = %v", create.Bind)
	}
	properties, ok := create.Input["properties"].(map[string]any)
	if !ok {
		t.Fatalf("input properties missing: %v", create.Input)
	}
	if _, ok := properties["content"]; !ok {
		t.Fatalf("content missing from flattened input")
	}
	required, ok := create.Input["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "content" {
		t.Fatalf("required = %v", create.Input["required"])
	}

	// $ref inside an array item must be inlined too.
	tags, _ := properties["tags"].(map[string]any)
	items, _ := tags["items"].(map[string]any)
	if _, ok := items["properties"].(map[string]any)["name"]; !ok {
		t.Fatalf("nested ref not inlined: %v", tags)
	}

	get := doc.Ops["echo.get"]
	if get.Bind["id"] != BindPath || get.Effect != EffectRead {
		t.Fatalf("get op = %+v", get)
	}

	sum := sha256.Sum256(body)
	if fingerprint != hex.EncodeToString(sum[:]) {
		t.Fatalf("fingerprint mismatch")
	}
}

func TestBuild_StableFingerprint(t *testing.T) {
	_, first, firstFingerprint, err := Build(testDocument())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	_, second, secondFingerprint, err := Build(testDocument())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if firstFingerprint != secondFingerprint {
		t.Fatalf("fingerprint not stable: %s vs %s", firstFingerprint, secondFingerprint)
	}
	if string(first) != string(second) {
		t.Fatalf("canonical bytes differ")
	}

	var decoded Doc
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatalf("catalog is not valid JSON: %v", err)
	}
	if len(decoded.Ops) != 4+len(RawOps()) {
		t.Fatalf("ops = %d", len(decoded.Ops))
	}
}

func TestBuild_NilDocument(t *testing.T) {
	if _, _, _, err := Build(nil); err == nil {
		t.Fatal("expected an error for a nil document")
	}
}

func TestBuild_RawOpsCarryTransport(t *testing.T) {
	doc, _, _, err := Build(testDocument())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	logs := doc.Ops["system.logs.stream"]
	if logs.Kind != KindStream || logs.Transport == nil || logs.Transport.Response == nil {
		t.Fatalf("system.logs.stream = %+v", logs)
	}
	if logs.Transport.Response.ContentType != ContentTypeSSE {
		t.Fatalf("logs content type = %q", logs.Transport.Response.ContentType)
	}
	if logs.Transport.Request == nil || logs.Transport.Request.Auth != AuthQuery {
		t.Fatalf("logs must authenticate via query: %+v", logs.Transport.Request)
	}
	if _, ok := doc.Ops["chat.ask"]; ok {
		t.Fatal("chat.ask must stay out of the catalog while the CLI does not support Copilot")
	}

	upload := doc.Ops["file.upload"]
	if upload.Bind["file"] != BindForm {
		t.Fatalf("file.upload bind = %v", upload.Bind)
	}
	properties := upload.Input["properties"].(map[string]any)
	if properties["file"].(map[string]any)["format"] != "binary" {
		t.Fatalf("file part not marked binary: %v", properties["file"])
	}

	ws := doc.Ops["ws.system.logs"]
	if ws.Mount != MountWS || ws.Kind != KindWebSocket {
		t.Fatalf("ws op = %+v", ws)
	}
	if ws.Transport == nil || ws.Transport.Request == nil || ws.Transport.Request.Auth != AuthQuery {
		t.Fatalf("ws transport = %+v", ws.Transport)
	}
}

func TestBuild_NoNullCollections(t *testing.T) {
	_, body, _, err := Build(testDocument())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	var decoded struct {
		Ops map[string]struct {
			Scopes *[]string          `json:"scopes"`
			Input  *map[string]any    `json:"input"`
			Bind   *map[string]string `json:"bind"`
		} `json:"ops"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for id, op := range decoded.Ops {
		if op.Scopes == nil {
			t.Errorf("%s: scopes is null", id)
		}
		if op.Input == nil {
			t.Errorf("%s: input is null", id)
		}
		if op.Bind == nil {
			t.Errorf("%s: bind is null", id)
		}
	}
}

func TestRawOpsValidate(t *testing.T) {
	for id, op := range RawOps() {
		if err := validateOp(op); err != nil {
			t.Errorf("%s: %v", id, err)
		}
	}
}
