// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package leakcheck

import (
	"runtime"
	"strings"
	"testing"
)

func TestParseTotal(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		want    int
		wantErr bool
	}{
		{name: "no leaks", header: "goroutineleak profile: total 0", want: 0},
		{name: "several leaks", header: "goroutineleak profile: total 12", want: 12},
		{name: "missing count", header: "goroutineleak profile: total", wantErr: true},
		{name: "not a number", header: "goroutineleak profile: total many", wantErr: true},
		{name: "foreign header", header: "goroutine profile: 3", wantErr: true},
		{name: "empty", header: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTotal(tt.header)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseTotal(%q) = %d, want error", tt.header, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTotal(%q) error = %v", tt.header, err)
			}
			if got != tt.want {
				t.Fatalf("parseTotal(%q) = %d, want %d", tt.header, got, tt.want)
			}
		})
	}
}

func TestProfileCountsLeakedGoroutine(t *testing.T) {
	_, before, err := profile()
	if err != nil {
		t.Fatalf("profile() error = %v", err)
	}

	leakOneGoroutine()

	for attempt := range 100 {
		report, after, err := profile()
		if err != nil {
			t.Fatalf("profile() error = %v", err)
		}
		if after == before+1 {
			if !strings.Contains(report, "leakOneGoroutine") {
				t.Fatalf("report does not name the leaking function:\n%s", report)
			}
			return
		}
		if after != before {
			t.Fatalf("leak count = %d, want %d or %d (attempt %d)", after, before, before+1, attempt)
		}
		runtime.Gosched()
	}
	t.Fatal("leaked goroutine was never reported by the goroutineleak profile")
}

func leakOneGoroutine() {
	started := make(chan struct{})
	blocked := make(chan struct{})
	go func() {
		close(started)
		<-blocked
	}()
	<-started
}
