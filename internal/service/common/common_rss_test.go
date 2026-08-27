// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/cache"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeCache struct {
	mu   sync.Mutex
	data map[string]any
}

var _ cache.ICache[string, any] = (*fakeCache)(nil)

func newFakeCache() *fakeCache {
	return &fakeCache{data: make(map[string]any)}
}

func (c *fakeCache) Set(key string, value any, _ int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return true
}

func (c *fakeCache) SetWithTTL(key string, value any, cost int64, _ time.Duration) bool {
	return c.Set(key, value, cost)
}

func (c *fakeCache) Get(key string) (any, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[key]
	return v, ok, nil
}

func (c *fakeCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

func (c *fakeCache) Close() error { return nil }

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

func TestGenerateRSS_NormalFeed(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	echos := []echoModel.Echo{
		{
			ID:        "echo-1",
			Username:  "alice",
			Content:   "hello rss world",
			CreatedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC).Unix(),
		},
	}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey("rss:http:example.com").Return().Once()

	ctx := newRSSContext(t, "example.com")
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.Contains(t, atom, `<?xml-stylesheet type="text/xsl" href="/rss.xsl"?>`)
	assert.Contains(t, atom, "<title>Ech0</title>")
	assert.Contains(t, atom, "hello rss world")
	assert.Contains(t, atom, "alice")
	assert.Contains(t, atom, "http://example.com/echo/echo-1")
}

func TestGenerateRSS_TagHTMLEntityEscaping(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

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
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.NotContains(t, atom, "<script>", "RSS 不应含原始 script 标签")
	assert.NotContains(t, atom, "&lt;script&gt;", "标签名必须先做 HTML 实体转义，杜绝单层转义形态")
	assert.Contains(t, atom, "&amp;lt;script&amp;gt;", "应为双层转义，证明 HTML 实体转义已生效")
}

func TestGenerateRSS_RendersEchoImages(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

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
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.Contains(t, atom, "http://example.com/files/pic.png", "图片直链应出现在条目描述里")
	assert.Contains(t, atom, "look at this")
}

func TestGenerateRSS_RendersMediaByCategory(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

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
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.Contains(t, atom, `&lt;img src=&#34;http://example.com/files/pic.png&#34;`, "图片应渲染为 <img>")
	assert.Contains(t, atom, `&lt;video controls src=&#34;http://example.com/files/clip.mp4&#34;`, "视频应渲染为 <video controls>")
	assert.Contains(t, atom, `&lt;a href=&#34;http://example.com/files/clip.mp4&#34;&gt;打开视频&lt;/a&gt;`, "视频应内嵌链接兜底")
	assert.Contains(t, atom, `&lt;audio controls src=&#34;http://example.com/files/song.mp3&#34;`, "音频应渲染为 <audio controls>")
	assert.Contains(t, atom, `&lt;a href=&#34;http://example.com/files/song.mp3&#34;&gt;打开音频&lt;/a&gt;`, "音频应内嵌链接兜底")
	assert.Contains(t, atom, `&lt;a href=&#34;http://example.com/files/doc.pdf&#34;&gt;doc.pdf&lt;/a&gt;`, "普通文件应渲染为下载链接")
}

func TestGenerateRSS_MediaFieldEscaping(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

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
	atom, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.NotContains(t, atom, "<script>", "URL/文件名注入的原始 script 标签不得出现")
	assert.NotContains(t, atom, "&lt;script&gt;", "媒体字段必须先做 HTML 实体转义，杜绝单层转义形态")
}

func TestGenerateRSS_ReadThrough(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	echos := []echoModel.Echo{{ID: "e1", Username: "u", Content: "c", CreatedAt: time.Now().UTC().Unix()}}

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(echos, nil).Once()
	repo.EXPECT().TrackRSSCacheKey("rss:http:example.com").Return().Once()

	ctx := newRSSContext(t, "example.com")

	first, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)
	second, err := svc.GenerateRSS(ctx)
	require.NoError(t, err)

	assert.Equal(t, first, second, "缓存命中应返回与首回相同的内容")
}

func TestGenerateRSS_RepositoryError(t *testing.T) {
	repo := commonmock.NewMockCommonRepository(t)
	svc := commonService.NewCommonService(repo, newFakeCache())

	repo.EXPECT().GetAllEchos(mock.Anything, false).Return(nil, assert.AnError).Once()

	ctx := newRSSContext(t, "example.com")
	atom, err := svc.GenerateRSS(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Empty(t, atom)
}
