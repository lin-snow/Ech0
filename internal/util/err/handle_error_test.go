// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"errors"
	"testing"

	model "github.com/lin-snow/ech0/internal/model/common"
)

func TestHandleError(t *testing.T) {
	t.Run("uses error message when msg empty", func(t *testing.T) {
		se := &model.ServerError{Err: errors.New("boom")}
		got := HandleError(se)
		if got != "boom" {
			t.Fatalf("HandleError returned %q, want %q", got, "boom")
		}
		if se.Msg != "boom" {
			t.Fatalf("HandleError should backfill Msg, got %q", se.Msg)
		}
	})

	t.Run("keeps explicit msg over error", func(t *testing.T) {
		se := &model.ServerError{Msg: "explicit", Err: errors.New("inner")}
		if got := HandleError(se); got != "explicit" {
			t.Fatalf("HandleError returned %q, want %q", got, "explicit")
		}
	})

	t.Run("no error just returns msg", func(t *testing.T) {
		se := &model.ServerError{Msg: "only message"}
		if got := HandleError(se); got != "only message" {
			t.Fatalf("HandleError returned %q, want %q", got, "only message")
		}
	})
}

func TestHandlePanicErrorPanicsWithMsg(t *testing.T) {
	se := &model.ServerError{Msg: "panic message"}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("HandlePanicError did not panic")
		}
		if msg, ok := r.(string); !ok || msg != "panic message" {
			t.Fatalf("HandlePanicError panicked with %v, want %q", r, "panic message")
		}
	}()

	HandlePanicError(se)
}
