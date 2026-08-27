// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import (
	"testing"
	"time"

	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	"github.com/lin-snow/ech0/internal/visitor"
)

func TestBuildDailyStat(t *testing.T) {
	input := visitor.DayStat{
		Date: "2026-01-02",
		PV:   13,
		UV:   9,
	}
	output := buildDailyStat(input)
	if output.Date != input.Date || output.PV != input.PV || output.UV != input.UV {
		t.Fatalf("unexpected mapping result: %+v", output)
	}
}

func TestConvertHistory(t *testing.T) {
	input := []visitorModel.DailyStat{
		{Date: "2026-01-01", PV: 1, UV: 1},
		{Date: "2026-01-02", PV: 2, UV: 2},
	}
	output := convertHistory(input)
	if len(output) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(output))
	}
	if output[0].Date != "2026-01-01" || output[1].Date != "2026-01-02" {
		t.Fatalf("unexpected order after conversion: %+v", output)
	}
}

func TestCutoffDate_KeepRecentSevenDays(t *testing.T) {
	now := time.Date(2026, 4, 16, 23, 59, 59, 0, time.UTC)
	cutoff := cutoffDate(now)
	if cutoff != "2026-04-10" {
		t.Fatalf("expected cutoff 2026-04-10, got %s", cutoff)
	}
}
