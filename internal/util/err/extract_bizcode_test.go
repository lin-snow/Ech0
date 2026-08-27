// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"errors"
	"fmt"
	"testing"

	model "github.com/lin-snow/ech0/internal/model/common"
)

func TestExtractBizErrorCode(t *testing.T) {
	bare := model.NewBizError(model.ErrCodeInvalidRequest, "invalid request")

	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "nil error", err: nil, want: ""},
		{name: "plain biz error", err: bare, want: model.ErrCodeInvalidRequest},
		{
			name: "biz error carrying inner cause",
			err:  &model.BizError{Code: model.ErrCodeInternal, Msg: "boom", Err: errors.New("disk full")},
			want: model.ErrCodeInternal,
		},
		{
			name: "single wrap",
			err:  fmt.Errorf("context: %w", model.NewBizError(model.ErrCodePermissionDenied, "denied")),
			want: model.ErrCodePermissionDenied,
		},
		{
			name: "deeply wrapped",
			err: fmt.Errorf("outer: %w",
				fmt.Errorf("middle: %w",
					model.NewBizError(model.ErrCodeTokenInvalid, "bad token"))),
			want: model.ErrCodeTokenInvalid,
		},
		{name: "non biz error", err: errors.New("just an error"), want: ""},
		{name: "wrapped non biz error", err: fmt.Errorf("ctx: %w", errors.New("plain")), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractBizErrorCode(tt.err); got != tt.want {
				t.Fatalf("ExtractBizErrorCode(%v) = %q, want %q", tt.err, got, tt.want)
			}
		})
	}
}
