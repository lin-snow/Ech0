// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	versionPkg "github.com/lin-snow/ech0/internal/version"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const toolTimeout = 10 * time.Second

type ctxKey int

const (
	ctxKeyRawToken ctxKey = iota
	ctxKeyBaseURL
)

func RawTokenFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyRawToken).(string)
	return v
}

func BaseURLFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyBaseURL).(string)
	return v
}

type Server struct {
	registry *Registry
}

func NewServer(registry *Registry) *Server {
	return &Server{registry: registry}
}

func serverInfo() ServerInfo {
	return ServerInfo{Name: ServerName, Version: versionPkg.Version}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	s.handlePost(w, r)
}

func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	body, err := io.ReadAll(io.LimitReader(r.Body, 256*1024))
	if err != nil {
		writeRPCError(w, nil, &RPCError{Code: ErrCodeParse, Message: "failed to read request body"})
		return
	}

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		writeRPCError(w, nil, &RPCError{Code: ErrCodeParse, Message: "invalid JSON"})
		return
	}
	if req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, &RPCError{Code: ErrCodeInvalidRequest, Message: "jsonrpc must be 2.0"})
		return
	}

	if len(req.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if string(req.ID) == "null" {
		writeRPCError(w, nil, &RPCError{Code: ErrCodeInvalidRequest, Message: "request id must be a string or number"})
		return
	}

	ctx := r.Context()
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		ctx = context.WithValue(ctx, ctxKeyRawToken, strings.TrimPrefix(auth, "Bearer "))
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	ctx = context.WithValue(ctx, ctxKeyBaseURL, scheme+"://"+r.Host)
	r = r.WithContext(ctx)

	v := viewer.MustFromContext(ctx)

	result, rpcErr := s.dispatch(r, &req, v)

	logUtil.GetLogger().Info("mcp_request",
		slog.String("method", req.Method),
		slog.String("user_id", v.UserID()),
		slog.String("token_id", v.TokenID()),
		slog.Duration("latency", time.Since(start)),
		slog.Bool("error", rpcErr != nil),
	)

	if rpcErr != nil {
		writeRPCError(w, req.ID, rpcErr)
		return
	}
	if c, ok := result.(completer); ok {
		c.complete(serverInfo())
	}
	writeRPCResult(w, req.ID, result)
}

type requestParams struct {
	Meta map[string]any `json:"_meta"`
	Name string         `json:"name"`
	URI  string         `json:"uri"`
}

func (s *Server) dispatch(r *http.Request, req *Request, v viewer.Context) (any, *RPCError) {
	if req.Method == "initialize" {
		return nil, &RPCError{
			Code: ErrCodeMethodNotFound,
			Message: "the initialize handshake was removed in MCP 2026-07-28; " +
				"this server only speaks stateless protocol versions: " + strings.Join(SupportedVersions, ", "),
			Data: map[string]any{"supported": SupportedVersions},
		}
	}

	var params requestParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &RPCError{Code: ErrCodeInvalidRequest, Message: "params must be an object"}
		}
	}
	if rpcErr := validateTransport(r, req.Method, &params); rpcErr != nil {
		return nil, rpcErr
	}

	switch req.Method {
	case "server/discover":
		return s.handleDiscover(), nil
	case "tools/list":
		return s.handleToolsList()
	case "tools/call":
		return s.handleToolsCall(r, req, v)
	case "resources/list":
		return s.handleResourcesList()
	case "resources/read":
		return s.handleResourcesRead(r, req, v)
	default:
		return nil, &RPCError{Code: ErrCodeMethodNotFound, Message: fmt.Sprintf("method %q not found", req.Method)}
	}
}

func validateTransport(r *http.Request, method string, params *requestParams) *RPCError {
	headerVersion := r.Header.Get("Mcp-Protocol-Version")
	if headerVersion == "" {
		return headerMismatch("required header MCP-Protocol-Version is missing")
	}
	bodyVersion, _ := params.Meta[metaKeyProtocolVersion].(string)
	if bodyVersion == "" {
		return headerMismatch("params._meta is missing " + metaKeyProtocolVersion)
	}
	if headerVersion != bodyVersion {
		return headerMismatch(fmt.Sprintf("MCP-Protocol-Version header %q does not match body value %q", headerVersion, bodyVersion))
	}
	if bodyVersion != ProtocolVersion {
		return &RPCError{
			Code:    ErrCodeUnsupportedProtocolVersion,
			Message: "Unsupported protocol version",
			Data:    map[string]any{"supported": SupportedVersions, "requested": bodyVersion},
		}
	}

	headerMethod := r.Header.Get("Mcp-Method")
	if headerMethod == "" {
		return headerMismatch("required header Mcp-Method is missing")
	}
	if headerMethod != method {
		return headerMismatch(fmt.Sprintf("Mcp-Method header %q does not match body method %q", headerMethod, method))
	}

	if method != "tools/call" && method != "resources/read" {
		return nil
	}
	bodyName := params.Name
	if method == "resources/read" {
		bodyName = params.URI
	}
	headerName, err := decodeSentinel(r.Header.Get("Mcp-Name"))
	if err != nil {
		return headerMismatch("Mcp-Name header is not valid Base64 sentinel encoding")
	}
	if headerName == "" {
		return headerMismatch("required header Mcp-Name is missing")
	}
	if headerName != bodyName {
		return headerMismatch(fmt.Sprintf("Mcp-Name header %q does not match body value %q", headerName, bodyName))
	}
	return nil
}

