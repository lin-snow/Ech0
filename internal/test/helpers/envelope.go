// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

type APIResult struct {
	Code       int             `json:"code"`
	Msg        string          `json:"msg"`
	ErrorCode  string          `json:"error_code"`
	MessageKey string          `json:"message_key"`
	Data       json.RawMessage `json:"data"`
}

func ParseResult(t *testing.T, rec *httptest.ResponseRecorder) APIResult {
	t.Helper()
	var r APIResult
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("helpers: parse result envelope: %v\nbody: %s", err, rec.Body.String())
	}
	return r
}

func DecodeData(t *testing.T, raw json.RawMessage, dest any) {
	t.Helper()
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatalf("helpers: decode data: %v\nraw: %s", err, string(raw))
	}
}
