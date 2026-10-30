// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package catalog turns the Huma OpenAPI document into the operation catalog the
// CLI discovers at run time: one flat input schema per operation plus where each
// field goes. The document is canonicalised (sorted keys, compact JSON) and
// fingerprinted with sha256 so clients can tell when their cache is stale.
//
// Design reference: docs/dev/cli-design.md §"catalog 驱动的命令发现".
package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

const (
	// HeaderName carries the catalog fingerprint on responses and requests.
	HeaderName = "X-Ech0-Catalog"
	// RoutePath is mounted under the API base (`/api/_catalog`).
	RoutePath = "/_catalog"
	// StaleErrorCode is returned with 412 when a request carries an outdated fingerprint.
	StaleErrorCode = "catalog_stale"

	bearerScheme    = "bearerAuth"
	schemaRefPrefix = "#/components/schemas/"
)

// Bind says where an input field travels.
const (
	BindPath  = "path"
	BindQuery = "query"
	BindBody  = "body"
	// BindForm carries the field as a multipart/form-data part.
	BindForm = "form"
)

// Effect is the operation's blast radius, used by clients for confirmation prompts.
const (
	EffectRead        = "read"
	EffectWrite       = "write"
	EffectDestructive = "destructive"
)

// Kind is how the client renders the success body.
const (
	KindJSON      = "json"
	KindStream    = "stream"
	KindBinary    = "binary"
	KindWebSocket = "websocket"
)

// Mount says which router prefix the path hangs under. Empty means the API
// base (`/api`); "ws" marks the unauthenticated-by-middleware websocket group.
const (
	MountAPI = "api"
	MountWS  = "ws"
)

// Content types the catalog can name. Only non-JSON ones need to appear.
const (
	ContentTypeJSON   = "application/json"
	ContentTypeSSE    = "text/event-stream"
	ContentFormData   = "multipart/form-data"
	ContentTypeBinary = "application/octet-stream"
)

