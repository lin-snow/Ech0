// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"strings"
	"testing"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

type pagingEchoSvc struct {
	EchoService
	all       []echoModel.Echo
	gotUserID string
}

const serverMaxPageSize = 40

func (f *pagingEchoSvc) QueryEchos(
	_ context.Context,
	dto commonModel.EchoQueryDto,
) (commonModel.PageQueryResult[[]echoModel.Echo], error) {
	f.gotUserID = dto.UserID
	ps := dto.PageSize
	if ps < 1 || ps > serverMaxPageSize {
		ps = serverMaxPageSize
	}
	page := max(dto.Page, 1)
	start := (page - 1) * ps
	total := int64(len(f.all))
	if start >= len(f.all) {
		return commonModel.PageQueryResult[[]echoModel.Echo]{Items: nil, Total: total}, nil
	}
	end := min(start+ps, len(f.all))
	return commonModel.PageQueryResult[[]echoModel.Echo]{Items: f.all[start:end], Total: total}, nil
}

func makeEchos(n int) []echoModel.Echo {
	echos := make([]echoModel.Echo, 0, n)
	base := time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)
	for i := range n {
		ts := base.Add(time.Duration(i) * 43 * time.Hour)
		echos = append(echos, echoModel.Echo{ID: string(rune('a' + i%26)), Content: "echo", CreatedAt: ts.Unix()})
	}
	for l, r := 0, len(echos)-1; l < r; l, r = l+1, r-1 {
		echos[l], echos[r] = echos[r], echos[l]
	}
	return echos
}

func TestCollectRange_PaginatesAll(t *testing.T) {
	const n = 102
	svc := &pagingEchoSvc{all: makeEchos(n)}
	s := &CopilotService{echoService: svc}

	echos, total, truncated, err := s.collectRange(context.Background(), "", nil, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != n {
		t.Fatalf("total = %d, want %d", total, n)
	}
	if len(echos) != n {
		t.Fatalf("collected %d echos, want %d（疑似只拿回第一页）", len(echos), n)
	}
	if truncated {
		t.Fatalf("truncated = true, want false（未触顶硬上限）")
	}
}

func TestCollectRange_ScopesByUser(t *testing.T) {
	svc := &pagingEchoSvc{all: makeEchos(3)}
	s := &CopilotService{echoService: svc}

	if _, _, _, err := s.collectRange(context.Background(), "user-42", nil, 0, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.gotUserID != "user-42" {
		t.Fatalf("QueryEchos saw UserID = %q, want %q（区间聚合未按作者收口）", svc.gotUserID, "user-42")
	}
}

func TestQueryEchos_ScopesByUser(t *testing.T) {
	svc := &pagingEchoSvc{all: makeEchos(3)}
	s := &CopilotService{echoService: svc}

	if _, _, err := s.queryEchos(context.Background(), "user-7", "三体", nil, 0, 0, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if svc.gotUserID != "user-7" {
		t.Fatalf("QueryEchos saw UserID = %q, want %q（点查未按作者收口）", svc.gotUserID, "user-7")
	}
}

func TestCollectRange_TruncatesAtCap(t *testing.T) {
	svc := &pagingEchoSvc{all: makeEchos(maxAggregateEchos + 50)}
	s := &CopilotService{echoService: svc}

	echos, total, truncated, err := s.collectRange(context.Background(), "", nil, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(echos) != maxAggregateEchos {
		t.Fatalf("collected %d, want cap %d", len(echos), maxAggregateEchos)
	}
	if total != int64(maxAggregateEchos+50) {
		t.Fatalf("total = %d, want %d", total, maxAggregateEchos+50)
	}
	if !truncated {
		t.Fatalf("truncated = false, want true（超过硬上限）")
	}
}

func TestChunkEchosByBudget(t *testing.T) {
	echos := makeEchos(50)
	budget := estimateTokens(formatEchoLine(echos[0], time.UTC)) * 7
	chunks := chunkEchosByBudget(echos, budget, time.UTC)
	if len(chunks) < 2 {
		t.Fatalf("got %d chunk(s), expected multiple", len(chunks))
	}

	seen := 0
	for _, ch := range chunks {
		if len(ch) == 0 {
			t.Fatalf("empty chunk")
		}
		tok := 0
		for _, e := range ch {
			tok += estimateTokens(formatEchoLine(e, time.UTC))
		}
		if len(ch) > 1 && tok > budget {
			t.Fatalf("multi-echo chunk exceeds budget: %d > %d", tok, budget)
		}
		seen += len(ch)
	}
	if seen != len(echos) {
		t.Fatalf("chunks cover %d echos, want %d（顺序/完整性被破坏）", seen, len(echos))
	}
}

func TestFormatEchosByMonth(t *testing.T) {
	echos := makeEchos(102)
	out := formatEchosByMonth(echos, time.UTC)
	for _, m := range []string{"## 2025-07", "## 2025-08", "## 2025-09", "## 2025-10", "## 2025-11", "## 2025-12"} {
		if !strings.Contains(out, m) {
			t.Fatalf("月份小标题缺失：%q\n%s", m, out)
		}
	}
}

func TestFormatEchoLine_Enrichment(t *testing.T) {
	withText := echoModel.Echo{
		Content:   "今天读完一本书",
		CreatedAt: time.Date(2025, 3, 4, 0, 0, 0, 0, time.UTC).Unix(),
		Tags:      []echoModel.Tag{{Name: "读书"}},
	}
	line := formatEchoLine(withText, time.UTC)
	if !strings.Contains(line, "#读书") || !strings.Contains(line, "今天读完一本书") {
		t.Fatalf("缺标签或正文：%q", line)
	}
}
