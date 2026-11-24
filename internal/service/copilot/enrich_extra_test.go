// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"reflect"
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
)

type enrichEchoSvc struct {
	EchoService
	echo *echoModel.Echo
}

func (f *enrichEchoSvc) GetEchoById(_ context.Context, _ string) (*echoModel.Echo, error) {
	return f.echo, nil
}

func TestEnrichHits_MediaCategoriesToFiles(t *testing.T) {
	echoFile := func(contentType string) echoModel.EchoFile {
		return echoModel.EchoFile{File: fileModel.File{ID: contentType, ContentType: contentType}}
	}
	svc := &enrichEchoSvc{echo: &echoModel.Echo{
		ID: "e1",
		EchoFiles: []echoModel.EchoFile{
			echoFile("image/png"),
			echoFile("video/mp4"),
			echoFile("audio/mpeg"),
			echoFile("application/pdf"),
			echoFile("application/octet-stream"),
		},
	}}
	s := &CopilotService{echoService: svc}

	results := []embeddingModel.SearchResult{{EchoID: "e1"}}
	_, images := s.enrichHits(context.Background(), results, false, nil)

	gotCats := make([]string, 0, len(results[0].Files))
	for _, f := range results[0].Files {
		gotCats = append(gotCats, fileModel.CategoryOf(f.ContentType))
	}
	want := []string{"image", "video", "audio"}
	if !reflect.DeepEqual(gotCats, want) {
		t.Fatalf("Files 类别 = %v, want %v（应带图片/视频/音频，排除 pdf/file）", gotCats, want)
	}
	if len(images) != 0 {
		t.Fatalf("多模态关闭时不应产出 base64 图片，got %d", len(images))
	}
}

func TestFormatExtension(t *testing.T) {
	ext := func(typ string, kv map[string]any) *echoModel.EchoExtension {
		return &echoModel.EchoExtension{Type: typ, Payload: kv}
	}
	cases := []struct {
		name string
		in   *echoModel.EchoExtension
		want string
	}{
		{"nil", nil, ""},
		{"music", ext(echoModel.Extension_MUSIC, map[string]any{"url": "https://song"}), "[音乐分享] https://song"},
		{"music missing url", ext(echoModel.Extension_MUSIC, map[string]any{}), ""},
		{"video", ext(echoModel.Extension_VIDEO, map[string]any{"videoId": "abc"}), "[视频分享] 视频ID abc"},
		{"github", ext(echoModel.Extension_GITHUBPROJ, map[string]any{"repoUrl": "https://gh"}), "[GitHub 项目] https://gh"},
		{"website title+site", ext(echoModel.Extension_WEBSITE, map[string]any{"title": "T", "site": "S"}), "[网站] T S"},
		{"website site only", ext(echoModel.Extension_WEBSITE, map[string]any{"site": "S"}), "[网站] S"},
		{"website title only", ext(echoModel.Extension_WEBSITE, map[string]any{"title": "T"}), "[网站] T"},
		{"website empty", ext(echoModel.Extension_WEBSITE, map[string]any{}), ""},
		{"location", ext(echoModel.Extension_LOCATION, map[string]any{"placeholder": "北京"}), "[位置] 北京"},
		{"tweet url+user", ext(echoModel.Extension_TWEET, map[string]any{"url": "https://x", "username": "bob"}), "[X 推文] @bob https://x"},
		{"tweet url only", ext(echoModel.Extension_TWEET, map[string]any{"url": "https://x"}), "[X 推文] https://x"},
		{"unknown type", ext("WHATEVER", map[string]any{"url": "x"}), ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatExtension(c.in); got != c.want {
				t.Fatalf("formatExtension = %q, want %q", got, c.want)
			}
		})
	}
}

func TestLoadImagePart_NonStorageBranches(t *testing.T) {
	s := &CopilotService{}
	ref := func(u string) *string { return &u }

	t.Run("ref image uses RefURL", func(t *testing.T) {
		part, ok := s.loadImagePart(context.Background(), fileModel.File{
			Kind:        string(fileModel.KindRef),
			RefURL:      ref("https://img/x.png"),
			ContentType: "image/png",
		})
		if !ok {
			t.Fatalf("ref image should produce an ImagePart")
		}
		if part.URL != "https://img/x.png" || part.MediaType != "image/png" {
			t.Fatalf("unexpected image part: %+v", part)
		}
	})

	t.Run("ref without RefURL is skipped", func(t *testing.T) {
		if _, ok := s.loadImagePart(context.Background(), fileModel.File{
			Kind:        string(fileModel.KindRef),
			ContentType: "image/png",
		}); ok {
			t.Fatalf("ref image without RefURL should be skipped")
		}
	})

	t.Run("non-image is skipped", func(t *testing.T) {
		if _, ok := s.loadImagePart(context.Background(), fileModel.File{
			ContentType: "audio/mpeg",
		}); ok {
			t.Fatalf("non-image file should be skipped")
		}
	})
}

func TestPlanContext(t *testing.T) {
	const fixed = 3_000 // system prompt + tool declarations

	for _, window := range []int{0, 8_000, 32_000, 128_000, 1_000_000} {
		p := planContext(window, fixed)
		w := window
		if w == 0 {
			w = defaultContextWindow
		}
		if p.Input >= w {
			t.Fatalf("window %d: input %d leaves no room for output", window, p.Input)
		}
		if fixed+p.History+p.Material > p.Input && p.Material > minAggregateBudget {
			t.Fatalf("window %d: history %d + material %d + fixed %d exceed input %d — the loop would trim the summary",
				window, p.History, p.Material, fixed, p.Input)
		}
	}

	if planContext(0, fixed).Input >= 256_000 {
		t.Fatalf("an unset window must not be assumed huge")
	}
	small := planContext(1_000, fixed)
	if small.Material != minAggregateBudget {
		t.Fatalf("a window smaller than the prompt should floor the material budget, got %d", small.Material)
	}
	if big := planContext(1_000_000, fixed); big.History != maxHistoryTokens {
		t.Fatalf("history should cap at %d, got %d", maxHistoryTokens, big.History)
	}
}

func TestRunStringsFor(t *testing.T) {
	zh := runStringsFor("zh-CN")
	if zh.ToolError == "" || zh.UnknownTool == "" {
		t.Fatalf("zh run strings should be populated: %+v", zh)
	}
	en := runStringsFor("en-US")
	if en.UnknownTool == zh.UnknownTool {
		t.Fatalf("en/zh run strings should differ: %q vs %q", en.UnknownTool, zh.UnknownTool)
	}
}
