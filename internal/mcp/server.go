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

const (
	toolTimeout     = 10 * time.Second
	maxRequestBytes = 256 * 1024
)

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

	body, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBytes))
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

	result, requestEra, rpcErr := s.dispatch(r, &req, v)

	logUtil.GetLogger().Info("mcp_request",
		slog.String("method", req.Method),
		slog.String("user_id", v.UserID()),
		slog.String("token_id", v.TokenID()),
		slog.String("era", eraName(requestEra)),
		slog.Duration("latency", time.Since(start)),
		slog.Bool("error", rpcErr != nil),
	)

	if rpcErr != nil {
		writeRPCError(w, req.ID, rpcErr)
		return
	}
	if requestEra == eraModern {
		if c, ok := result.(completer); ok {
			c.complete(serverInfo())
		}
	}
	writeRPCResult(w, req.ID, result)
}

type requestParams struct {
	Meta map[string]any `json:"_meta"`
	Name string         `json:"name"`
	URI  string         `json:"uri"`
}

func (s *Server) dispatch(r *http.Request, req *Request, v viewer.Context) (any, era, *RPCError) {
	var params requestParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, eraLegacy, &RPCError{Code: ErrCodeInvalidRequest, Message: "params must be an object"}
		}
	}

	requestEra, rpcErr := resolveEra(r, &params)
	if rpcErr != nil {
		return nil, requestEra, rpcErr
	}
	if requestEra == eraModern {
		if rpcErr := validateModernTransport(r, req.Method, &params); rpcErr != nil {
			return nil, requestEra, rpcErr
		}
	}

	switch req.Method {
	case "initialize":
		if requestEra == eraModern {
			return nil, requestEra, initializeRemoved()
		}
		return s.handleInitialize(req), requestEra, nil
	case "ping":
		if requestEra == eraModern {
			return nil, requestEra, methodNotFound(req.Method)
		}
		return struct{}{}, requestEra, nil
	case "server/discover":
		return s.handleDiscover(requestEra), requestEra, nil
	case "tools/list":
		return s.handleToolsList(requestEra), requestEra, nil
	case "tools/call":
		result, rpcErr := s.handleToolsCall(r, req, v)
		return result, requestEra, rpcErr
	case "resources/list":
		return s.handleResourcesList(requestEra), requestEra, nil
	case "resources/templates/list":
		return s.handleResourceTemplatesList(requestEra), requestEra, nil
	case "resources/read":
		result, rpcErr := s.handleResourcesRead(r, req, v, requestEra)
		return result, requestEra, rpcErr
	default:
		return nil, requestEra, methodNotFound(req.Method)
	}
}

func resolveEra(r *http.Request, params *requestParams) (era, *RPCError) {
	headerVersion := r.Header.Get("MCP-Protocol-Version")
	bodyVersion, _ := params.Meta[metaKeyProtocolVersion].(string)

	if headerVersion != "" && bodyVersion != "" && headerVersion != bodyVersion {
		return eraLegacy, headerMismatch(fmt.Sprintf(
			"MCP-Protocol-Version header %q does not match body value %q", headerVersion, bodyVersion))
	}

	version := headerVersion
	if version == "" {
		version = bodyVersion
	}

	switch {
	case version == "":
		return eraLegacy, nil
	case version == ProtocolVersion:
		return eraModern, nil
	case isLegacyVersion(version):
		return eraLegacy, nil
	default:
		return eraLegacy, unsupportedVersion(version)
	}
}

