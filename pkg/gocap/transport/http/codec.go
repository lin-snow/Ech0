// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package caphttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lin-snow/ech0/pkg/gocap/core"
)

type errorBody struct {
	Success bool   `json:"success"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error"`
}

type decodeOptions struct {
	Strict bool
}

func decodeJSON(r *http.Request, out any, opts decodeOptions) error {
	dec := json.NewDecoder(r.Body)
	if opts.Strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(out); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return core.NewBadRequest("Malformed JSON body")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeCoreError(w http.ResponseWriter, err error) {
	if ce, ok := errors.AsType[*core.Error](err); ok {
		writeJSON(w, ce.StatusCode, errorBody{
			Success: false,
			Code:    string(ce.Code),
			Error:   ce.Message,
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, errorBody{
		Success: false,
		Code:    string(core.ErrCodeInternal),
		Error:   "Internal server error",
	})
}

func writeDecodeError(w http.ResponseWriter, err error) {
	if ce, ok := errors.AsType[*core.Error](err); ok {
		writeCoreError(w, ce)
		return
	}
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		writeCoreError(w, core.NewBadRequest("Request body too large"))
		return
	}
	writeCoreError(w, core.NewBadRequest("Malformed JSON body"))
}
