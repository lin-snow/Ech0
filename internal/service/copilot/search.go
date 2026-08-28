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
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
)

const defaultTopK = 6

const maxTopK = 20

const largeWindowTopK = 10

const largeWindowThreshold = 256_000

func effectiveTopK(window, requested int) int {
	if requested > 0 {
		if requested > maxTopK {
			return maxTopK
		}
		return requested
	}
	if window >= largeWindowThreshold {
		return largeWindowTopK
	}
	return defaultTopK
}

type searchArgs struct {
	Query    string   `json:"query"`
	Tags     []string `json:"tags"`
	DateFrom string   `json:"date_from"`
	DateTo   string   `json:"date_to"`
	Limit    int      `json:"limit"`
}

func (s *CopilotService) searchEchosTool(allTags []echoModel.Tag, multimodal bool, locale string, loc *time.Location, window int, user chatUser) agent.Tool {
	return agent.Tool{
		Def: agent.ToolDef{
			Name:        "search_echos",
			Description: "检索用户过往发布的 Echo（微博客/碎碎念）。可用 query 做语义/关键词检索，并可选地用 tags（标签名）与 date_from/date_to（日期范围）做精确筛选；三者可组合，但至少提供其一。query 传精炼核心词，不要整句。每条结果形如「【1】(2026-01-02) id=019ce0ea-… 正文」——其中 id= 后面那串 UUID 是这条 Echo 的真实 ID，要改或要删时照抄它；【1】只是本次结果的编号，不能当 ID 用。",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"语义/关键词检索词，传与问题最相关的核心词（精炼，不要整句）；仅按标签或时间筛选时可省略"},"tags":{"type":"array","items":{"type":"string"},"description":"按标签名筛选（标签名而非ID），如 [\"读书\",\"旅行\"]；可用标签见系统提示"},"date_from":{"type":"string","description":"起始日期，格式 YYYY-MM-DD，含当天"},"date_to":{"type":"string","description":"结束日期，格式 YYYY-MM-DD，含当天"},"limit":{"type":"integer","description":"可选，返回条数（1~20），默认按上下文自动取值；需要更多结果再综合时可调大"}}}`),
		},
		Effect: agent.EffectRead,
		Run: func(ctx context.Context, args json.RawMessage) (agent.ToolOutput, error) {
			var a searchArgs
			_ = json.Unmarshal(args, &a)
			a.Query = strings.TrimSpace(a.Query)
			tagIDs := resolveTagIDs(allTags, a.Tags)
			from := parseDay(a.DateFrom, false, loc)
			to := parseDay(a.DateTo, true, loc)
			topK := effectiveTopK(window, a.Limit)

			structured := len(tagIDs) > 0 || from > 0 || to > 0
			if a.Query == "" && !structured {
				return agent.ToolOutput{}, errors.New("检索需要 query、tags 或日期范围至少其一")
			}

			var results []embeddingModel.SearchResult
			var total int64
			var execErr error
			switch {
			case structured:
				results, total, execErr = s.queryEchos(ctx, user.ID, a.Query, tagIDs, from, to, topK)
			case s.embedding.Enabled(ctx):
				results, execErr = s.embedding.Search(ctx, a.Query, topK, user.Username)
			default:
				results, total, execErr = s.queryEchos(ctx, user.ID, a.Query, nil, 0, 0, topK)
			}
			if execErr != nil {
				return agent.ToolOutput{}, execErr
			}
			exts, images := s.enrichHits(ctx, results, multimodal)
			content := formatSearchResults(results, exts, loc)
			if total > int64(len(results)) {
				content = searchCoverageNoteFor(locale, int(total), len(results)) + "\n" + content
			}
			return agent.ToolOutput{
				Content: content,
				Meta:    results,
				Images:  images,
			}, nil
		},
	}
}

func (s *CopilotService) queryEchos(ctx context.Context, userID, search string, tagIDs []string, from, to int64, limit int) ([]embeddingModel.SearchResult, int64, error) {
	if limit <= 0 {
		limit = defaultTopK
	}
	page, err := s.echoService.QueryEchos(ctx, commonModel.EchoQueryDto{
		Page:     1,
		PageSize: limit,
		Search:   search,
		TagIDs:   tagIDs,
		DateFrom: from,
		DateTo:   to,
		UserID:   userID,
	})
	if err != nil {
		return nil, 0, err
	}
	results := make([]embeddingModel.SearchResult, 0, len(page.Items))
	for i := range page.Items {
		results = append(results, echoToSearchResult(page.Items[i]))
	}
	return results, page.Total, nil
}

func echoToSearchResult(e echoModel.Echo) embeddingModel.SearchResult {
	return embeddingModel.SearchResult{
		EchoID:      e.ID,
		Content:     e.Content,
		Username:    e.Username,
		EchoCreated: e.CreatedAt,
		Distance:    0,
	}
}

func resolveTagIDs(allTags []echoModel.Tag, names []string) []string {
	if len(names) == 0 {
		return nil
	}
	byName := make(map[string]string, len(allTags))
	for _, t := range allTags {
		byName[strings.ToLower(t.Name)] = t.ID
	}
	ids := make([]string, 0, len(names))
	for _, n := range names {
		if id, ok := byName[strings.ToLower(strings.TrimSpace(n))]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func parseDay(s string, endOfDay bool, loc *time.Location) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if loc == nil {
		loc = time.UTC
	}
	t, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		return 0
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return t.Unix()
}

func searchHintOf(args json.RawMessage) string {
	var a searchArgs
	_ = json.Unmarshal(args, &a)
	parts := make([]string, 0, 3)
	if q := strings.TrimSpace(a.Query); q != "" {
		parts = append(parts, q)
	}
	for _, t := range a.Tags {
		if t = strings.TrimSpace(t); t != "" {
			parts = append(parts, "#"+t)
		}
	}
	if from, to := strings.TrimSpace(a.DateFrom), strings.TrimSpace(a.DateTo); from != "" || to != "" {
		parts = append(parts, from+"~"+to)
	}
	return strings.Join(parts, " ")
}

// formatSearchResults renders hits for the model to read.
//
// The `id=` field is what makes update_echo and delete_echo usable at all: the
// bracketed 【N】 is a position within this one result set, and a model told to
// pass "the id" will quote it — so a request to edit the first hit arrives as
// id "1" and resolves to nothing. The primary key is spelled out instead, and
// verbatim, because it is the only handle that survives the next round.
func formatSearchResults(results []embeddingModel.SearchResult, exts map[string]string, loc *time.Location) string {
	if len(results) == 0 {
		return "（没有检索到相关的 Echo）"
	}
	if loc == nil {
		loc = time.UTC
	}
	var b strings.Builder
	for i, r := range results {
		day := time.Unix(r.EchoCreated, 0).In(loc).Format("2006-01-02")
		parts := []string{fmt.Sprintf("【%d】(%s) id=%s", i+1, day, r.EchoID)}
		if c := strings.TrimSpace(r.Content); c != "" {
			parts = append(parts, c)
		}
		if ext := exts[r.EchoID]; ext != "" {
			parts = append(parts, ext)
		}
		b.WriteString(strings.Join(parts, " "))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
