// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/kvstore"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRSSContext(t *testing.T, host string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "http://"+host+"/rss", nil)
	req.Host = host
	req.TLS = nil
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = req
	return ctx
}

// newRSSService builds the service over an in-memory settings store; passing a
// setting seeds it, omitting it exercises the pristine defaults.
func newRSSService(
	t *testing.T,
	repo commonService.CommonRepository,
	settings ...settingModel.SystemSetting,
) *commonService.CommonService {
	t.Helper()
	kv := kvstore.NewMemory()
	if len(settings) > 0 {
		require.NoError(t, coreSetting.System.Set(context.Background(), kv, settings[0]))
	}
	return commonService.NewCommonService(repo, helpers.NewTestCache(), kv)
}

// matchesRSSCacheKey asserts the key shape without pinning the identity hash.
func matchesRSSCacheKey(prefix string) any {
	return mock.MatchedBy(func(key string) bool {
		return strings.HasPrefix(key, prefix+":") && len(key) == len(prefix)+1+12
	})
}

func TestGenerateRSS_NormalFeed(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	echos := []echoModel.Echo{
		{
			ID:        "echo-1",
			Username:  "alice",
			Content:   "hello rss world",
			CreatedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC).Unix(),
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(matchesRSSCacheKey("rss:http:example.com")).Return().Once()

	ctx := newRSSContext(t, "example.com")
	doc, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	atom := doc.Atom

	assert.NotContains(t, atom, "xml-stylesheet", "浏览器已移除 XSLT，Feed 不应再引用样式表")
	assert.Contains(t, atom, "<title>Ech0</title>")
	assert.Contains(t, atom, "hello rss world")
	assert.Contains(t, atom, "alice")
	assert.Contains(t, atom, "http://example.com/echo/echo-1")

	assert.Contains(t, doc.HTML, "<title>Ech0 · RSS</title>")
	assert.Contains(t, doc.HTML, `rel="alternate" type="application/atom+xml" href="http://example.com/rss"`)
	assert.Contains(t, doc.HTML, "共 1 条动态")
	assert.Contains(t, doc.HTML, `datetime="2025-01-02T03:04:05Z"`)
	assert.NotContains(t, doc.HTML, "hello rss world", "落地页不重复 feed 正文")
	assert.NotContains(t, doc.HTML, "alice", "落地页不重复 feed 作者")
	assert.NotContains(t, doc.HTML, "http://example.com/echo/echo-1", "落地页不重复条目链接")
}

func TestGenerateRSS_TagHTMLEntityEscaping(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	echos := []echoModel.Echo{
		{
			ID:        "echo-xss",
			Username:  "mallory",
			Content:   "benign body",
			CreatedAt: time.Now().UTC().Unix(),
			Tags: []echoModel.Tag{
				{Name: `<script>alert(1)</script>`},
			},
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Once()

	ctx := newRSSContext(t, "example.com")
	doc, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	atom := doc.Atom

	assert.NotContains(t, atom, "<script>", "RSS 不应含原始 script 标签")
	assert.NotContains(t, atom, "&lt;script&gt;", "标签名必须先做 HTML 实体转义，杜绝单层转义形态")
	assert.Contains(t, atom, "&amp;lt;script&amp;gt;", "应为双层转义，证明 HTML 实体转义已生效")
}

func TestGenerateRSS_RendersEchoImages(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	echos := []echoModel.Echo{
		{
			ID:        "echo-img",
			Username:  "bob",
			Content:   "look at this",
			CreatedAt: time.Now().UTC().Unix(),
			EchoFiles: []echoModel.EchoFile{
				{File: fileModel.File{URL: "http://example.com/files/pic.png"}},
			},
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Once()

	ctx := newRSSContext(t, "example.com")
	doc, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	atom := doc.Atom

	assert.Contains(t, atom, "http://example.com/files/pic.png", "图片直链应出现在条目描述里")
	assert.Contains(t, atom, "look at this")
}

func TestGenerateRSS_RendersMediaByCategory(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	echos := []echoModel.Echo{
		{
			ID:        "echo-media",
			Username:  "carol",
			Content:   "mixed media",
			CreatedAt: time.Now().UTC().Unix(),
			EchoFiles: []echoModel.EchoFile{
				{File: fileModel.File{Category: "image", URL: "http://example.com/files/pic.png"}},
				{File: fileModel.File{Category: "video", URL: "http://example.com/files/clip.mp4"}},
				{File: fileModel.File{Category: "audio", URL: "http://example.com/files/song.mp3"}},
				{File: fileModel.File{Category: "pdf", URL: "http://example.com/files/doc.pdf", Name: "doc.pdf"}},
			},
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Once()

	ctx := newRSSContext(t, "example.com")
	doc, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	atom := doc.Atom

	assert.Contains(t, atom, `&lt;img src=&#34;http://example.com/files/pic.png&#34;`, "图片应渲染为 <img>")
	assert.Contains(t, atom, `&lt;video controls=&#34;&#34; src=&#34;http://example.com/files/clip.mp4&#34;`, "视频应渲染为 <video controls>")
	assert.Contains(t, atom, `&lt;a href=&#34;http://example.com/files/clip.mp4&#34;&gt;打开视频&lt;/a&gt;`, "视频应内嵌链接兜底")
	assert.Contains(t, atom, `&lt;audio controls=&#34;&#34; src=&#34;http://example.com/files/song.mp3&#34;`, "音频应渲染为 <audio controls>")
	assert.Contains(t, atom, `&lt;a href=&#34;http://example.com/files/song.mp3&#34;&gt;打开音频&lt;/a&gt;`, "音频应内嵌链接兜底")
	assert.Contains(t, atom, `&lt;a href=&#34;http://example.com/files/doc.pdf&#34;&gt;doc.pdf&lt;/a&gt;`, "普通文件应渲染为下载链接")
}

func TestGenerateRSS_MediaFieldEscaping(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	echos := []echoModel.Echo{
		{
			ID:        "echo-evil",
			Username:  "mallory",
			Content:   "benign",
			CreatedAt: time.Now().UTC().Unix(),
			EchoFiles: []echoModel.EchoFile{
				{File: fileModel.File{
					Category: "file",
					URL:      `http://x/"><script>alert(1)</script>`,
					Name:     `<script>alert(2)</script>`,
				}},
			},
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Once()

	ctx := newRSSContext(t, "example.com")
	doc, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	atom := doc.Atom

	assert.NotContains(t, atom, "<script>", "URL/文件名注入的原始 script 标签不得出现")
	assert.NotContains(t, atom, "&lt;script&gt;", "媒体字段必须先做 HTML 实体转义，杜绝单层转义形态")
}

func TestGenerateRSS_ReadThrough(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	echos := []echoModel.Echo{{ID: "e1", Username: "u", Content: "c", CreatedAt: time.Now().UTC().Unix()}}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(matchesRSSCacheKey("rss:http:example.com")).Return().Once()

	ctx := newRSSContext(t, "example.com")

	first, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	second, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.Equal(t, first, second, "缓存命中应返回与首回相同的内容")
}

func TestGenerateRSS_RepositoryError(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(nil, assert.AnError).Once()

	ctx := newRSSContext(t, "example.com")
	doc, err := svc.GenerateRSS(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, doc.Atom)
	assert.Empty(t, doc.HTML)
}

func TestGenerateRSS_AbsolutizesRelativeContentURLs(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	echos := []echoModel.Echo{
		{
			ID:        "echo-rel",
			Username:  "dana",
			Content:   "[原文](/echo/abc) ![图](/files/a.png) [外链](https://example.org/x) [邮件](mailto:a@example.org)",
			CreatedAt: time.Now().UTC().Unix(),
			EchoFiles: []echoModel.EchoFile{
				{File: fileModel.File{Category: "image", URL: "/api/files/images/pic.png"}},
			},
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Once()

	doc, err := svc.GenerateRSS(newRSSContext(t, "example.com"))
	require.NoError(t, err)

	assert.Contains(t, doc.Atom, `<content type="html">`, "Reeder 等阅读器按 content 渲染正文")
	assert.Contains(t, doc.Atom, `src=&#34;http://example.com/api/files/images/pic.png&#34;`, "本地附件相对路径需补成绝对地址")
	assert.Contains(t, doc.Atom, `href=&#34;http://example.com/echo/abc&#34;`)
	assert.Contains(t, doc.Atom, `src=&#34;http://example.com/files/a.png&#34;`)
	assert.Contains(t, doc.Atom, `href=&#34;https://example.org/x&#34;`, "绝对地址保持不变")
	assert.Contains(t, doc.Atom, `href=&#34;mailto:a@example.org&#34;`, "非 HTTP 协议保持不变")
	assert.NotContains(t, doc.Atom, `src=&#34;/api/files`, "feed 正文不得保留相对引用")
	assert.NotContains(t, doc.Atom, `href=&#34;/echo/`, "feed 正文不得保留相对引用")
}

func TestGenerateRSS_EmptyFeedLandingPage(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo)

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(nil, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(matchesRSSCacheKey("rss:http:example.com")).Return().Once()

	doc, err := svc.GenerateRSS(newRSSContext(t, "example.com"))
	require.NoError(t, err)

	assert.Contains(t, doc.HTML, "暂无内容")
	assert.NotContains(t, doc.HTML, "最近更新", "无动态时不应伪造更新时间")
	assert.NotContains(t, doc.HTML, "条动态")
}

func TestGenerateRSS_UsesSiteIdentity(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := newRSSService(t, repo, settingModel.SystemSetting{
		SiteTitle:  "晨间笔记",
		ServerName: "My Ech0",
	})

	echos := []echoModel.Echo{{ID: "e1", Username: "alice", Content: "hi", CreatedAt: time.Now().UTC().Unix()}}
	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey(matchesRSSCacheKey("rss:http:example.com")).Return().Once()

	doc, err := svc.GenerateRSS(newRSSContext(t, "example.com"))
	require.NoError(t, err)
	atom := doc.Atom

	assert.Contains(t, atom, "<title>晨间笔记</title>")
	assert.Contains(t, atom, "My Ech0")
	assert.NotContains(t, atom, "<title>Ech0</title>")
	assert.Contains(t, doc.HTML, "<title>晨间笔记 · RSS</title>")
}

func TestGenerateRSS_RenameMissesCache(t *testing.T) {
	// Renaming the site must not keep serving the previous name: the identity is
	// part of the cache key, so the second render loads instead of hitting cache.
	repo := commonmock.NewMockCommonRepository(t)
	cached := helpers.NewTestCache()
	kv := kvstore.NewMemory()
	require.NoError(t, coreSetting.System.Set(context.Background(), kv, settingModel.SystemSetting{SiteTitle: "Old Name"}))
	svc := commonService.NewCommonService(repo, cached, kv)

	echos := []echoModel.Echo{{ID: "e1", Username: "alice", Content: "hi", CreatedAt: time.Now().UTC().Unix()}}
	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Twice()
	repo.EXPECT().TrackRSSCacheKey(mock.Anything).Return().Twice()

	ctx := newRSSContext(t, "example.com")
	before, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	require.Contains(t, before.Atom, "<title>Old Name</title>")
	require.Contains(t, before.HTML, "Old Name")

	require.NoError(t, coreSetting.System.Set(context.Background(), kv, settingModel.SystemSetting{SiteTitle: "New Name"}))

	after, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	assert.Contains(t, after.Atom, "<title>New Name</title>")
	assert.NotContains(t, after.Atom, "<title>Old Name</title>")
	assert.Contains(t, after.HTML, "New Name")
	assert.NotContains(t, after.HTML, "Old Name")
}
