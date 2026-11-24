// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/agent"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	logUtil "github.com/lin-snow/ech0/pkg/log"
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
	images := newImageBudget()
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
			tagIDs, err := resolveTagIDs(allTags, a.Tags, locale)
			if err != nil {
				return agent.ToolOutput{}, err
			}
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
			if structured {
				results, total, execErr = s.queryEchos(ctx, user.ID, a.Query, tagIDs, from, to, topK)
			} else {
				results, total, execErr = s.searchByQuery(ctx, user.ID, a.Query, topK)
			}
			if execErr != nil {
				return agent.ToolOutput{}, execErr
			}
			exts, images := s.enrichHits(ctx, results, multimodal, images)
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

// searchByQuery answers a free-text query: semantically when an index is
// there, by keyword otherwise, and by both when the index comes up short.
//
// The semantic index is an accelerator, not the source of truth. It can be
// switched on but never built, mid-rebuild, or built for another model, and
// then an error or an empty hit list says nothing about whether the Echos
// exist — answering "no such record" from it is wrong when a LIKE query would
// find them. So any shortfall is topped up from keyword search, deduplicated,
// with the semantic hits kept first.
func (s *CopilotService) searchByQuery(ctx context.Context, userID, query string, topK int) ([]embeddingModel.SearchResult, int64, error) {
	var semantic []embeddingModel.SearchResult
	if s.embedding.Enabled(ctx) {
		hits, err := s.embedding.Search(ctx, query, topK, userID)
		if err != nil {
			logUtil.GetLogger().Warn("semantic search failed, falling back to keyword search",
				slog.String("module", "copilot"), logUtil.Err(err))
		}
		semantic = hits
	}
	if len(semantic) >= topK {
		return semantic, 0, nil
	}

	keyword, total, err := s.queryEchos(ctx, userID, query, nil, 0, 0, topK)
	if err != nil {
		if len(semantic) > 0 {
			return semantic, 0, nil
		}
		return nil, 0, err
	}
	if len(semantic) == 0 {
		return keyword, total, nil
	}
	merged := semantic
	have := make(map[string]bool, len(semantic))
	for _, r := range semantic {
		have[r.EchoID] = true
	}
	for _, r := range keyword {
		if len(merged) >= topK {
			break
		}
		if !have[r.EchoID] {
			merged = append(merged, r)
		}
	}
	return merged, 0, nil
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
	var total int64
	if page.Total != nil {
		total = *page.Total
	}
	return results, total, nil
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

// resolveTagIDs maps tag names to IDs, and refuses rather than guesses when a
// name matches nothing.
//
// Dropping an unknown name would quietly widen the query: tags=["阅读"] against
// a site that calls it "读书" becomes a search over everything, and the model
// reports the count for all Echos as the count for #阅读. An error goes back to
// the model instead, which can pick an existing tag or search without one.
func resolveTagIDs(allTags []echoModel.Tag, names []string, locale string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	byName := make(map[string]string, len(allTags))
	for _, t := range allTags {
		byName[strings.ToLower(strings.TrimSpace(t.Name))] = t.ID
	}
	ids := make([]string, 0, len(names))
	var unknown []string
	for _, n := range names {
		n = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(n), "#"))
		if n == "" {
			continue
		}
		if id, ok := byName[strings.ToLower(n)]; ok {
			ids = append(ids, id)
		} else {
			unknown = append(unknown, n)
		}
	}
	if len(unknown) > 0 {
		return nil, errors.New(unknownTagsMessageFor(locale, unknown))
	}
	return ids, nil
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
