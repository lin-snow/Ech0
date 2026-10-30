// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	openapiCatalog "github.com/lin-snow/ech0/internal/openapi/catalog"
)

func TestCatalog_RouteFingerprintAndStaleGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	initTestDatabase(t)
	engine := gin.New()
	SetupRouter(engine, buildTestHandlers(), buildTestMWDeps())

	// 1. The catalog is served as canonical bytes with a matching sha256 header.
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/_catalog", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("catalog status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	fingerprint := recorder.Header().Get(openapiCatalog.HeaderName)
	sum := sha256.Sum256(recorder.Body.Bytes())
	if fingerprint == "" || fingerprint != hex.EncodeToString(sum[:]) {
		t.Fatalf("fingerprint header %q does not match body hash", fingerprint)
	}

	var doc openapiCatalog.Doc
	if err := json.Unmarshal(recorder.Body.Bytes(), &doc); err != nil {
		t.Fatalf("catalog body is not JSON: %v", err)
	}
	create, ok := doc.Ops["echo.create"]
	if !ok {
		t.Fatalf("echo.create missing; got %d ops", len(doc.Ops))
	}
	if create.Method != http.MethodPost || create.Path != "/echo" || create.Effect != openapiCatalog.EffectWrite {
		t.Fatalf("echo.create = %+v", create)
	}
	if len(create.Scopes) != 1 || create.Scopes[0] != "echo:write" {
		t.Fatalf("echo.create scopes = %v", create.Scopes)
	}

	// 2. A request built against the current catalog passes through.
	matching := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	matching.Header.Set(openapiCatalog.HeaderName, fingerprint)
	matchedRecorder := httptest.NewRecorder()
	engine.ServeHTTP(matchedRecorder, matching)
	if matchedRecorder.Code != http.StatusOK {
		t.Fatalf("matching fingerprint status = %d", matchedRecorder.Code)
	}
	if got := matchedRecorder.Header().Get(openapiCatalog.HeaderName); got != fingerprint {
		t.Fatalf("response fingerprint = %q", got)
	}

	// 3. A stale fingerprint is refused before the handler runs.
	stale := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	stale.Header.Set(openapiCatalog.HeaderName, "stale-fingerprint")
	staleRecorder := httptest.NewRecorder()
	engine.ServeHTTP(staleRecorder, stale)
	if staleRecorder.Code != http.StatusPreconditionFailed {
		t.Fatalf("stale status = %d", staleRecorder.Code)
	}
	var failure commonModel.Result[any]
	if err := json.Unmarshal(staleRecorder.Body.Bytes(), &failure); err != nil {
		t.Fatalf("stale body is not JSON: %v", err)
	}
	if failure.ErrorCode != openapiCatalog.StaleErrorCode {
		t.Fatalf("error_code = %q", failure.ErrorCode)
	}
	if got := staleRecorder.Header().Get(openapiCatalog.HeaderName); got != fingerprint {
		t.Fatalf("stale response fingerprint = %q", got)
	}
}
