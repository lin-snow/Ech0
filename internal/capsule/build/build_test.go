// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
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

func runBuild(t *testing.T, baseURL string) (string, *Result) {
	t.Helper()
	ctx := context.Background()

	src, err := capsule.Open(writeCapsule(t))
	require.NoError(t, err)
	t.Cleanup(func() { _ = src.Close() })

	loaded, err := capsule.Load(ctx, src)
	require.NoError(t, err)

	out := filepath.Join(t.TempDir(), "dist")
	res, err := run(ctx, loaded, Options{Output: out, BaseURL: baseURL}, fixtureSPA())
	require.NoError(t, err)
	return out, res
}

func fixtureSPA() fs.FS {
	const index = `<!doctype html>
<html>
<head>
<link rel="icon" href="/favicon.ico" />
<link rel="manifest" href="/app.webmanifest" />
<link rel="alternate" type="application/atom+xml" href="/rss" />
<link rel="stylesheet" href="/assets/index-abc.css" />
<script type="module" src="/assets/index-abc.js"></script>
</head>
<body><div id="app"></div></body>
</html>`
	return fstest.MapFS{
		spaRoot + "/" + indexFile:          {Data: []byte(index)},
		spaRoot + "/assets/index-abc.js":   {Data: []byte("console.log(0)")},
		spaRoot + "/assets/index-abc.css":  {Data: []byte(".a{}")},
		spaRoot + "/favicon.ico":           {Data: []byte("ico")},
		spaRoot + "/_plugin-vue_helper.js": {Data: []byte("export {}")},
	}
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

	require.Len(t, ds.Echos, 2)
	assert.Equal(t, newEchoID, ds.Echos[0].ID)
	assert.Equal(t, newCreated.Unix(), ds.Echos[0].CreatedAt)
	assert.Equal(t, oldEchoID, ds.Echos[1].ID)
	assert.Equal(t, oldCreated.Unix(), ds.Echos[1].CreatedAt)
	assert.Greater(t, ds.Echos[0].CreatedAt, ds.Echos[1].CreatedAt)

	assert.Equal(t, "grid", ds.Echos[0].Layout)
	assert.Equal(t, capsule.DefaultLayout, ds.Echos[1].Layout)
	assert.False(t, ds.Echos[0].Private)

	require.Len(t, ds.Echos[0].EchoFiles, 1)
	ef := ds.Echos[0].EchoFiles[0]
	assert.Equal(t, "/api/files/images/"+imageKey, ef.File.URL)
	assert.Equal(t, "local", ef.File.StorageType)
	assert.Equal(t, imageKey, ef.File.Key)
	assert.Equal(t, 0, ef.SortOrder)
	assert.Equal(t, newEchoID, ef.EchoID)
	assert.NotEmpty(t, ef.ID)
	assert.Equal(t, ef.File.ID, ef.FileID)

	got, err := os.ReadFile(filepath.Join(dir, "api", "files", "images", imageKey))
	require.NoError(t, err)
	assert.Equal(t, imageBytes, got)

	index, err := os.ReadFile(filepath.Join(dir, "index.html"))
	require.NoError(t, err)
	assert.Contains(t, string(index), "window.__ECH0_STATIC__=true")
	assert.Contains(t, string(index), `window.__ECH0_STATIC_BASE__="/"`)
	assert.Contains(t, string(index), `href="/rss.xml"`, "feed link must point at the baked file")
	notFound, err := os.ReadFile(filepath.Join(dir, "404.html"))
	require.NoError(t, err)
	assert.Equal(t, index, notFound)

	envRaw, err := os.ReadFile(filepath.Join(dir, "api", "connect"))
	require.NoError(t, err)
	env := resultEnvelope{}
	require.NoError(t, json.Unmarshal(envRaw, &env))
	assert.Equal(t, 1, env.Code)
	assert.Equal(t, 2, env.Data.TotalEchos)
	assert.Equal(t, "tester", env.Data.SysUsername)
	assert.Equal(t, env.Data, ds.Connect)

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

	assert.False(t, ds.Settings.AllowRegister)
	assert.Equal(t, "Test Site", ds.Settings.SiteTitle)
	assert.Equal(t, "Test Site", ds.Hello.Hello)
	assert.NotEmpty(t, ds.Hello.Version)
	assert.False(t, ds.CommentForm.EnableComment)
	assert.False(t, ds.Agent.Enable)
	assert.True(t, ds.InitStatus.Initialized)
	require.Len(t, ds.Connects, 1)
	assert.Equal(t, "https://peer.example.com", ds.Connects[0].ConnectURL)

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

	require.Len(t, ds.Comments, 1)
	assert.Equal(t, "c1", ds.Comments[0].ID)
	assert.Empty(t, ds.Comments[0].Email)
	assert.Nil(t, ds.Comments[0].ParentID)
	assert.Nil(t, ds.Comments[0].UserID)
	assert.Contains(t, string(mustRead(t, filepath.Join(dir, "dataset.json"))), `"parent_id":null`)

	require.Len(t, ds.Heatmap, heatmapDays)
	assert.Equal(t, time.Now().UTC().Format(time.DateOnly), ds.Heatmap[heatmapDays-1].Date)

	rss := string(mustRead(t, filepath.Join(dir, "rss.xml")))
	assert.Contains(t, rss, "/echo/"+newEchoID)
	sitemap := string(mustRead(t, filepath.Join(dir, "sitemap.xml")))
	assert.Contains(t, sitemap, "/echo/"+oldEchoID)
	assert.Contains(t, sitemap, oldCreated.Format(time.DateOnly))
}

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
	assert.FileExists(t, filepath.Join(dir, "favicon.ico"))
}

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

	_, err = run(ctx, loaded, Options{Output: out}, fixtureSPA())
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
