// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

type stubEchoSvc struct {
	EchoService
	queryFn   func(dto commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error)
	getByIDFn func(id string) (*echoModel.Echo, error)
	tags      []echoModel.Tag
}

func (f *stubEchoSvc) QueryEchos(_ context.Context, dto commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
	if f.queryFn != nil {
		return f.queryFn(dto)
	}
	return commonModel.PageQueryResult[[]echoModel.Echo]{}, nil
}

func (f *stubEchoSvc) GetEchoById(_ context.Context, id string) (*echoModel.Echo, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(id)
	}
	return nil, nil
}

func (f *stubEchoSvc) GetAllTags() ([]echoModel.Tag, error) { return f.tags, nil }

type stubEmbeddingSvc struct {
	EmbeddingService
	enabled    bool
	searchFn   func(query string, k int, author string) ([]embeddingModel.SearchResult, error)
	gotAuthor  string
	gotQuery   string
	searchSeen bool
}

func (f *stubEmbeddingSvc) Enabled(_ context.Context) bool { return f.enabled }

func (f *stubEmbeddingSvc) Search(_ context.Context, query string, k int, author string) ([]embeddingModel.SearchResult, error) {
	f.searchSeen = true
	f.gotAuthor = author
	f.gotQuery = query
	if f.searchFn != nil {
		return f.searchFn(query, k, author)
	}
	return nil, nil
}

type stubUserReader struct {
	user userModel.User
	err  error
}

func (f *stubUserReader) GetUserByID(_ string) (userModel.User, error) { return f.user, f.err }

func singlePage(items []echoModel.Echo, total int64) func(commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
	return func(dto commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
		if dto.Page > 1 {
			return commonModel.PageQueryResult[[]echoModel.Echo]{Items: nil, Total: total}, nil
		}
		return commonModel.PageQueryResult[[]echoModel.Echo]{Items: items, Total: total}, nil
	}
}

func mustArgs(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}
	return b
}

func ts(date string) int64 {
	t, _ := time.Parse("2006-01-02", date)
	return t.Unix()
}

func TestEffectiveTopK(t *testing.T) {
	cases := []struct {
		name              string
		window, requested int
		want              int
	}{
		{"explicit within bound", 0, 5, 5},
		{"explicit over max clamps", 0, 999, maxTopK},
		{"no request large window", largeWindowThreshold, 0, largeWindowTopK},
		{"no request small window", 1000, 0, defaultTopK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := effectiveTopK(c.window, c.requested); got != c.want {
				t.Fatalf("effectiveTopK(%d,%d) = %d, want %d", c.window, c.requested, got, c.want)
			}
		})
	}
}

func newSearchUser() chatUser { return chatUser{ID: "u1", Username: "alice"} }

func TestSearchEchosTool_NeedsQueryOrFilter(t *testing.T) {
	s := &CopilotService{echoService: &stubEchoSvc{}, embedding: &stubEmbeddingSvc{}}
	tool := s.searchEchosTool(nil, false, "zh-CN", time.UTC, 0, newSearchUser())

	_, err := tool.Execute(context.Background(), mustArgs(t, searchArgs{}))
	if err == nil || !strings.Contains(err.Error(), "检索需要 query") {
		t.Fatalf("want missing-criteria error, got %v", err)
	}
}

