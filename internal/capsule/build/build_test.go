// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	oldEchoID = "11111111-1111-4111-8111-111111111111"
	newEchoID = "22222222-2222-4222-8222-222222222222"
	imageKey  = "cover.png"
)

var (
	oldCreated = time.Date(2026, 1, 5, 8, 30, 0, 0, time.UTC)
	newCreated = time.Date(2026, 2, 9, 12, 0, 0, 0, time.UTC)
)

var imageBytes = []byte("\x89PNG\r\n\x1a\nfake-image-bytes")

// writeCapsule 在临时目录里手搭一个最小但完整的胶囊：清单 + 两条 Echo
// （新的带媒体与标签）+ 一个媒体字节 + 一条评论 + 一条被剔除的 private Echo。
func writeCapsule(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	write := func(rel string, data []byte) {
		t.Helper()
		dest := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(dest), 0o755))
		require.NoError(t, os.WriteFile(dest, data, 0o644))
	}

	manifest, err := capsule.EncodeYAML(&capsule.Manifest{
		SchemaVersion: capsule.SchemaVersion,
		Generator:     "ech0-test",
		ExportedAt:    capsule.FormatUnix(newCreated.Unix()),
		Site: capsule.Site{
			SiteTitle:  "Test Site",
			ServerName: "tester",
			ServerLogo: "/api/files/images/logo.png",
		},
		Owner:    capsule.Owner{Username: "tester"},
		Connects: []capsule.Connect{{URL: "https://peer.example.com"}},
	})
	require.NoError(t, err)
	write(capsule.ManifestPath, manifest)

	echoes := []*capsule.EchoDoc{
		{
			ID:        oldEchoID,
			CreatedAt: capsule.FormatUnix(oldCreated.Unix()),
			Username:  "tester",
			Tags:      []string{"life"},
			FavCount:  3,
			Content:   "older echo\n",
		},
		{
			ID:        newEchoID,
			CreatedAt: capsule.FormatUnix(newCreated.Unix()),
			Username:  "tester",
			Tags:      []string{"life", "tech"},
			Layout:    "grid",
			Files: []capsule.FileRef{{
				Key:         imageKey,
				Category:    "image",
				Name:        "cover.png",
				ContentType: "image/png",
				Size:        int64(len(imageBytes)),
				Width:       64,
				Height:      48,
			}},
			Content: "newer echo\n",
		},
		{
			ID:        "33333333-3333-4333-8333-333333333333",
			CreatedAt: capsule.FormatUnix(newCreated.Unix() + 60),
			Username:  "tester",
			Private:   true,
			Content:   "secret\n",
		},
	}
	for _, doc := range echoes {
		createdAt, parseErr := capsule.ParseTime(doc.CreatedAt)
		require.NoError(t, parseErr)
		data, encErr := capsule.EncodeEcho(doc)
		require.NoError(t, encErr)
		write(capsule.EchoPath(doc.ID, time.Unix(createdAt, 0)), data)
	}

	write(capsule.MediaPath(imageKey), imageBytes)

	comments, err := capsule.EncodeYAML(&capsule.CommentsDoc{
		SchemaVersion: capsule.SchemaVersion,
		Comments: []capsule.Comment{
			{
				ID:        "c1",
				EchoID:    newEchoID,
				Nickname:  "guest",
				Content:   "nice",
				Status:    capsule.DefaultCommentStatus,
				CreatedAt: capsule.FormatUnix(newCreated.Unix() + 10),
			},
			{
				// 宿主是被剔除的 private Echo，必须不进 dataset。
				ID:        "c2",
				EchoID:    "33333333-3333-4333-8333-333333333333",
				Nickname:  "guest",
				Content:   "orphan",
				CreatedAt: capsule.FormatUnix(newCreated.Unix() + 20),
			},
		},
	})
	require.NoError(t, err)
	write(capsule.CommentsPath, comments)

	return root
}

// runBuild 打开临时胶囊并烘焙到一个全新目录。
func runBuild(t *testing.T, baseURL string) (string, *Result) {
	t.Helper()
	ctx := context.Background()

	src, err := capsule.Open(writeCapsule(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = src.Close() })

	loaded, err := capsule.Load(ctx, src)
	require.NoError(t, err)

	out := filepath.Join(t.TempDir(), "dist")
	res, err := Run(ctx, loaded, Options{Output: out, BaseURL: baseURL})
	require.NoError(t, err)
	return out, res
}

func readDataset(t *testing.T, dir string) *dataset {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "dataset.json"))
	require.NoError(t, err)
	ds := &dataset{}
	require.NoError(t, json.Unmarshal(raw, ds))
	return ds
}

