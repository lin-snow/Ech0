// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package log

import "testing"

func TestDefaultLogConfig(t *testing.T) {
	cfg := DefaultLogConfig()

	if cfg.Level != "info" {
		t.Errorf("Level = %q, want %q", cfg.Level, "info")
	}
	if cfg.Format != "json" {
		t.Errorf("Format = %q, want %q", cfg.Format, "json")
	}
	if cfg.Console {
		t.Errorf("Console = %v, want false", cfg.Console)
	}

	wantFile := FileConfig{
		Enable:     true,
		Filename:   "data/app.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
	if cfg.File != wantFile {
		t.Errorf("File = %+v, want %+v", cfg.File, wantFile)
	}

	wantStream := StreamConfig{
		BufferSize:      2048,
		RecentSize:      2000,
		DropPolicy:      "drop_oldest",
		FlushBatch:      128,
		FlushIntervalMs: 500,
	}
	if cfg.Stream != wantStream {
		t.Errorf("Stream = %+v, want %+v", cfg.Stream, wantStream)
	}
}

func TestSafePositive(t *testing.T) {
	tests := []struct {
		name     string
		v        int
		fallback int
		want     int
	}{
		{name: "positive returns value", v: 1, fallback: 99, want: 1},
		{name: "large positive returns value", v: 100, fallback: 5, want: 100},
		{name: "zero returns fallback", v: 0, fallback: 99, want: 99},
		{name: "negative returns fallback", v: -5, fallback: 99, want: 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := safePositive(tt.v, tt.fallback); got != tt.want {
				t.Errorf("safePositive(%d, %d) = %d, want %d", tt.v, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestNormalizeDropPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy string
		want   string
	}{
		{name: "drop_newest passthrough", policy: "drop_newest", want: "drop_newest"},
		{name: "drop_newest uppercase", policy: "DROP_NEWEST", want: "drop_newest"},
		{name: "drop_newest with spaces", policy: "  drop_newest  ", want: "drop_newest"},
		{name: "drop_oldest passthrough", policy: "drop_oldest", want: "drop_oldest"},
		{name: "empty defaults to drop_oldest", policy: "", want: "drop_oldest"},
		{name: "unknown defaults to drop_oldest", policy: "garbage", want: "drop_oldest"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeDropPolicy(tt.policy); got != tt.want {
				t.Errorf("normalizeDropPolicy(%q) = %q, want %q", tt.policy, got, tt.want)
			}
		})
	}
}
