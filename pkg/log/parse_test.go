// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package log

import (
	"reflect"
	"testing"
)

func TestToString(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "string passthrough", in: "hello", want: "hello"},
		{name: "empty string passthrough", in: "", want: ""},
		{name: "number marshalled", in: float64(42), want: "42"},
		{name: "bool marshalled", in: true, want: "true"},
		{name: "map marshalled", in: map[string]any{"k": "v"}, want: `{"k":"v"}`},
		{name: "slice marshalled", in: []any{float64(1), "x"}, want: `[1,"x"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toString(tt.in); got != tt.want {
				t.Errorf("toString(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseMapAsEntry(t *testing.T) {
	t.Run("full payload with extra field", func(t *testing.T) {
		payload := map[string]any{
			"time":   "2024-01-01T00:00:00Z",
			"level":  "INFO",
			"msg":    "hello",
			"module": "mod",
			"caller": "c.go:1",
			"error":  "boom",
			"x":      "y",
		}
		got := parseMapAsEntry(payload, "RAW")

		if got.Time != "2024-01-01T00:00:00Z" {
			t.Errorf("Time = %q", got.Time)
		}
		if got.Level != "info" {
			t.Errorf("Level = %q, want %q", got.Level, "info")
		}
		if got.Msg != "hello" {
			t.Errorf("Msg = %q", got.Msg)
		}
		if got.Module != "mod" {
			t.Errorf("Module = %q", got.Module)
		}
		if got.Caller != "c.go:1" {
			t.Errorf("Caller = %q", got.Caller)
		}
		if got.Error != "boom" {
			t.Errorf("Error = %q", got.Error)
		}
		if got.Raw != "RAW" {
			t.Errorf("Raw = %q", got.Raw)
		}
		wantFields := map[string]any{"x": "y"}
		if !reflect.DeepEqual(got.Fields, wantFields) {
			t.Errorf("Fields = %v, want %v", got.Fields, wantFields)
		}
	})

	t.Run("err field promoted to Error when error empty", func(t *testing.T) {
		payload := map[string]any{
			"level": "error",
			"err":   "kaboom",
		}
		got := parseMapAsEntry(payload, "RAWLINE")

		if got.Error != "kaboom" {
			t.Errorf("Error = %q, want %q (err promoted)", got.Error, "kaboom")
		}
		if got.Fields["err"] != "kaboom" {
			t.Errorf("Fields[err] = %v, want %q", got.Fields["err"], "kaboom")
		}
		if got.Msg != "RAWLINE" {
			t.Errorf("Msg = %q, want raw %q", got.Msg, "RAWLINE")
		}
	})

	t.Run("explicit error wins over err field", func(t *testing.T) {
		payload := map[string]any{
			"error": "E1",
			"err":   "E2",
			"msg":   "m",
		}
		got := parseMapAsEntry(payload, "R")

		if got.Error != "E1" {
			t.Errorf("Error = %q, want %q (explicit error wins)", got.Error, "E1")
		}
		if got.Fields["err"] != "E2" {
			t.Errorf("Fields[err] = %v, want %q", got.Fields["err"], "E2")
		}
		if got.Msg != "m" {
			t.Errorf("Msg = %q, want %q", got.Msg, "m")
		}
	})

	t.Run("empty msg falls back to raw", func(t *testing.T) {
		got := parseMapAsEntry(map[string]any{"level": "info"}, "RAWONLY")
		if got.Msg != "RAWONLY" {
			t.Errorf("Msg = %q, want raw %q", got.Msg, "RAWONLY")
		}
		if got.Fields != nil {
			t.Errorf("Fields = %v, want nil (no extra fields)", got.Fields)
		}
	})
}

func TestParseLogLine(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		line := `{"level":"warn","msg":"hello","module":"m"}`
		got := parseLogLine(line)
		if got.Level != "warn" {
			t.Errorf("Level = %q, want %q", got.Level, "warn")
		}
		if got.Msg != "hello" {
			t.Errorf("Msg = %q, want %q", got.Msg, "hello")
		}
		if got.Module != "m" {
			t.Errorf("Module = %q, want %q", got.Module, "m")
		}
		if got.Raw != line {
			t.Errorf("Raw = %q, want %q", got.Raw, line)
		}
	})

	t.Run("invalid json falls back to info level raw", func(t *testing.T) {
		line := `garbage{ not json`
		got := parseLogLine(line)
		if got.Level != "info" {
			t.Errorf("Level = %q, want %q", got.Level, "info")
		}
		if got.Msg != line {
			t.Errorf("Msg = %q, want %q", got.Msg, line)
		}
		if got.Raw != line {
			t.Errorf("Raw = %q, want %q", got.Raw, line)
		}
	})
}

func TestMatchLogFilters(t *testing.T) {
	tests := []struct {
		name    string
		entry   LogEntry
		level   string
		keyword string
		want    bool
	}{
		{
			name:  "no filters matches",
			entry: LogEntry{Level: "info", Msg: "hi"},
			want:  true,
		},
		{
			name:  "level all matches any",
			entry: LogEntry{Level: "error"},
			level: "all",
			want:  true,
		},
		{
			name:  "matching level",
			entry: LogEntry{Level: "info"},
			level: "info",
			want:  true,
		},
		{
			name:  "non-matching level filtered out",
			entry: LogEntry{Level: "info"},
			level: "error",
			want:  false,
		},
		{
			name:  "entry level case-insensitive",
			entry: LogEntry{Level: "INFO"},
			level: "info",
			want:  true,
		},
		{
			name:    "keyword matches msg",
			entry:   LogEntry{Level: "info", Msg: "FooBar", Raw: "{}"},
			keyword: "foobar",
			want:    true,
		},
		{
			name:    "keyword matches raw",
			entry:   LogEntry{Level: "info", Msg: "x", Raw: `{"msg":"needle"}`},
			keyword: "needle",
			want:    true,
		},
		{
			name:    "keyword not found filtered out",
			entry:   LogEntry{Level: "info", Msg: "abc", Raw: "abc"},
			keyword: "zzz",
			want:    false,
		},
		{
			name:    "level and keyword both required",
			entry:   LogEntry{Level: "error", Msg: "boom", Raw: "boom"},
			level:   "error",
			keyword: "boom",
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchLogFilters(tt.entry, tt.level, tt.keyword); got != tt.want {
				t.Errorf("matchLogFilters(%+v, %q, %q) = %v, want %v", tt.entry, tt.level, tt.keyword, got, tt.want)
			}
		})
	}
}
