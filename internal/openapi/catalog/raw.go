// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package catalog

import (
	"fmt"
	"strings"
)

// rawOps are the operations that live on bare gin, so the Huma document cannot
// describe them: SSE streams, multipart uploads, binary downloads and the
// websocket the dashboard subscribes to. They carry their own transport block
// because there is no OpenAPI content map to read it from.
//
// Drift protection lives in internal/router: a test walks the gin route table
// and fails unless every /api and /ws route is described here or by Huma.
var rawOps = map[string]Op{
	"system.logs.stream": {
		Summary: "订阅系统日志（SSE，data 为信封）",
		Method:  "GET",
		Path:    "/system/logs/stream",
		Scopes:  []string{"admin:settings"},
		Effect:  EffectRead,
		Kind:    KindStream,
		Input: objectSchema(
			nil,
			map[string]any{"level": stringSchema(), "keyword": stringSchema()},
		),
		Bind: map[string]string{"level": BindQuery, "keyword": BindQuery},
		Transport: &Transport{
			// The handler re-reads `?token=` itself, so the CLI must pass it even
			// though the middleware would also accept the Authorization header.
			Request:  &RequestTransport{Auth: AuthQuery},
			Response: &ResponseTransport{Kind: KindStream, ContentType: ContentTypeSSE},
		},
	},
	"ws.system.logs": {
		Summary: "订阅系统日志（WebSocket，token 走查询参数）",
		Method:  "GET",
		Path:    "/system/logs",
		Mount:   MountWS,
		Effect:  EffectRead,
		Kind:    KindWebSocket,
		Input: objectSchema(
			nil,
			map[string]any{"level": stringSchema(), "keyword": stringSchema()},
		),
		Bind: map[string]string{"level": BindQuery, "keyword": BindQuery},
		Transport: &Transport{
			Request:  &RequestTransport{Auth: AuthQuery},
			Response: &ResponseTransport{Kind: KindWebSocket},
		},
	},
	"file.upload": {
		Summary: "上传文件（multipart）",
		Method:  "POST",
		Path:    "/files/upload",
		Scopes:  []string{"file:write"},
		Effect:  EffectWrite,
		Kind:    KindJSON,
		Input: objectSchema(
			[]string{"file", "category", "storage_type"},
			map[string]any{
				"file":         binarySchema(),
				"category":     stringSchema(),
				"storage_type": stringSchema(),
			},
		),
		Bind: map[string]string{"file": BindForm, "category": BindForm, "storage_type": BindForm},
		Transport: &Transport{
			Request: &RequestTransport{ContentType: ContentFormData},
		},
	},
	"migration.upload": {
		Summary: "上传迁移源文件（multipart zip）",
		Method:  "POST",
		Path:    "/migration/upload",
		Scopes:  []string{"admin:settings"},
		Effect:  EffectWrite,
		Kind:    KindJSON,
		Input: objectSchema(
			[]string{"file", "source_type"},
			map[string]any{"file": binarySchema(), "source_type": stringSchema()},
		),
		Bind: map[string]string{"file": BindForm, "source_type": BindForm},
		Transport: &Transport{
			Request: &RequestTransport{ContentType: ContentFormData},
		},
	},
	"file.stream": {
		Summary: "按存储路径读取文件内容（二进制）",
		Method:  "GET",
		Path:    "/file/stream",
		Scopes:  []string{"file:read"},
		Effect:  EffectRead,
		Kind:    KindBinary,
		Input: objectSchema(
			[]string{"path", "storage_type"},
			map[string]any{
				"path":         stringSchema(),
				"storage_type": stringSchema(),
				"name":         stringSchema(),
				"content_type": stringSchema(),
			},
		),
		Bind: map[string]string{
			"path": BindQuery, "storage_type": BindQuery,
			"name": BindQuery, "content_type": BindQuery,
		},
		Transport: &Transport{Response: &ResponseTransport{Kind: KindBinary}},
	},
	"file.stream.by_id": {
		Summary:   "按文件 ID 读取内容（二进制）",
		Method:    "GET",
		Path:      "/file/{id}/stream",
		Scopes:    []string{"file:read"},
		Effect:    EffectRead,
		Kind:      KindBinary,
		Input:     objectSchema([]string{"id"}, map[string]any{"id": stringSchema()}),
		Bind:      map[string]string{"id": BindPath},
		Transport: &Transport{Response: &ResponseTransport{Kind: KindBinary}},
	},
	"migration.export.download": {
		Summary: "下载当前快照（zip）",
		Method:  "GET",
		Path:    "/migration/export/download",
		Scopes:  []string{"admin:settings"},
		Effect:  EffectRead,
		Kind:    KindBinary,
		Input:   objectSchema(nil, map[string]any{"format": stringSchema()}),
		Bind:    map[string]string{"format": BindQuery},
		Transport: &Transport{
			Response: &ResponseTransport{Kind: KindBinary, ContentType: "application/zip"},
		},
	},
}

