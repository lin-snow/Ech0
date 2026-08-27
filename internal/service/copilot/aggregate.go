// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/internal/storage"
)

const (
	aggregatePageSize = 100
	maxAggregateEchos = 5000
)

type summarizeArgs struct {
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
	Tags     []string `json:"tags"`
	Focus    string   `json:"focus"`
}

type aggregateCoverage struct {
	Total     int64 `json:"total"`
	Returned  int   `json:"returned"`
	Buckets   int   `json:"buckets"`
	Truncated bool  `json:"truncated"`
}

func (s *CopilotService) summarizeEchosTool(
	allTags []echoModel.Tag,
	setting settingModel.AgentSetting,
	locale string,
	loc *time.Location,
	user chatUser,
) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "summarize_echos",
			Description: "聚合某时间区间内的【全部】Echo，用于生成跨度较长的总结/回顾（如年终、年度、季度、月度总结）。它会覆盖区间内所有记录而非只采样几条，正是「帮我写年终/年度总结」这类需求该用的工具——这类请求请直接用它，不要先用 search_echos 采样。必须提供 date_from 与 date_to；可选 tags（按标签名限定主题）、focus（侧重点，如“工作”“读书”“心情”）。",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"date_from":{"type":"string","description":"起始日期，格式 YYYY-MM-DD，含当天"},"date_to":{"type":"string","description":"结束日期，格式 YYYY-MM-DD，含当天"},"tags":{"type":"array","items":{"type":"string"},"description":"可选，按标签名限定主题；可用标签见系统提示"},"focus":{"type":"string","description":"可选，总结的侧重点，如“工作”“读书”“心情”"}},"required":["date_from","date_to"]}`),
		},
		Execute: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var a summarizeArgs
			_ = json.Unmarshal(args, &a)
			from := parseDay(a.DateFrom, false, loc)
			to := parseDay(a.DateTo, true, loc)
			if from == 0 && to == 0 {
				return agent.ToolOutput{}, errors.New("summarize_echos 需要 date_from 与 date_to 指定时间区间")
			}
			tagIDs := resolveTagIDs(allTags, a.Tags)

			echos, total, truncated, err := s.collectRange(ctx, user.ID, tagIDs, from, to)
			if err != nil {
				return agent.ToolOutput{}, err
			}
			reverseEchos(echos)

			material, buckets, err := s.mapReduceSummary(ctx, setting, locale, echos, aggregateBudgetTokens(setting), loc)
			if err != nil {
				return agent.ToolOutput{}, err
			}

			cov := aggregateCoverage{
				Total:     total,
				Returned:  len(echos),
				Buckets:   buckets,
				Truncated: truncated,
			}
			header := aggregateMaterialHeaderFor(locale, int(total), len(echos), buckets, truncated)
			if focus := strings.TrimSpace(a.Focus); focus != "" {
				if localeIsZH(locale) {
					header += "（用户希望侧重：" + focus + "）"
				} else {
					header += " (User wants emphasis on: " + focus + ")"
				}
			}
			return agent.ToolOutput{
				Content: header + "\n\n" + material,
				Meta:    cov,
			}, nil
		},
	}
}

func (s *CopilotService) collectRange(
	ctx context.Context,
	userID string,
	tagIDs []string,
	from, to int64,
) (echos []echoModel.Echo, total int64, truncated bool, err error) {
	for page := 1; ; page++ {
		res, qErr := s.echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
			Page:     page,
			PageSize: aggregatePageSize,
			TagIDs:   tagIDs,
			DateFrom: from,
			DateTo:   to,
			UserID:   userID,
		})
		if qErr != nil {
			return nil, 0, false, qErr
		}
		total = res.Total
		for i := range res.Items {
			echos = append(echos, res.Items[i])
			if len(echos) >= maxAggregateEchos {
				return echos, total, total > int64(len(echos)), nil
			}
		}
		if len(res.Items) == 0 || int64(len(echos)) >= total {
			return echos, total, false, nil
		}
	}
}

func reverseEchos(echos []echoModel.Echo) {
	for l, r := 0, len(echos)-1; l < r; l, r = l+1, r-1 {
		echos[l], echos[r] = echos[r], echos[l]
	}
}

func (s *CopilotService) mapReduceSummary(
	ctx context.Context,
	setting settingModel.AgentSetting,
	locale string,
	echos []echoModel.Echo,
	budget int,
	loc *time.Location,
) (content string, buckets int, err error) {
	full := formatEchosByMonth(echos, loc)
	if estimateTokens(full) <= budget {
		return full, 1, nil
	}

	chunks := chunkEchosByBudget(echos, budget, loc)
	var b strings.Builder
	for _, ch := range chunks {
		digest, mErr := agent.Generate(ctx, setting, []agent.Message{
			{Role: agent.RoleSystem, Content: aggregateMapPromptFor(locale)},
			{Role: agent.RoleUser, Content: formatEchosByMonth(ch, loc)},
		}, false, nil)
		if mErr != nil {
			return "", 0, mErr
		}
		fmt.Fprintf(&b, "【%s】\n%s\n\n", dateSpanOf(ch, loc), strings.TrimSpace(digest))
	}
	joined := strings.TrimSpace(b.String())

	if estimateTokens(joined) > budget {
		reduced, rErr := agent.Generate(ctx, setting, []agent.Message{
			{Role: agent.RoleSystem, Content: aggregateReducePromptFor(locale)},
			{Role: agent.RoleUser, Content: joined},
		}, false, nil)
		if rErr != nil {
			return "", 0, rErr
		}
		joined = strings.TrimSpace(reduced)
	}

	return joined, len(chunks), nil
}

func chunkEchosByBudget(echos []echoModel.Echo, budget int, loc *time.Location) [][]echoModel.Echo {
	var chunks [][]echoModel.Echo
	var cur []echoModel.Echo
	curTokens := 0
	for _, e := range echos {
		t := estimateTokens(formatEchoLine(e, loc))
		if len(cur) > 0 && curTokens+t > budget {
			chunks = append(chunks, cur)
			cur = nil
			curTokens = 0
		}
		cur = append(cur, e)
		curTokens += t
	}
	if len(cur) > 0 {
		chunks = append(chunks, cur)
	}
	return chunks
}

func formatEchosByMonth(echos []echoModel.Echo, loc *time.Location) string {
	if len(echos) == 0 {
		return "（该区间内没有 Echo）"
	}
	counts := make(map[string]int, 12)
	for i := range echos {
		counts[monthOf(echos[i], loc)]++
	}
	var b strings.Builder
	curMonth := ""
	for i := range echos {
		m := monthOf(echos[i], loc)
		if m != curMonth {
			if curMonth != "" {
				b.WriteString("\n")
			}
			fmt.Fprintf(&b, "## %s (%d)\n", m, counts[m])
			curMonth = m
		}
		b.WriteString(formatEchoLine(echos[i], loc))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func formatEchoLine(e echoModel.Echo, loc *time.Location) string {
	day := time.Unix(e.CreatedAt, 0).In(loc).Format("2006-01-02")
	parts := []string{"(" + day + ")"}
	content := strings.TrimSpace(e.Content)
	if content != "" {
		parts = append(parts, content)
	}
	if tags := tagLabels(e.Tags); tags != "" {
		parts = append(parts, tags)
	}
	if ext := formatExtension(e.Extension); ext != "" {
		parts = append(parts, ext)
	}
	if n := imageCountOf(e); n > 0 {
		if content == "" {
			parts = append(parts, fmt.Sprintf("[img-only×%d]", n))
		} else {
			parts = append(parts, fmt.Sprintf("[img×%d]", n))
		}
	}
	return strings.Join(parts, " ")
}

func tagLabels(tags []echoModel.Tag) string {
	if len(tags) == 0 {
		return ""
	}
	parts := make([]string, 0, len(tags))
	for _, t := range tags {
		if n := strings.TrimSpace(t.Name); n != "" {
			parts = append(parts, "#"+n)
		}
	}
	return strings.Join(parts, " ")
}

func imageCountOf(e echoModel.Echo) int {
	n := 0
	for _, ef := range e.EchoFiles {
		if storage.NormalizeCategory(ef.File.Category).IsImageLike() {
			n++
		}
	}
	return n
}

func monthOf(e echoModel.Echo, loc *time.Location) string {
	return time.Unix(e.CreatedAt, 0).In(loc).Format("2006-01")
}

func dateSpanOf(echos []echoModel.Echo, loc *time.Location) string {
	if len(echos) == 0 {
		return ""
	}
	first := time.Unix(echos[0].CreatedAt, 0).In(loc).Format("2006-01-02")
	last := time.Unix(echos[len(echos)-1].CreatedAt, 0).In(loc).Format("2006-01-02")
	if first == last {
		return first
	}
	return first + " ~ " + last
}
