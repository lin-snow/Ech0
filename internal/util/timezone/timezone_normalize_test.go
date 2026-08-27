// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"testing"
	"time"
)

func TestNormalizeTimezone(t *testing.T) {
	tests := []struct {
		name string
		tz   string
		want string
	}{
		{name: "valid named zone", tz: "Asia/Shanghai", want: "Asia/Shanghai"},
		{name: "valid us zone", tz: "America/New_York", want: "America/New_York"},
		{name: "explicit utc", tz: "UTC", want: "UTC"},
		{name: "empty falls back to utc", tz: "", want: "UTC"},
		{name: "invalid falls back to utc", tz: "Not/AZone", want: "UTC"},
		{name: "garbage falls back to utc", tz: "!!!", want: "UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeTimezone(tt.tz); got != tt.want {
				t.Fatalf("NormalizeTimezone(%q) = %q, want %q", tt.tz, got, tt.want)
			}
		})
	}
}

func TestLoadLocationOrUTC(t *testing.T) {
	tests := []struct {
		name string
		tz   string
		want string
	}{
		{name: "valid named zone", tz: "Asia/Shanghai", want: "Asia/Shanghai"},
		{name: "explicit utc", tz: "UTC", want: "UTC"},
		{name: "empty falls back to utc", tz: "", want: "UTC"},
		{name: "invalid falls back to utc", tz: "Not/AZone", want: "UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := LoadLocationOrUTC(tt.tz)
			if loc == nil {
				t.Fatalf("LoadLocationOrUTC(%q) returned nil location", tt.tz)
			}
			if loc.String() != tt.want {
				t.Fatalf("LoadLocationOrUTC(%q) = %q, want %q", tt.tz, loc.String(), tt.want)
			}
		})
	}
}

func TestLoadLocationOrUTCInvalidIsUTCEquivalent(t *testing.T) {
	loc := LoadLocationOrUTC("definitely/not/real")
	_, offset := time.Date(2026, 1, 1, 0, 0, 0, 0, loc).Zone()
	if offset != 0 {
		t.Fatalf("expected UTC zero offset for invalid tz, got offset=%d", offset)
	}
}
