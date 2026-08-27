// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"strings"
	"testing"
	"time"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

func mkStatsEcho(date string, tags ...string) echoModel.Echo {
	ts, _ := time.Parse("2006-01-02", date)
	var tg []echoModel.Tag
	for _, n := range tags {
		tg = append(tg, echoModel.Tag{Name: n})
	}
	return echoModel.Echo{CreatedAt: ts.Unix(), Tags: tg}
}

func TestFormatStatsOverview_ZH(t *testing.T) {
	echos := []echoModel.Echo{
		mkStatsEcho("2025-01-05", "读书"),
		mkStatsEcho("2025-01-20", "读书"),
		mkStatsEcho("2025-03-10", "旅行"),
	}
	got := formatStatsOverview(echos, 3, false, time.UTC, "zh-CN")

	for _, want := range []string{
		"总条数：3 条",
		"活跃天数：3 天",
		"最活跃月份：2025-01（2 条）",
		"2025-01 (2条)",
		"2025-03 (1条)",
		"#读书 (2)",
		"#旅行 (1)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stats output missing %q\n%s", want, got)
		}
	}
}

func TestFormatStatsOverview_TruncatedNote(t *testing.T) {
	echos := []echoModel.Echo{mkStatsEcho("2025-01-05")}
	got := formatStatsOverview(echos, 6000, true, time.UTC, "zh-CN")
	if !strings.Contains(got, "仅统计最近") {
		t.Fatalf("truncated note missing:\n%s", got)
	}
	if !strings.Contains(got, "共命中 6000 条") {
		t.Fatalf("total count missing:\n%s", got)
	}
}

func TestFormatStatsOverview_EN(t *testing.T) {
	echos := []echoModel.Echo{mkStatsEcho("2025-02-01", "books")}
	got := formatStatsOverview(echos, 1, false, time.UTC, "en-US")
	if !strings.Contains(got, "Total: 1") || !strings.Contains(got, "#books (1)") {
		t.Fatalf("EN stats output unexpected:\n%s", got)
	}
}

func TestFormatStatsOverview_Timezone(t *testing.T) {
	sh, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	ts := time.Date(2025, 1, 31, 20, 0, 0, 0, time.UTC)
	echos := []echoModel.Echo{{CreatedAt: ts.Unix()}}

	if got := formatStatsOverview(echos, 1, false, sh, "zh-CN"); !strings.Contains(got, "2025-02 (1条)") {
		t.Fatalf("Shanghai tz should bucket into 2025-02:\n%s", got)
	}
	if got := formatStatsOverview(echos, 1, false, time.UTC, "zh-CN"); !strings.Contains(got, "2025-01 (1条)") {
		t.Fatalf("UTC should bucket into 2025-01:\n%s", got)
	}
}