// AuthQuery marks transports that authenticate with a `token` query parameter
// instead of the Authorization header (websocket handshakes cannot carry it).
const AuthQuery = "query"

// RawOps returns a copy of the hand-written operations, for the drift test and
// for documentation tooling.
func RawOps() map[string]Op {
	out := make(map[string]Op, len(rawOps))
	for id, op := range rawOps {
		out[id] = op
	}
	return out
}

func mergeRawOps(doc *Doc) error {
	seen := make(map[string]string, len(doc.Ops)+len(rawOps))
	for id, op := range doc.Ops {
		seen[routeKey(op)] = id
	}
	for id, op := range rawOps {
		if other, clash := doc.Ops[id]; clash {
			return fmt.Errorf("catalog: raw op %q clashes with %s %s", id, other.Method, other.Path)
		}
		if other, clash := seen[routeKey(op)]; clash {
			return fmt.Errorf("catalog: raw op %q clashes with %s on %s %s", id, other, op.Method, op.Path)
		}
		if err := validateOp(op); err != nil {
			return fmt.Errorf("catalog: raw op %q: %w", id, err)
		}
		if op.Scopes == nil {
			op.Scopes = []string{}
		}
		doc.Ops[id] = op
		seen[routeKey(op)] = id
	}
	return nil
}

// routeKey normalises path parameters so the hand-written registration and the
// gin route table can be compared.
func routeKey(op Op) string {
	mount := op.Mount
	if mount == "" {
		mount = MountAPI
	}
	return mount + " " + strings.ToUpper(op.Method) + " " + normalizePath(op.Path)
}

func normalizePath(path string) string {
	segments := strings.Split(path, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ":") || strings.HasPrefix(segment, "*") {
			segments[index] = "{" + segment[1:] + "}"
		}
	}
	return strings.Join(segments, "/")
}

func validateOp(op Op) error {
	if op.Method == "" {
		return fmt.Errorf("method is required")
	}
	if !strings.HasPrefix(op.Path, "/") {
		return fmt.Errorf("path %q must start with /", op.Path)
	}
	switch op.Kind {
	case KindJSON, KindStream, KindBinary, KindWebSocket:
	default:
		return fmt.Errorf("unknown kind %q", op.Kind)
	}
	switch op.Mount {
	case "", MountAPI, MountWS:
	default:
		return fmt.Errorf("unknown mount %q", op.Mount)
	}

	properties, _ := op.Input["properties"].(map[string]any)
	if len(op.Bind) != len(properties) {
		return fmt.Errorf("bind covers %d fields, schema has %d", len(op.Bind), len(properties))
	}
	for name, location := range op.Bind {
		if _, ok := properties[name]; !ok {
			return fmt.Errorf("bind names unknown field %q", name)
		}
		switch location {
		case BindPath, BindQuery, BindBody, BindForm:
		default:
			return fmt.Errorf("field %q has unknown bind %q", name, location)
		}
		if schema, ok := properties[name].(map[string]any); ok && schema["format"] == "binary" && location != BindForm {
			return fmt.Errorf("file field %q must bind to form", name)
		}
	}

	if op.Transport != nil && op.Transport.Response != nil && op.Transport.Response.Kind != op.Kind {
		return fmt.Errorf("transport response kind %q != op kind %q", op.Transport.Response.Kind, op.Kind)
	}
	if op.Transport != nil && op.Transport.Request != nil {
		switch op.Transport.Request.Auth {
		case "", AuthQuery:
		default:
			return fmt.Errorf("unknown request auth %q", op.Transport.Request.Auth)
		}
	}
	return nil
}

func objectSchema(required []string, properties map[string]any) map[string]any {
	schema := map[string]any{"type": "object", "properties": properties}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringSchema() map[string]any {
	return map[string]any{"type": "string"}
}

// binarySchema marks a multipart file part: the CLI turns a filesystem path
// into an upload instead of a text field.
func binarySchema() map[string]any {
	return map[string]any{"type": "string", "format": "binary"}
}