func TestSearchEchosTool_StructuredWithCoverageNote(t *testing.T) {
	items := []echoModel.Echo{
		{ID: "e1", Content: "读了三体", CreatedAt: ts("2026-01-05")},
		{ID: "e2", Content: "读了球状闪电", CreatedAt: ts("2026-01-09")},
	}
	echoSvc := &stubEchoSvc{
		queryFn: singlePage(items, 10),
		getByIDFn: func(id string) (*echoModel.Echo, error) {
			return &echoModel.Echo{
				ID:        id,
				Extension: &echoModel.EchoExtension{Type: echoModel.Extension_MUSIC, Payload: map[string]any{"url": "https://song"}},
			}, nil
		},
	}
	s := &CopilotService{echoService: echoSvc, embedding: &stubEmbeddingSvc{}}
	tool := s.searchEchosTool(nil, false, "zh-CN", time.UTC, 0, newSearchUser())

	out, err := tool.Execute(context.Background(), mustArgs(t, searchArgs{
		Query: "三体", DateFrom: "2026-01-01", DateTo: "2026-01-31",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results, ok := out.Meta.([]embeddingModel.SearchResult)
	if !ok || len(results) != 2 {
		t.Fatalf("Meta should be 2 SearchResults, got %#v", out.Meta)
	}
	if !strings.Contains(out.Content, "共命中 10 条") {
		t.Fatalf("coverage note missing: %q", out.Content)
	}
	if !strings.Contains(out.Content, "读了三体") || !strings.Contains(out.Content, "[音乐分享]") {
		t.Fatalf("content should include echo text + folded extension: %q", out.Content)
	}
}

func TestSearchEchosTool_SemanticPath(t *testing.T) {
	emb := &stubEmbeddingSvc{
		enabled: true,
		searchFn: func(query string, k int, author string) ([]embeddingModel.SearchResult, error) {
			return []embeddingModel.SearchResult{
				{EchoID: "e1", Content: "向量命中", EchoCreated: ts("2026-02-01")},
			}, nil
		},
	}
	echoSvc := &stubEchoSvc{getByIDFn: func(id string) (*echoModel.Echo, error) {
		return &echoModel.Echo{ID: id}, nil
	}}
	s := &CopilotService{echoService: echoSvc, embedding: emb}
	tool := s.searchEchosTool(nil, false, "zh-CN", time.UTC, 0, newSearchUser())

	out, err := tool.Execute(context.Background(), mustArgs(t, searchArgs{Query: "意识"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !emb.searchSeen {
		t.Fatalf("expected embedding.Search to be used for pure-query path")
	}
	if emb.gotAuthor != "alice" {
		t.Fatalf("semantic search must scope by username, got author=%q", emb.gotAuthor)
	}
	results, ok := out.Meta.([]embeddingModel.SearchResult)
	if !ok || len(results) != 1 || results[0].EchoID != "e1" {
		t.Fatalf("unexpected results: %#v", out.Meta)
	}
}

func TestSearchEchosTool_DefaultSQLFallback(t *testing.T) {
	var gotSearch string
	echoSvc := &stubEchoSvc{
		queryFn: func(dto commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
			gotSearch = dto.Search
			return commonModel.PageQueryResult[[]echoModel.Echo]{
				Items: []echoModel.Echo{{ID: "e9", Content: "fallback", CreatedAt: ts("2026-03-01")}},
				Total: 1,
			}, nil
		},
		getByIDFn: func(id string) (*echoModel.Echo, error) { return &echoModel.Echo{ID: id}, nil },
	}
	s := &CopilotService{echoService: echoSvc, embedding: &stubEmbeddingSvc{enabled: false}}
	tool := s.searchEchosTool(nil, false, "zh-CN", time.UTC, 0, newSearchUser())

	out, err := tool.Execute(context.Background(), mustArgs(t, searchArgs{Query: "fallback"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotSearch != "fallback" {
		t.Fatalf("queryEchos should receive the search term, got %q", gotSearch)
	}
	if !strings.Contains(out.Content, "fallback") {
		t.Fatalf("content missing fallback echo: %q", out.Content)
	}
}

func TestSearchEchosTool_QueryErrorPropagates(t *testing.T) {
	wantErr := errors.New("db down")
	echoSvc := &stubEchoSvc{queryFn: func(commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
		return commonModel.PageQueryResult[[]echoModel.Echo]{}, wantErr
	}}
	s := &CopilotService{echoService: echoSvc, embedding: &stubEmbeddingSvc{}}
	tool := s.searchEchosTool(nil, false, "zh-CN", time.UTC, 0, newSearchUser())

	_, err := tool.Execute(context.Background(), mustArgs(t, searchArgs{Query: "x", DateFrom: "2026-01-01"}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("want propagated db error, got %v", err)
	}
}

func TestStatsOverviewTool_NeedsDateRange(t *testing.T) {
	s := &CopilotService{echoService: &stubEchoSvc{}}
	tool := s.statsOverviewTool(nil, "zh-CN", time.UTC, newSearchUser())

	_, err := tool.Execute(context.Background(), mustArgs(t, statsArgs{}))
	if err == nil || !strings.Contains(err.Error(), "stats_overview 需要") {
		t.Fatalf("want date-range error, got %v", err)
	}
}

func TestStatsOverviewTool_HappyPath(t *testing.T) {
	items := []echoModel.Echo{
		{ID: "a", CreatedAt: ts("2026-01-05"), Tags: []echoModel.Tag{{Name: "读书"}}},
		{ID: "b", CreatedAt: ts("2026-01-20"), Tags: []echoModel.Tag{{Name: "读书"}}},
		{ID: "c", CreatedAt: ts("2026-02-10")},
	}
	s := &CopilotService{echoService: &stubEchoSvc{queryFn: singlePage(items, int64(len(items)))}}
	tool := s.statsOverviewTool(nil, "zh-CN", time.UTC, newSearchUser())

	out, err := tool.Execute(context.Background(), mustArgs(t, statsArgs{DateFrom: "2026-01-01", DateTo: "2026-02-28"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.Content, "总条数：3 条") || !strings.Contains(out.Content, "#读书 (2)") {
		t.Fatalf("stats content unexpected: %q", out.Content)
	}
}

func TestStatsOverviewTool_CollectErrorPropagates(t *testing.T) {
	wantErr := errors.New("query boom")
	s := &CopilotService{echoService: &stubEchoSvc{queryFn: func(commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
		return commonModel.PageQueryResult[[]echoModel.Echo]{}, wantErr
	}}}
	tool := s.statsOverviewTool(nil, "zh-CN", time.UTC, newSearchUser())

	_, err := tool.Execute(context.Background(), mustArgs(t, statsArgs{DateFrom: "2026-01-01", DateTo: "2026-02-28"}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("want propagated error, got %v", err)
	}
}

func TestSummarizeEchosTool_NeedsDateRange(t *testing.T) {
	s := &CopilotService{echoService: &stubEchoSvc{}}
	tool := s.summarizeEchosTool(nil, settingModel.AgentSetting{}, "zh-CN", time.UTC, newSearchUser())

	_, err := tool.Execute(context.Background(), mustArgs(t, summarizeArgs{}))
	if err == nil || !strings.Contains(err.Error(), "summarize_echos 需要") {
		t.Fatalf("want date-range error, got %v", err)
	}
}

func TestSummarizeEchosTool_HappyPathFitsBudget(t *testing.T) {
	items := []echoModel.Echo{
		{ID: "a", Content: "一月读书", CreatedAt: ts("2026-01-05")},
		{ID: "b", Content: "二月旅行", CreatedAt: ts("2026-02-10")},
	}
	s := &CopilotService{echoService: &stubEchoSvc{queryFn: singlePage(items, int64(len(items)))}}
	tool := s.summarizeEchosTool(nil, settingModel.AgentSetting{}, "zh-CN", time.UTC, newSearchUser())

	out, err := tool.Execute(context.Background(), mustArgs(t, summarizeArgs{
		DateFrom: "2026-01-01", DateTo: "2026-02-28", Focus: "工作",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cov, ok := out.Meta.(aggregateCoverage)
	if !ok {
		t.Fatalf("Meta should be aggregateCoverage, got %#v", out.Meta)
	}
	if cov.Total != 2 || cov.Returned != 2 || cov.Buckets != 1 || cov.Truncated {
		t.Fatalf("unexpected coverage: %+v", cov)
	}
	if !strings.Contains(out.Content, "聚合材料") {
		t.Fatalf("content should carry aggregate material header: %q", out.Content)
	}
	if !strings.Contains(out.Content, "用户希望侧重：工作") {
		t.Fatalf("focus emphasis missing: %q", out.Content)
	}
	if !strings.Contains(out.Content, "一月读书") || !strings.Contains(out.Content, "二月旅行") {
		t.Fatalf("material should include the echos by month: %q", out.Content)
	}
}

func TestSummarizeEchosTool_CollectErrorPropagates(t *testing.T) {
	wantErr := errors.New("range boom")
	s := &CopilotService{echoService: &stubEchoSvc{queryFn: func(commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
		return commonModel.PageQueryResult[[]echoModel.Echo]{}, wantErr
	}}}
	tool := s.summarizeEchosTool(nil, settingModel.AgentSetting{}, "zh-CN", time.UTC, newSearchUser())

	_, err := tool.Execute(context.Background(), mustArgs(t, summarizeArgs{DateFrom: "2026-01-01", DateTo: "2026-02-28"}))
	if !errors.Is(err, wantErr) {
		t.Fatalf("want propagated error, got %v", err)
	}
}
