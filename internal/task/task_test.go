package task

import (
	"testing"
	"time"

	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	"github.com/lin-snow/ech0/internal/visitor"
)

func TestBuildVisitorDailyStat(t *testing.T) {
	input := visitor.DayStat{
		Date: "2026-01-02",
		PV:   13,
		UV:   9,
	}
	output := buildVisitorDailyStat(input)
	if output.Date != input.Date || output.PV != input.PV || output.UV != input.UV {
		t.Fatalf("unexpected mapping result: %+v", output)
	}
}

func TestConvertVisitorHistory(t *testing.T) {
	input := []visitorModel.DailyStat{
		{Date: "2026-01-01", PV: 1, UV: 1},
		{Date: "2026-01-02", PV: 2, UV: 2},
	}
	output := convertVisitorHistory(input)
	if len(output) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(output))
	}
	if output[0].Date != "2026-01-01" || output[1].Date != "2026-01-02" {
		t.Fatalf("unexpected order after conversion: %+v", output)
	}
}

func TestVisitorCutoffDate_KeepRecentSixDays(t *testing.T) {
	now := time.Date(2026, 4, 16, 23, 59, 59, 0, time.UTC)
	cutoff := visitorCutoffDate(now)
	if cutoff != "2026-04-11" {
		t.Fatalf("expected cutoff 2026-04-11, got %s", cutoff)
	}
}
