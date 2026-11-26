// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package visitor

import (
	"net/http"
	"testing"
	"time"
)

func TestParseDate_UsesUTC(t *testing.T) {
	parsed := parseDate("2026-01-02")
	if parsed.Location() != time.UTC {
		t.Fatalf("expected UTC location, got %s", parsed.Location())
	}
}

func TestRecordIP_UsesUTCDayBoundary(t *testing.T) {
	state := runtimeState{
		byDay:    make(map[string]DayStat),
		todayUV:  make(map[string]struct{}),
		lastPVAt: make(map[string]time.Time),
	}

	ev := recordEvent{
		ipHash: "ip1",
		at:     time.Date(2026, 1, 2, 1, 0, 0, 0, time.FixedZone("CST", 8*3600)),
	}
	recordIP(&state, ev)

	stat, ok := state.byDay["2026-01-01"]
	if !ok {
		t.Fatalf("expected UTC day bucket 2026-01-01, got %+v", state.byDay)
	}
	if stat.PV != 1 || stat.UV != 1 {
		t.Fatalf("expected pv=1 uv=1, got pv=%d uv=%d", stat.PV, stat.UV)
	}
}

func TestIsCrawlerUA(t *testing.T) {
	cases := []struct {
		ua   string
		want bool
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0", false},
		{"Mozilla/5.0 (Linux; Android 12; Cubot X70) AppleWebKit/537.36", false},
		{"", false},
		{"Mozilla/5.0 (compatible; Applebot/0.1; +http://www.apple.com/go/applebot)", true},
		{"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", true},
		{"SofyaBot/1.0", true},
		{"facebookexternalhit/1.1", true},
	}
	for _, tc := range cases {
		if got := isCrawlerUA(tc.ua); got != tc.want {
			t.Fatalf("isCrawlerUA(%q)=%v, want %v", tc.ua, got, tc.want)
		}
	}
}

func TestTrackerRecord_SkipsCrawlerUA(t *testing.T) {
	tracker := NewTracker()

	botReq, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	botReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Applebot/0.1)")
	tracker.Record(botReq, "203.0.113.10")

	// Give the actor time to drain the crawler event, then assert it stayed at zero.
	time.Sleep(20 * time.Millisecond)
	if got := tracker.TodayStat(); got.PV != 0 || got.UV != 0 {
		t.Fatalf("crawler should not count, got pv=%d uv=%d", got.PV, got.UV)
	}

	humanReq, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	humanReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/120.0.0.0")
	tracker.Record(humanReq, "198.51.100.20")

	got := waitTodayStat(t, tracker, func(s DayStat) bool { return s.PV == 1 && s.UV == 1 })
	if got.PV != 1 || got.UV != 1 {
		t.Fatalf("expected only human visit pv=1 uv=1, got pv=%d uv=%d", got.PV, got.UV)
	}
}

func waitTodayStat(t *testing.T, tracker *Tracker, ready func(DayStat) bool) DayStat {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	var last DayStat
	for time.Now().Before(deadline) {
		last = tracker.TodayStat()
		if ready(last) {
			return last
		}
		time.Sleep(time.Millisecond)
	}
	return last
}

func TestTrackerLoadHistory_ResumesTodayFromPersistedStats(t *testing.T) {
	tracker := NewTracker()
	now := time.Now().UTC()
	today := now.Format(dateLayout)
	yesterday := now.AddDate(0, 0, -1).Format(dateLayout)

	tracker.LoadHistory([]DayStat{
		{Date: today, PV: 7, UV: 3},
		{Date: yesterday, PV: 12, UV: 8},
	})

	todayStat := tracker.TodayStat()
	if todayStat.Date != today {
		t.Fatalf("expected today date %s, got %s", today, todayStat.Date)
	}
	if todayStat.PV != 7 || todayStat.UV != 3 {
		t.Fatalf("expected today stat resumed pv=7 uv=3, got pv=%d uv=%d", todayStat.PV, todayStat.UV)
	}

	stats := tracker.Last7Days()
	byDate := make(map[string]DayStat, len(stats))
	for _, stat := range stats {
		byDate[stat.Date] = stat
	}
	loaded := byDate[yesterday]
	if loaded.PV != 12 || loaded.UV != 8 {
		t.Fatalf("expected yesterday stat pv=12 uv=8, got pv=%d uv=%d", loaded.PV, loaded.UV)
	}
}
