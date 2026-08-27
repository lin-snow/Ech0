// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import "testing"

func TestValidateCrontabExpression(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{name: "five fields every minute", expr: "* * * * *", wantErr: false},
		{name: "five fields midnight daily", expr: "0 0 * * *", wantErr: false},
		{name: "five fields with descriptor", expr: "@daily", wantErr: true},
		{name: "six fields every second", expr: "* * * * * *", wantErr: false},
		{name: "six fields midnight daily", expr: "0 0 0 * * *", wantErr: false},
		{name: "leading and trailing whitespace five fields", expr: "  * * * * *  ", wantErr: false},
		{name: "leading and trailing whitespace six fields", expr: "\t0 0 0 * * *\n", wantErr: false},
		{name: "empty string", expr: "", wantErr: true},
		{name: "only whitespace", expr: "   ", wantErr: true},
		{name: "four fields too few", expr: "* * * *", wantErr: true},
		{name: "seven fields too many", expr: "* * * * * * *", wantErr: true},
		{name: "five fields malformed minute out of range", expr: "99 * * * *", wantErr: true},
		{name: "five fields non numeric field", expr: "bad * * * *", wantErr: true},
		{name: "six fields malformed second out of range", expr: "60 * * * * *", wantErr: true},
		{name: "six fields non numeric field", expr: "* * * * * bad", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCrontabExpression(tt.expr)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateCrontabExpression(%q) = nil, want error", tt.expr)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateCrontabExpression(%q) = %v, want nil", tt.expr, err)
			}
		})
	}
}
