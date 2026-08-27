// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"testing"
	"time"
)

func TestOnThisDayUnixRanges(t *testing.T) {
	utc := time.UTC

	t.Run("past years for a normal date", func(t *testing.T) {
		got := onThisDayUnixRanges(2020, 2024, time.March, 15, utc)
		if len(got) != 4 {
			t.Fatalf("expected 4 ranges (2020-2023), got %d", len(got))
		}
		for i, year := range []int{2020, 2021, 2022, 2023} {
			wantStart := time.Date(year, time.March, 15, 0, 0, 0, 0, utc).Unix()
			wantEnd := wantStart + 86400
			if got[i][0] != wantStart || got[i][1] != wantEnd {
				t.Errorf("year %d: got [%d,%d), want [%d,%d)", year, got[i][0], got[i][1], wantStart, wantEnd)
			}
		}
	})

	t.Run("current year is excluded", func(t *testing.T) {
		if got := onThisDayUnixRanges(2024, 2024, time.March, 15, utc); len(got) != 0 {
			t.Fatalf("expected 0 ranges when no past years, got %d", len(got))
		}
	})

	t.Run("feb 29 only matches leap years", func(t *testing.T) {
		got := onThisDayUnixRanges(2018, 2025, time.February, 29, utc)
		if len(got) != 2 {
			t.Fatalf("expected 2 ranges (2020, 2024), got %d", len(got))
		}
		for i, year := range []int{2020, 2024} {
			wantStart := time.Date(year, time.February, 29, 0, 0, 0, 0, utc).Unix()
			if got[i][0] != wantStart {
				t.Errorf("range %d: got start %d, want %d (Feb 29 %d)", i, got[i][0], wantStart, year)
			}
		}
	})

	t.Run("DST day is shorter than 24h in its own timezone", func(t *testing.T) {
		loc, err := time.LoadLocation("America/New_York")
		if err != nil {
			t.Skip("tzdata unavailable:", err)
		}
		got := onThisDayUnixRanges(2021, 2022, time.March, 14, loc)
		if len(got) != 1 {
			t.Fatalf("expected 1 range, got %d", len(got))
		}
		if span := got[0][1] - got[0][0]; span != 23*3600 {
			t.Errorf("spring-forward day span = %ds, want %ds", span, 23*3600)
		}
	})
}
