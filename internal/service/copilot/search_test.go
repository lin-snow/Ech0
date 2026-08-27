// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
)

func TestResolveTagIDs(t *testing.T) {
	tags := []echoModel.Tag{
		{ID: "id-read", Name: "读书"},
		{ID: "id-travel", Name: "Travel"},
	}
	cases := []struct {
		name  string
		in    []string
		wantN int
		want  []string
	}{
		{"empty input → nil", nil, 0, nil},
		{"exact match", []string{"读书"}, 1, []string{"id-read"}},
		{"case-insensitive + trim", []string{"  travel  "}, 1, []string{"id-travel"}},
		{"unknown ignored", []string{"读书", "不存在"}, 1, []string{"id-read"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := resolveTagIDs(tags, c.in)
			if len(got) != c.wantN {
				t.Fatalf("len = %d, want %d (got %v)", len(got), c.wantN, got)
			}
			for i := range c.want {
				if got[i] != c.want[i] {
					t.Fatalf("got %v, want %v", got, c.want)
				}
			}
		})
	}
}

func TestParseDay(t *testing.T) {
	start := parseDay("2026-01-15", false, time.UTC)
	wantStart := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC).Unix()
	if start != wantStart {
		t.Fatalf("start = %d, want %d", start, wantStart)
	}

	end := parseDay("2026-01-15", true, time.UTC)
	wantEnd := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC).Add(24*time.Hour - time.Second).Unix()
	if end != wantEnd {
		t.Fatalf("end = %d, want %d", end, wantEnd)
	}
	if end-start != 24*3600-1 {
		t.Fatalf("end-of-day span = %d, want %d", end-start, 24*3600-1)
	}

	if got := parseDay("", false, time.UTC); got != 0 {
		t.Fatalf("empty → %d, want 0", got)
	}
	if got := parseDay("not-a-date", false, time.UTC); got != 0 {
		t.Fatalf("invalid → %d, want 0", got)
	}

	sh, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	gotSH := parseDay("2026-01-15", false, sh)
	wantSH := time.Date(2026, 1, 15, 0, 0, 0, 0, sh).Unix()
	if gotSH != wantSH {
		t.Fatalf("Shanghai start = %d, want %d", gotSH, wantSH)
	}
	if gotSH != wantStart-8*3600 {
		t.Fatalf("Shanghai start should be 8h earlier than UTC: got %d, utc %d", gotSH, wantStart)
	}
}

func TestSearchHintOf(t *testing.T) {
	mk := func(a searchArgs) json.RawMessage {
		b, _ := json.Marshal(a)
		return b
	}
	cases := []struct {
		name string
		args searchArgs
		want string
	}{
		{"empty → empty", searchArgs{}, ""},
		{"query only", searchArgs{Query: "三体"}, "三体"},
		{"tags get # prefix", searchArgs{Tags: []string{"读书"}}, "#读书"},
		{"combined", searchArgs{Query: "旅行", Tags: []string{"游记"}, DateFrom: "2026-01-01", DateTo: "2026-02-01"}, "旅行 #游记 2026-01-01~2026-02-01"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := searchHintOf(mk(c.args)); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestFormatSearchResults(t *testing.T) {
	if got := formatSearchResults(nil, nil, time.UTC); got != "（没有检索到相关的 Echo）" {
		t.Fatalf("empty results = %q", got)
	}

	day := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC).Unix()
	results := []embeddingModel.SearchResult{
		{EchoID: "e1", Content: "今天读了三体", EchoCreated: day},
	}
	exts := map[string]string{"e1": "[音乐分享] https://example.com/song"}

	got := formatSearchResults(results, exts, time.UTC)
	if !strings.Contains(got, "【1】(2026-03-01)") {
		t.Fatalf("missing index/date frame: %q", got)
	}
	if !strings.Contains(got, "今天读了三体") {
		t.Fatalf("missing content: %q", got)
	}
	if !strings.Contains(got, "[音乐分享]") {
		t.Fatalf("extension text should be appended: %q", got)
	}
}
