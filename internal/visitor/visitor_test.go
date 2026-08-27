// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package visitor

import (
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