func TestRunBakesDataset(t *testing.T) {
	dir, res := runBuild(t, "")

	assert.Equal(t, dir, res.Path)
	assert.Equal(t, 2, res.Echoes, "private echo must be excluded")
	assert.Equal(t, 1, res.Files)
	assert.Equal(t, 1, res.Comments, "comment on a dropped echo must be excluded")

	ds := readDataset(t, dir)
	assert.Equal(t, datasetSchemaVersion, ds.SchemaVersion)
	assert.Equal(t, "/", ds.BaseURL)

	// ① created_at 是 Unix 秒，且按降序排好。
	require.Len(t, ds.Echos, 2)
	assert.Equal(t, newEchoID, ds.Echos[0].ID)
	assert.Equal(t, newCreated.Unix(), ds.Echos[0].CreatedAt)
	assert.Equal(t, oldEchoID, ds.Echos[1].ID)
	assert.Equal(t, oldCreated.Unix(), ds.Echos[1].CreatedAt)
	assert.Greater(t, ds.Echos[0].CreatedAt, ds.Echos[1].CreatedAt)

	// 缺省 layout 回落到 waterfall，显式 layout 原样保留。
	assert.Equal(t, "grid", ds.Echos[0].Layout)
	assert.Equal(t, capsule.DefaultLayout, ds.Echos[1].Layout)
	assert.False(t, ds.Echos[0].Private)

	// ② 托管媒体的 URL 由 baseURL + 路由表现算。
	require.Len(t, ds.Echos[0].EchoFiles, 1)
	ef := ds.Echos[0].EchoFiles[0]
	assert.Equal(t, "/api/files/images/"+imageKey, ef.File.URL)
	assert.Equal(t, "local", ef.File.StorageType)
	assert.Equal(t, imageKey, ef.File.Key)
	assert.Equal(t, 0, ef.SortOrder)
	assert.Equal(t, newEchoID, ef.EchoID)
	assert.NotEmpty(t, ef.ID)
	assert.Equal(t, ef.File.ID, ef.FileID)

	// ③ 字节确实铺到了 api/files/ 下且与胶囊一致。
	got, err := os.ReadFile(filepath.Join(dir, "api", "files", "images", imageKey))
	require.NoError(t, err)
	assert.Equal(t, imageBytes, got)

	// ④ 入口 HTML 注入了静态开关，404.html 是它的副本。
	index, err := os.ReadFile(filepath.Join(dir, "index.html"))
	require.NoError(t, err)
	assert.Contains(t, string(index), "window.__ECH0_STATIC__=true")
	assert.Contains(t, string(index), `window.__ECH0_STATIC_BASE__="/"`)
	assert.Contains(t, string(index), `href="/rss.xml"`, "feed link must point at the baked file")
	notFound, err := os.ReadFile(filepath.Join(dir, "404.html"))
	require.NoError(t, err)
	assert.Equal(t, index, notFound)

	// ⑤ api/connect 与活实例响应体同形，统计值是冻结快照。
	envRaw, err := os.ReadFile(filepath.Join(dir, "api", "connect"))
	require.NoError(t, err)
	env := resultEnvelope{}
	require.NoError(t, json.Unmarshal(envRaw, &env))
	assert.Equal(t, 1, env.Code)
	assert.Equal(t, 2, env.Data.TotalEchos)
	assert.Equal(t, "tester", env.Data.SysUsername)
	assert.Equal(t, env.Data, ds.Connect)

	// api/connect 是给远端实例消费的名片：相对 logo 在对方页面上必然解析错，
	// 故必须按 server_url 绝对化，分支与活实例 ConnectService.GetConnect 同形。
	assert.Equal(t, "/api/files/images/logo.png", ds.Settings.ServerLogo,
		"站内渲染仍用相对路径")
	for _, tc := range []struct{ logo, serverURL, want string }{
		{"/api/files/images/logo.png", "https://a.example.com/", "https://a.example.com/api/files/images/logo.png"},
		{"https://cdn.example.com/l.png", "https://a.example.com", "https://cdn.example.com/l.png"},
		{"", "https://a.example.com", "https://a.example.com/Ech0.svg"},
		{"/Ech0.svg", "https://a.example.com", "https://a.example.com/Ech0.svg"},
		{"/api/files/images/logo.png", "", "/api/files/images/logo.png"},
	} {
		assert.Equal(t, tc.want, connectLogo(tc.logo, tc.serverURL, "/"),
			"logo=%q server_url=%q", tc.logo, tc.serverURL)
	}

	// settings / hello / 冻结开关。
	assert.False(t, ds.Settings.AllowRegister)
	assert.Equal(t, "Test Site", ds.Settings.SiteTitle)
	assert.Equal(t, "Test Site", ds.Hello.Hello)
	assert.NotEmpty(t, ds.Hello.Version)
	assert.False(t, ds.CommentForm.EnableComment)
	assert.False(t, ds.Agent.Enable)
	assert.True(t, ds.InitStatus.Initialized)
	require.Len(t, ds.Connects, 1)
	assert.Equal(t, "https://peer.example.com", ds.Connects[0].ConnectURL)

	// 标签：全站聚合，usage_count 按引用次数，Echo 内引用同一套 id。
	require.Len(t, ds.Tags, 2)
	assert.Equal(t, "life", ds.Tags[0].Name)
	assert.Equal(t, 2, ds.Tags[0].UsageCount)
	assert.Equal(t, "tech", ds.Tags[1].Name)
	assert.Equal(t, 1, ds.Tags[1].UsageCount)
	byName := map[string]string{}
	for _, tg := range ds.Tags {
		byName[tg.Name] = tg.ID
	}
	for _, tg := range ds.Echos[0].Tags {
		assert.Equal(t, byName[tg.Name], tg.ID, "echo tag id must match dataset.tags")
	}

	// 评论：只留宿主还在的那条，email 恒空，parent_id / user_id 为 null。
	require.Len(t, ds.Comments, 1)
	assert.Equal(t, "c1", ds.Comments[0].ID)
	assert.Empty(t, ds.Comments[0].Email)
	assert.Nil(t, ds.Comments[0].ParentID)
	assert.Nil(t, ds.Comments[0].UserID)
	assert.Contains(t, string(mustRead(t, filepath.Join(dir, "dataset.json"))), `"parent_id":null`)

	// 热力图：最近 30 天，末端是构建当天（UTC）。
	require.Len(t, ds.Heatmap, heatmapDays)
	assert.Equal(t, time.Now().UTC().Format(time.DateOnly), ds.Heatmap[heatmapDays-1].Date)

	// rss / sitemap 产出且包含每条 Echo。
	rss := string(mustRead(t, filepath.Join(dir, "rss.xml")))
	assert.Contains(t, rss, "/echo/"+newEchoID)
	sitemap := string(mustRead(t, filepath.Join(dir, "sitemap.xml")))
	assert.Contains(t, sitemap, "/echo/"+oldEchoID)
	assert.Contains(t, sitemap, oldCreated.Format(time.DateOnly))
}