func headerMismatch(msg string) *RPCError {
	return &RPCError{Code: ErrCodeHeaderMismatch, Message: "Header mismatch: " + msg}
}

const (
	b64SentinelPrefix = "=?base64?"
	b64SentinelSuffix = "?="
)

func decodeSentinel(v string) (string, error) {
	if len(v) < len(b64SentinelPrefix)+len(b64SentinelSuffix) ||
		!strings.HasPrefix(v, b64SentinelPrefix) || !strings.HasSuffix(v, b64SentinelSuffix) {
		return v, nil
	}
	raw, err := base64.StdEncoding.DecodeString(v[len(b64SentinelPrefix) : len(v)-len(b64SentinelSuffix)])
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (s *Server) handleDiscover() *DiscoverResult {
	return &DiscoverResult{
		SupportedVersions: SupportedVersions,
		Capabilities: ServerCapabilities{
			Tools:     &ToolsCapability{ListChanged: false},
			Resources: &ResourcesCapability{Subscribe: false, ListChanged: false},
		},
		Instructions: "Ech0 personal microblog. Manage posts, tags, comments, files, connects and webhooks via tools; read site data via ech0:// resources.",
		TTLMs:        discoverTTLMs, CacheScope: cacheScopePublic,
	}
}

func (s *Server) handleToolsList() (*ToolsListResult, *RPCError) {
	return &ToolsListResult{
		TTLMs: listTTLMs, CacheScope: cacheScopePublic,
		Tools: s.registry.ToolDefinitions(),
	}, nil
}

func (s *Server) handleToolsCall(r *http.Request, req *Request, v viewer.Context) (*ToolCallResult, *RPCError) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid tool call params"}
	}

	handler, requiredScopes, ok := s.registry.LookupTool(params.Name)
	if !ok {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: fmt.Sprintf("tool %q not found", params.Name)}
	}

	if !checkScopes(v.Scopes(), requiredScopes) {
		return &ToolCallResult{
			Content: []ContentItem{{Type: "text", Text: "permission denied: insufficient scopes"}},
			IsError: true,
		}, nil
	}

	ctx, cancel := context.WithTimeout(r.Context(), toolTimeout)
	defer cancel()

	result, err := handler(ctx, params.Arguments)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &ToolCallResult{
				Content: []ContentItem{{Type: "text", Text: "tool execution timed out"}},
				IsError: true,
			}, nil
		}
		return &ToolCallResult{
			Content: []ContentItem{{Type: "text", Text: err.Error()}},
			IsError: true,
		}, nil
	}
	return result, nil
}

func (s *Server) handleResourcesList() (*ResourcesListResult, *RPCError) {
	return &ResourcesListResult{
		TTLMs: listTTLMs, CacheScope: cacheScopePublic,
		Resources: s.registry.ResourceDefinitions(),
	}, nil
}

func (s *Server) handleResourcesRead(r *http.Request, req *Request, v viewer.Context) (*ResourceReadResult, *RPCError) {
	var params ResourceReadParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid resource read params"}
	}

	handler, requiredScopes, ok := s.registry.LookupResource(params.URI)
	if !ok {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: fmt.Sprintf("resource %q not found", params.URI)}
	}

	if !checkScopes(v.Scopes(), requiredScopes) {
		return nil, &RPCError{Code: ErrCodeInternal, Message: "permission denied: insufficient scopes"}
	}

	result, err := handler(r.Context(), params.URI)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeInternal, Message: err.Error()}
	}
	result.CacheInfo = CacheInfo{TTLMs: 0, CacheScope: cacheScopePrivate}
	return result, nil
}

func checkScopes(actual, required []string) bool {
	if len(required) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(actual))
	for _, s := range actual {
		set[s] = struct{}{}
	}
	for _, s := range required {
		if _, ok := set[s]; !ok {
			return false
		}
	}
	return true
}

func writeRPCResult(w http.ResponseWriter, id json.RawMessage, result any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, rpcErr *RPCError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusFor(rpcErr.Code))
	_ = json.NewEncoder(w).Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   rpcErr,
	})
}

func httpStatusFor(code int) int {
	switch code {
	case ErrCodeParse, ErrCodeInvalidRequest, ErrCodeHeaderMismatch, ErrCodeUnsupportedProtocolVersion:
		return http.StatusBadRequest
	case ErrCodeMethodNotFound:
		return http.StatusNotFound
	default:
		return http.StatusOK
	}
}
