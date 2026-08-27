// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import (
	"strings"
	"testing"
)

func TestStripReasoning(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "无标签原样返回（仅去首尾空白）",
			in:   "  最近作者在读《三体》。  ",
			want: "最近作者在读《三体》。",
		},
		{
			name: "去掉单个 think 块",
			in:   "<think>先分析一下这些 echo……</think>\n\n作者最近在读《三体》。",
			want: "作者最近在读《三体》。",
		},
		{
			name: "跨行 think 块",
			in:   "<think>\n第一步：看标签\n第二步：归纳\n</think>作者状态不错。",
			want: "作者状态不错。",
		},
		{
			name: "多个 think 块",
			in:   "<think>a</think>正文一。<think>b</think>正文二。",
			want: "正文一。正文二。",
		},
		{
			name: "大小写无关与 thinking 变体",
			in:   "<Thinking>reason</Thinking>结论。",
			want: "结论。",
		},
		{
			name: "纯推理被全部剥离后为空",
			in:   "<think>只有推理没有答案</think>",
			want: "",
		},
		{
			name: "正文里不带标签的尖括号不受影响",
			in:   "比较 a < b 且 c > d 的关系。",
			want: "比较 a < b 且 c > d 的关系。",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripReasoning(tc.in); got != tc.want {
				t.Errorf("stripReasoning() = %q, want %q", got, tc.want)
			}
		})
	}
}

func runSplitter(chunks ...string) (answer, reasoning string) {
	r := &reasoningSplitter{}
	var ans, rea strings.Builder
	for _, c := range chunks {
		a, re := r.feed(c)
		ans.WriteString(a)
		rea.WriteString(re)
	}
	a, re := r.flush()
	ans.WriteString(a)
	rea.WriteString(re)
	return ans.String(), rea.String()
}

func TestReasoningSplitter(t *testing.T) {
	cases := []struct {
		name          string
		chunks        []string
		wantAnswer    string
		wantReasoning string
	}{
		{
			name:       "无标签全是答案",
			chunks:     []string{"晚饭", "吃啥都行"},
			wantAnswer: "晚饭吃啥都行",
		},
		{
			name:          "单片含完整 think 块",
			chunks:        []string{"<think>先想想</think>吃粥吧。"},
			wantAnswer:    "吃粥吧。",
			wantReasoning: "先想想",
		},
		{
			name:          "open 标签跨 chunk 切开",
			chunks:        []string{"<thi", "nk>推理", "中</think>答案"},
			wantAnswer:    "答案",
			wantReasoning: "推理中",
		},
		{
			name:          "close 标签跨 chunk 切开",
			chunks:        []string{"<think>推理</thi", "nk>答案来了"},
			wantAnswer:    "答案来了",
			wantReasoning: "推理",
		},
		{
			name:          "thinking 变体 + 大小写无关",
			chunks:        []string{"<Thinking>reason</THINKING>done"},
			wantAnswer:    "done",
			wantReasoning: "reason",
		},
		{
			name:          "逐字符喂入",
			chunks:        []string{"<", "t", "h", "i", "n", "k", ">", "r", "</", "t", "h", "i", "n", "k", ">", "A"},
			wantAnswer:    "A",
			wantReasoning: "r",
		},
		{
			name:          "未闭合 think 收尾归推理",
			chunks:        []string{"<think>没说完就断了"},
			wantReasoning: "没说完就断了",
		},
		{
			name:       "答案里裸露的小于号不误判为标签",
			chunks:     []string{"判断 a < b 是否成立"},
			wantAnswer: "判断 a < b 是否成立",
		},
		{
			name:          "think 块前后都有答案",
			chunks:        []string{"前言。<think>推理</think>结论。"},
			wantAnswer:    "前言。结论。",
			wantReasoning: "推理",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			answer, reasoning := runSplitter(tc.chunks...)
			if answer != tc.wantAnswer {
				t.Errorf("answer = %q, want %q", answer, tc.wantAnswer)
			}
			if reasoning != tc.wantReasoning {
				t.Errorf("reasoning = %q, want %q", reasoning, tc.wantReasoning)
			}
		})
	}
}