// ⑥ 子路径部署时所有站内 URL 都要带上前缀。
func TestRunWithBaseURL(t *testing.T) {
	dir, _ := runBuild(t, "/blog/")

	ds := readDataset(t, dir)
	assert.Equal(t, "/blog/", ds.BaseURL)
	require.Len(t, ds.Echos, 2)
	require.Len(t, ds.Echos[0].EchoFiles, 1)
	assert.Equal(t, "/blog/api/files/images/"+imageKey, ds.Echos[0].EchoFiles[0].File.URL)
	assert.Equal(t, "/blog/api/files/images/logo.png", ds.Settings.ServerLogo)

	index := string(mustRead(t, filepath.Join(dir, "index.html")))
	assert.Contains(t, index, `window.__ECH0_STATIC_BASE__="/blog/"`)
	assert.Contains(t, index, `href="/blog/favicon.ico"`)
	assert.Contains(t, index, `href="/blog/rss.xml"`)
	assert.NotContains(t, index, `href="/favicon.ico"`)
	// 资源真的躺在拷贝出来的位置（前缀只改 HTML 引用，不改产物布局）。
	assert.FileExists(t, filepath.Join(dir, "favicon.ico"))
}

// 派生 id 必须跨构建稳定：前端按 tagIds 过滤，id 一漂移收藏的过滤链接就全断。
func TestDerivedIDsAreStable(t *testing.T) {
	first, _ := runBuild(t, "")
	second, _ := runBuild(t, "")

	a, b := readDataset(t, first), readDataset(t, second)
	require.Len(t, a.Tags, len(b.Tags))
	for i := range a.Tags {
		assert.Equal(t, a.Tags[i].ID, b.Tags[i].ID)
	}
	require.NotEmpty(t, a.Echos[0].EchoFiles)
	require.NotEmpty(t, b.Echos[0].EchoFiles)
	assert.Equal(t, a.Echos[0].EchoFiles[0].ID, b.Echos[0].EchoFiles[0].ID)
	assert.Equal(t, a.Echos[0].EchoFiles[0].FileID, b.Echos[0].EchoFiles[0].FileID)
	assert.Equal(t, a.Connects[0].ID, b.Connects[0].ID)
}

func TestRunRejectsNonEmptyOutput(t *testing.T) {
	ctx := context.Background()
	src, err := capsule.Open(writeCapsule(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = src.Close() })
	loaded, err := capsule.Load(ctx, src)
	require.NoError(t, err)

	out := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(out, "stale.txt"), []byte("x"), 0o644))

	_, err = Run(ctx, loaded, Options{Output: out})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not empty")
}

func TestNormalizeBaseURL(t *testing.T) {
	cases := map[string]string{
		"":       "/",
		"  ":     "/",
		"/":      "/",
		"blog":   "/blog/",
		"/blog":  "/blog/",
		"/blog/": "/blog/",
	}
	for in, want := range cases {
		assert.Equal(t, want, NormalizeBaseURL(in), "input %q", in)
	}
}

func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	data, err := os.ReadFile(p)
	require.NoError(t, err)
	return data
}