// Op is one callable operation.
type Op struct {
	Summary    string            `json:"summary"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Mount      string            `json:"mount,omitempty"`
	Scopes     []string          `json:"scopes"`
	Effect     string            `json:"effect"`
	Kind       string            `json:"kind"`
	Input      map[string]any    `json:"input"`
	Bind       map[string]string `json:"bind"`
	Transport  *Transport        `json:"transport,omitempty"`
	Deprecated bool              `json:"deprecated"`
}

// Transport describes everything that is not a JSON body in and a JSON envelope
// out. It is omitted for plain operations, so older clients keep the short form.
type Transport struct {
	Request  *RequestTransport  `json:"request,omitempty"`
	Response *ResponseTransport `json:"response,omitempty"`
}

// RequestTransport names the request encoding; multipart parts themselves are
// described by Bind==BindForm plus `format: binary` on file fields. Auth says
// how the credential travels when the Authorization header cannot carry it.
type RequestTransport struct {
	ContentType string `json:"content_type,omitempty"`
	Auth        string `json:"auth,omitempty"`
}

// ResponseTransport names the success body encoding. Events is an advisory SSE
// vocabulary: clients render names they know and fall back to the raw frame.
type ResponseTransport struct {
	Kind        string     `json:"kind"`
	ContentType string     `json:"content_type,omitempty"`
	Filename    string     `json:"filename,omitempty"`
	Events      []SSEEvent `json:"events,omitempty"`
}

// SSEEvent is one named event a stream may emit.
type SSEEvent struct {
	Name string            `json:"name"`
	Data map[string]string `json:"data,omitempty"`
}

// Doc is the catalog document. Keys of Ops are dotted ids (`echo.create`).
type Doc struct {
	Ops map[string]Op `json:"ops"`
}

// Build derives the catalog from the OpenAPI document and returns it together
// with its canonical bytes and sha256 fingerprint.
func Build(document *huma.OpenAPI) (Doc, []byte, string, error) {
	if document == nil {
		return Doc{}, nil, "", fmt.Errorf("catalog: nil OpenAPI document")
	}

	components := componentSchemas(document)
	doc := Doc{Ops: make(map[string]Op, len(document.Paths))}

	for path, item := range document.Paths {
		if item == nil {
			continue
		}
		for _, entry := range pathOperations(item) {
			op := entry.op
			if op == nil || op.Hidden {
				continue
			}
			id := opID(op, entry.method, path)
			if _, clash := doc.Ops[id]; clash {
				id = id + "." + strings.ToLower(entry.method)
			}
			doc.Ops[id] = buildOp(op, entry.method, path, components)
		}
	}

	if err := mergeRawOps(&doc); err != nil {
		return Doc{}, nil, "", err
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return Doc{}, nil, "", fmt.Errorf("catalog: marshal document: %w", err)
	}
	sum := sha256.Sum256(body)
	return doc, body, hex.EncodeToString(sum[:]), nil
}

type operationEntry struct {
	method string
	op     *huma.Operation
}

func pathOperations(item *huma.PathItem) []operationEntry {
	return []operationEntry{
		{method: "GET", op: item.Get},
		{method: "PUT", op: item.Put},
		{method: "POST", op: item.Post},
		{method: "DELETE", op: item.Delete},
		{method: "PATCH", op: item.Patch},
	}
}

func opID(op *huma.Operation, method, path string) string {
	if op.OperationID != "" {
		return strings.ReplaceAll(op.OperationID, "-", ".")
	}
	return strings.ToLower(method) + strings.ReplaceAll(path, "/", ".")
}

func buildOp(op *huma.Operation, method, path string, components map[string]map[string]any) Op {
	properties := make(map[string]any)
	bind := make(map[string]string)
	var required []string

	addField := func(name string, schema map[string]any, location string, isRequired bool) {
		if _, exists := properties[name]; exists {
			return // path/query wins over a body field with the same name
		}
		properties[name] = schema
		bind[name] = location
		if isRequired {
			required = append(required, name)
		}
	}

	for _, parameter := range op.Parameters {
		if parameter == nil || parameter.Schema == nil {
			continue
		}
		switch parameter.In {
		case "path":
			addField(parameter.Name, inline(schemaJSON(parameter.Schema), components, map[string]bool{}), BindPath, true)
		case "query":
			addField(parameter.Name, inline(schemaJSON(parameter.Schema), components, map[string]bool{}), BindQuery, parameter.Required)
		}
	}

	if op.RequestBody != nil {
		if media := op.RequestBody.Content["application/json"]; media != nil && media.Schema != nil {
			body := inline(schemaJSON(media.Schema), components, map[string]bool{})
			if bodyProperties, ok := body["properties"].(map[string]any); ok {
				bodyRequired := requiredNames(body["required"])
				for name, sub := range bodyProperties {
					subSchema, _ := sub.(map[string]any)
					addField(name, subSchema, BindBody, bodyRequired[name])
				}
			} else {
				addField("body", body, BindBody, op.RequestBody.Required)
			}
		}
	}

	input := map[string]any{"type": "object", "properties": properties}
	if len(required) > 0 {
		input["required"] = sortedUnique(required)
	}

	kind := kindOf(op)
	return Op{
		Summary:    op.Summary,
		Method:     method,
		Path:       path,
		Scopes:     scopesOf(op),
		Effect:     effectOf(method),
		Kind:       kind,
		Input:      input,
		Bind:       bind,
		Transport:  transportOf(op, kind),
		Deprecated: op.Deprecated,
	}
}

func effectOf(method string) string {
	switch method {
	case "GET", "HEAD":
		return EffectRead
	case "DELETE":
		return EffectDestructive
	default:
		return EffectWrite
	}
}

func kindOf(op *huma.Operation) string {
	response := op.Responses[fmt.Sprintf("%d", op.DefaultStatus)]
	if response == nil {
		response = op.Responses["200"]
	}
	if response == nil {
		return KindJSON
	}
	switch {
	case response.Content["text/event-stream"] != nil:
		return KindStream
	case response.Content["application/json"] != nil:
		return KindJSON
	case len(response.Content) > 0:
		return KindBinary
	default:
		return KindJSON
	}
}

// transportOf fills the transport block for the rare Huma operation that is not
// plain JSON in / JSON out. Hand-written raw operations carry their own block.
func transportOf(op *huma.Operation, kind string) *Transport {
	requestType := ""
	if op.RequestBody != nil {
		for contentType := range op.RequestBody.Content {
			if contentType != ContentTypeJSON {
				requestType = contentType
				break
			}
		}
	}
	if kind == KindJSON && requestType == "" {
		return nil
	}

	transport := &Transport{}
	if requestType != "" {
		transport.Request = &RequestTransport{ContentType: requestType}
	}
	if kind != KindJSON {
		response := &ResponseTransport{Kind: kind}
		if media := responseMediaType(op); media != "" {
			response.ContentType = media
		}
		transport.Response = response
	}
	return transport
}

func responseMediaType(op *huma.Operation) string {
	response := op.Responses[fmt.Sprintf("%d", op.DefaultStatus)]
	if response == nil {
		response = op.Responses["200"]
	}
	if response == nil {
		return ""
	}
	if response.Content[ContentTypeSSE] != nil {
		return ContentTypeSSE
	}
	for contentType := range response.Content {
		if contentType != ContentTypeJSON {
			return contentType
		}
	}
	return ContentTypeJSON
}

func scopesOf(op *huma.Operation) []string {
	var scopes []string
	for _, requirement := range op.Security {
		scopes = append(scopes, requirement[bearerScheme]...)
	}
	return sortedUnique(scopes)
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func requiredNames(raw any) map[string]bool {
	names := make(map[string]bool)
	list, ok := raw.([]any)
	if !ok {
		return names
	}
	for _, item := range list {
		if name, ok := item.(string); ok {
			names[name] = true
		}
	}
	return names
}

func componentSchemas(document *huma.OpenAPI) map[string]map[string]any {
	if document.Components == nil || document.Components.Schemas == nil {
		return map[string]map[string]any{}
	}
	registry := document.Components.Schemas.Map()
	out := make(map[string]map[string]any, len(registry))
	for name, schema := range registry {
		out[name] = schemaJSON(schema)
	}
	return out
}

func schemaJSON(schema *huma.Schema) map[string]any {
	if schema == nil {
		return map[string]any{}
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

// inline replaces `$ref`s with the referenced component schema. Recursive
// schemas keep the reference once the cycle closes.
func inline(node map[string]any, components map[string]map[string]any, stack map[string]bool) map[string]any {
	if ref, ok := node["$ref"].(string); ok {
		name := strings.TrimPrefix(ref, schemaRefPrefix)
		target, found := components[name]
		if found && !stack[name] {
			stack[name] = true
			merged := inline(cloneMap(target), components, stack)
			delete(stack, name)
			for key, value := range node {
				if key != "$ref" {
					merged[key] = inlineValue(value, components, stack)
				}
			}
			return merged
		}
	}

	out := make(map[string]any, len(node))
	for key, value := range node {
		if key == "$defs" {
			continue
		}
		out[key] = inlineValue(value, components, stack)
	}
	return out
}

func inlineValue(value any, components map[string]map[string]any, stack map[string]bool) any {
	switch typed := value.(type) {
	case map[string]any:
		return inline(typed, components, stack)
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = inlineValue(item, components, stack)
		}
		return out
	default:
		return value
	}
}

func cloneMap(source map[string]any) map[string]any {
	out := make(map[string]any, len(source))
	for key, value := range source {
		out[key] = value
	}
	return out
}
