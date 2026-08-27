// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import "testing"

func TestStatusIsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{name: "success is terminal", status: StatusSuccess, want: true},
		{name: "failed is terminal", status: StatusFailed, want: true},
		{name: "cancelled is terminal", status: StatusCancelled, want: true},
		{name: "pending is not terminal", status: StatusPending, want: false},
		{name: "running is not terminal", status: StatusRunning, want: false},
		{name: "empty string is not terminal", status: Status(""), want: false},
		{name: "unknown value is not terminal", status: Status("paused"), want: false},
		{name: "case-sensitive: SUCCESS is not terminal", status: Status("SUCCESS"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.IsTerminal(); got != tt.want {
				t.Fatalf("Status(%q).IsTerminal() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
