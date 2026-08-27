// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

const statsTopTags = 10

type statsArgs struct {
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
	Tags     []string `json:"tags"`
}

func (s *CopilotService) statsOverviewTool(allTags []echoModel.Tag, locale string, loc *time.Location, user chatUser) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "stats_overview",
			Description: "统计某时间区间内 Echo 的精确量化指标（总条数、活跃天数、按月分布、最活跃月份、配图数、常用标签 Top N）。回答“我（今年/某段时间）发了多少条”“最活跃的月份”“最常用的标签”这类需要**确切数字**的问题时用它——数据来自数据库精确统计，而非采样估算。必须提供 date_from 与 date_to；可选 tags（按标签名限定主题）。",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"date_from":{"type":"string","description":"起始日期，格式 YYYY-MM-DD，含当天"},"date_to":{"type":"string","description":"结束日期，格式 YYYY-MM-DD，含当天"},"tags":{"type":"array","items":{"type":"string"},"description":"可选，按标签名限定主题；可用标签见系统提示"}},"required":["date_from","date_to"]}`),
		},
		Execute: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var a statsArgs
			_ = json.Unmarshal(args, &a)
			from := parseDay(a.DateFrom, false, loc)
			to := parseDay(a.DateTo, true, loc)
			if from == 0 && to == 0 {
				return agent.ToolOutput{}, errors.New("stats_overview 需要 date_from 与 date_to 指定时间区间")
			}
			tagIDs := resolveTagIDs(allTags, a.Tags)

			echos, total, truncated, err := s.collectRange(ctx, user.ID, tagIDs, from, to)
			if err != nil {
				return agent.ToolOutput{}, err
			}

			return agent.ToolOutput{Content: formatStatsOverview(echos, total, truncated, loc, locale)}, nil
		},
	}
}

func formatStatsOverview(echos []echoModel.Echo, total int64, truncated bool, loc *time.Location, locale string) string {
	returned := len(echos)
	monthCounts := make(map[string]int)
	activeDays := make(map[string]struct{})
	tagCounts := make(map[string]int)
	images := 0
	for i := range echos {
		t := time.Unix(echos[i].CreatedAt, 0).In(loc)
		monthCounts[t.Format("2006-01")]++
		activeDays[t.Format("2006-01-02")] = struct{}{}
		images += imageCountOf(echos[i])
		for _, tag := range echos[i].Tags {
			if n := strings.TrimSpace(tag.Name); n != "" {
				tagCounts[n]++
			}
		}
	}
	months := sortedMonthCounts(monthCounts)
	topMonth, topMonthN := topEntry(months)
	topTags := topNCounts(tagCounts, statsTopTags)

	zh := localeIsZH(locale)
	var b strings.Builder
	if zh {
		fmt.Fprintf(&b, "区间内 Echo 精确统计（数据来自数据库统计，请直接采用，不要另行估算）：\n")
		fmt.Fprintf(&b, "- 总条数：%d 条；活跃天数：%d 天；配图：%d 张\n", returned, len(activeDays), images)
		if topMonthN > 0 {
			fmt.Fprintf(&b, "- 最活跃月份：%s（%d 条）\n", topMonth, topMonthN)
		}
		b.WriteString("- 按月分布：" + joinMonthCounts(months, "条") + "\n")
		b.WriteString("- 常用标签 Top：" + joinTagCounts(topTags, "无") + "\n")
		if truncated {
			fmt.Fprintf(&b, "注意：区间内共命中 %d 条，超过单次统计上限，以上仅统计最近 %d 条，请在回答中说明这一点。\n", total, returned)
		}
		return strings.TrimSpace(b.String())
	}
	fmt.Fprintf(&b, "Exact stats for the range (from the database — use these numbers directly, do not estimate):\n")
	fmt.Fprintf(&b, "- Total: %d echos; active days: %d; images: %d\n", returned, len(activeDays), images)
	if topMonthN > 0 {
		fmt.Fprintf(&b, "- Most active month: %s (%d)\n", topMonth, topMonthN)
	}
	b.WriteString("- By month: " + joinMonthCounts(months, "") + "\n")
	b.WriteString("- Top tags: " + joinTagCounts(topTags, "none") + "\n")
	if truncated {
		fmt.Fprintf(&b, "Note: %d echos matched in total, exceeding the per-run cap; only the most recent %d are counted above — mention this in your answer.\n", total, returned)
	}
	return strings.TrimSpace(b.String())
}

type monthCount struct {
	Month string
	Count int
}

func sortedMonthCounts(m map[string]int) []monthCount {
	out := make([]monthCount, 0, len(m))
	for k, v := range m {
		out = append(out, monthCount{Month: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Month < out[j].Month })
	return out
}

func topEntry(months []monthCount) (string, int) {
	best, bestN := "", 0
	for _, mc := range months {
		if mc.Count > bestN {
			best, bestN = mc.Month, mc.Count
		}
	}
	return best, bestN
}

func topNCounts(m map[string]int, n int) []monthCount {
	out := make([]monthCount, 0, len(m))
	for k, v := range m {
		out = append(out, monthCount{Month: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Month < out[j].Month
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func joinMonthCounts(months []monthCount, unit string) string {
	if len(months) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(months))
	for _, mc := range months {
		if unit != "" {
			parts = append(parts, fmt.Sprintf("%s (%d%s)", mc.Month, mc.Count, unit))
		} else {
			parts = append(parts, fmt.Sprintf("%s (%d)", mc.Month, mc.Count))
		}
	}
	return strings.Join(parts, "、")
}

func joinTagCounts(tags []monthCount, emptyLabel string) string {
	if len(tags) == 0 {
		return emptyLabel
	}
	parts := make([]string, 0, len(tags))
	for _, tc := range tags {
		parts = append(parts, fmt.Sprintf("#%s (%d)", tc.Month, tc.Count))
	}
	return strings.Join(parts, "、")
}