func validateModernTransport(r *http.Request, method string, params *requestParams) *RPCError {
	if r.Header.Get("MCP-Protocol-Version") == "" {
		return headerMismatch("required header MCP-Protocol-Version is missing")
	}
	if _, ok := params.Meta[metaKeyProtocolVersion].(string); !ok {
		return malformedMeta("params._meta is missing " + metaKeyProtocolVersion)
	}
	if _, ok := params.Meta[metaKeyClientCapabilities]; !ok {
		return malformedMeta("params._meta is missing " + metaKeyClientCapabilities)
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

func malformedMeta(msg string) *RPCError {
	return (&RPCError{
		Code:    ErrCodeInvalidParams,
		Message: "Malformed request metadata: " + msg,
	}).withHTTPStatus(http.StatusBadRequest)
}

func unsupportedVersion(requested string) *RPCError {
	return &RPCError{
		Code:    ErrCodeUnsupportedProtocolVersion,
		Message: "Unsupported protocol version",
		Data:    map[string]any{"supported": SupportedVersions, "requested": requested},
	}
}

func methodNotFound(method string) *RPCError {
	return &RPCError{Code: ErrCodeMethodNotFound, Message: fmt.Sprintf("method %q not found", method)}
}

func initializeRemoved() *RPCError {
	return &RPCError{
		Code: ErrCodeMethodNotFound,
		Message: "the initialize handshake was removed in MCP " + ProtocolVersion +
			"; send requests statelessly, or open with one of: " + strings.Join(SupportedVersions[1:], ", "),
		Data: map[string]any{"supported": SupportedVersions},
	}
}

func insufficientScope(required []string) *RPCError {
	scope := strings.Join(required, " ")
	return (&RPCError{
		Code:    ErrCodeInsufficientScope,
		Message: "insufficient scope: this token is missing " + scope,
		Data:    map[string]any{"requiredScopes": required},
	}).
		withHTTPStatus(http.StatusForbidden).
		withChallenge(`Bearer error="insufficient_scope", scope="` + scope + `"`)
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

type initializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
}

func (s *Server) handleInitialize(req *Request) *InitializeResult {
	var params initializeParams
	if len(req.Params) > 0 {
		_ = json.Unmarshal(req.Params, &params)
	}
	negotiated := preferredLegacyVersion
	if isLegacyVersion(params.ProtocolVersion) {
		negotiated = params.ProtocolVersion
	}
	return &InitializeResult{
		ProtocolVersion: negotiated,
		Capabilities:    serverCapabilities(),
		ServerInfo:      serverInfo(),
		Instructions:    serverInstructions,
	}
}

func (s *Server) handleDiscover(requestEra era) *DiscoverResult {
	result := &DiscoverResult{
		SupportedVersions: SupportedVersions,
		Capabilities:      serverCapabilities(),
		Instructions:      serverInstructions,
	}
	if requestEra == eraModern {
		cache := publicCache(discoverTTLMs)
		result.CacheInfo = &cache
	}
	return result
}

func (s *Server) handleToolsList(requestEra era) *ToolsListResult {
	result := &ToolsListResult{Tools: s.registry.ToolDefinitions()}
	if requestEra == eraModern {
		cache := publicCache(listTTLMs)
		result.CacheInfo = &cache
	}
	return result
}

func (s *Server) handleToolsCall(r *http.Request, req *Request, v viewer.Context) (*ToolCallResult, *RPCError) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid tool call params"}
	}

	binding, ok := s.registry.LookupTool(params.Name)
	if !ok {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: fmt.Sprintf("tool %q not found", params.Name)}
	}

	if !checkScopes(v.Scopes(), binding.Scopes) {
		return nil, insufficientScope(binding.Scopes)
	}

	ctx, cancel := context.WithTimeout(r.Context(), toolTimeout)
	defer cancel()

	result, err := handlerResult(ctx, binding.Handler, params.Arguments)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func handlerResult(ctx context.Context, handler ToolHandler, args map[string]any) (*ToolCallResult, *RPCError) {
	result, err := handler(ctx, args)
	if err == nil {
		return result, nil
	}
	if ctx.Err() == context.DeadlineExceeded {
		return textError("tool execution timed out"), nil
	}
	return textError(err.Error()), nil
}

func (s *Server) handleResourcesList(requestEra era) *ResourcesListResult {
	result := &ResourcesListResult{Resources: s.registry.ResourceDefinitions()}
	if requestEra == eraModern {
		cache := publicCache(listTTLMs)
		result.CacheInfo = &cache
	}
	return result
}

func (s *Server) handleResourceTemplatesList(requestEra era) *ResourceTemplatesListResult {
	result := &ResourceTemplatesListResult{ResourceTemplates: s.registry.ResourceTemplateDefinitions()}
	if requestEra == eraModern {
		cache := publicCache(listTTLMs)
		result.CacheInfo = &cache
	}
	return result
}

func (s *Server) handleResourcesRead(
	r *http.Request,
	req *Request,
	v viewer.Context,
	requestEra era,
) (*ResourceReadResult, *RPCError) {
	var params ResourceReadParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "invalid resource read params"}
	}

	binding, ok := s.registry.LookupResource(params.URI)
	if !ok {
		return nil, &RPCError{
			Code:    ErrCodeInvalidParams,
			Message: fmt.Sprintf("resource %q not found", params.URI),
			Data:    map[string]any{"uri": params.URI},
		}
	}

	if !checkScopes(v.Scopes(), binding.Scopes) {
		return nil, insufficientScope(binding.Scopes)
	}

	result, err := binding.Handler(r.Context(), params.URI)
	if err != nil {
		return nil, &RPCError{Code: ErrCodeInternal, Message: err.Error()}
	}
	if requestEra == eraModern {
		cache := binding.Cache
		result.CacheInfo = &cache
	}
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

func eraName(requestEra era) string {
	if requestEra == eraModern {
		return ProtocolVersion
	}
	return "legacy"
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
	if rpcErr.challenge != "" {
		w.Header().Set("WWW-Authenticate", rpcErr.challenge)
	}
	w.WriteHeader(httpStatusFor(rpcErr))
	_ = json.NewEncoder(w).Encode(Response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   rpcErr,
	})
}

func httpStatusFor(rpcErr *RPCError) int {
	if rpcErr.httpStatus != 0 {
		return rpcErr.httpStatus
	}
	switch rpcErr.Code {
	case ErrCodeParse, ErrCodeInvalidRequest, ErrCodeHeaderMismatch, ErrCodeUnsupportedProtocolVersion:
		return http.StatusBadRequest
	case ErrCodeMethodNotFound:
		return http.StatusNotFound
	default:
		return http.StatusOK
	}
}
