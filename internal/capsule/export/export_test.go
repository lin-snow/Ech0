// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package export

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/lin-snow/ech0/internal/kvstore"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	publicEchoID  = "0193f1a1-1111-7000-8000-000000000001"
	privateEchoID = "0193f1a1-2222-7000-8000-000000000002"
	// 2023-11-14T22:13:20Z / 2023-11-15T22:13:20Z
	publicEchoAt  int64 = 1700000000
	privateEchoAt int64 = 1700086400

	publicBody = "hello **world**\n\n第二段：正文逐字保留。\n"
)

var testSite = settingModel.SystemSetting{
	SiteTitle:     "Ech0 测试站",
	ServerLogo:    "https://example.com/logo.png",
	ServerName:    "linsnow",
	ServerURL:     "https://example.com",
	AllowRegister: true, // 行为开关：必须留在库里，不得进胶囊
	DefaultLocale: "zh-CN",
	ICPNumber:     "京ICP备00000000号",
	FooterContent: "footer",
	FooterLink:    "https://example.com/about",
	MetingAPI:     "https://meting.example.com/api",
	CustomCSS:     "body{}",
	CustomJS:      "console.log(1)",
}

// newFixture 造一个最小但覆盖全部分流的实例：公开/私密 Echo、托管/外链/悬空/仅私密
// 引用四种 file 行、已审核/待审核/孤儿三种评论。
func newFixture(t *testing.T) (Deps, string) {
	t.Helper()

	db := helpers.NewTestDB(t)
	dataRoot := t.TempDir()
	selector := storage.NewStorageManagerForTest(dataRoot).GetSelector()
	kv := kvstore.NewMemory()
	require.NoError(t, coreSetting.Set(context.Background(), kv, coreSetting.System, testSite))

	require.NoError(t, db.Create(&userModel.User{
		ID: "u-owner", Username: "linsnow", IsAdmin: true, IsOwner: true,
	}).Error)
	require.NoError(t, db.Create(&userModel.User{ID: "u-other", Username: "guest"}).Error)

	require.NoError(t, db.Create(&connectModel.Connected{
		ID: "c-1", ConnectURL: "https://peer.example.com",
	}).Error)

	files := []fileModel.File{
		{
			ID: "f-pic", Key: "pic.png", StorageType: "local", Name: "pic.png",
			ContentType: "image/png", Category: "image", Size: 7, Width: 4, Height: 2, UserID: "u-owner",
		},
		{
			ID: "f-clip", Key: "clip.mp4", StorageType: "local", Name: "clip.mp4",
			ContentType: "video/mp4", Category: "video", Size: 5, UserID: "u-owner",
		},
		{
			ID: "f-orphan", Key: "orphan.pdf", StorageType: "local", Name: "orphan.pdf",
			ContentType: "application/pdf", Category: "pdf", Size: 6, UserID: "u-owner",
		},
		{
			ID: "f-secret", Key: "secret.png", StorageType: "local", Name: "secret.png",
			ContentType: "image/png", Category: "image", Size: 6, UserID: "u-owner",
		},
		{
			ID: "f-ext", StorageType: "external", URL: "https://cdn.example.com/remote.png",
			Name: "remote.png", ContentType: "image/png", Category: "image", UserID: "u-owner",
		},
	}
	for i := range files {
		require.NoError(t, db.Create(&files[i]).Error)
	}
	for key, body := range map[string]string{
		"pic.png":    "PNGDATA",
		"clip.mp4":   "MP4!!",
		"orphan.pdf": "%PDF-1",
		"secret.png": "SECRET",
	} {
		require.NoError(t, selector.Put(
			context.Background(), storage.StorageTypeLocal, key, strings.NewReader(body),
		))
	}

	require.NoError(t, db.Create(&echoModel.Echo{
		ID: publicEchoID, Content: publicBody, Username: "linsnow", Layout: "grid",
		Private: false, UserID: "u-owner", FavCount: 3, CreatedAt: publicEchoAt,
		Tags: []echoModel.Tag{{ID: "t-1", Name: "生活"}, {ID: "t-2", Name: "tech"}},
	}).Error)
	require.NoError(t, db.Create(&echoModel.EchoExtension{
		ID: "x-1", EchoID: publicEchoID, Type: echoModel.Extension_WEBSITE,
		Payload: map[string]any{"url": "https://example.org"},
	}).Error)
	require.NoError(t, db.Create(&echoModel.Echo{
		ID: privateEchoID, Content: "私密内容", Username: "linsnow", Layout: "waterfall",
		Private: true, UserID: "u-owner", CreatedAt: privateEchoAt,
	}).Error)

	// sort_order 与插入顺序故意相反：胶囊的 files 数组顺序必须由 sort_order 决定。
	links := []fileModel.EchoFile{
		{ID: "ef-1", EchoID: publicEchoID, FileID: "f-pic", SortOrder: 1},
		{ID: "ef-2", EchoID: publicEchoID, FileID: "f-clip", SortOrder: 0},
		{ID: "ef-3", EchoID: publicEchoID, FileID: "f-ext", SortOrder: 2},
		{ID: "ef-4", EchoID: privateEchoID, FileID: "f-secret", SortOrder: 0},
		{ID: "ef-5", EchoID: privateEchoID, FileID: "f-pic", SortOrder: 1},
	}
	for i := range links {
		require.NoError(t, db.Create(&links[i]).Error)
	}

	comments := []commentModel.Comment{
		{
			ID: "cm-1", EchoID: publicEchoID, Nickname: "alice", Email: "alice@example.com",
			Content: "nice", Status: commentModel.StatusApproved, Source: commentModel.SourceGuest,
			IPHash: "deadbeef", UserAgent: "curl", CreatedAt: publicEchoAt + 10,
		},
		{
			ID: "cm-2", EchoID: publicEchoID, Nickname: "bob", Email: "bob@example.com",
			Content: "pending", Status: commentModel.StatusPending, Source: commentModel.SourceGuest,
			CreatedAt: publicEchoAt + 20,
		},
		{
			ID: "cm-3", EchoID: privateEchoID, Nickname: "eve", Email: "eve@example.com",
			Content: "on private", Status: commentModel.StatusApproved, Source: commentModel.SourceGuest,
			CreatedAt: publicEchoAt + 30,
		},
	}
	for i := range comments {
		require.NoError(t, db.Create(&comments[i]).Error)
	}

	return Deps{DB: db, Selector: selector, KV: kv}, dataRoot
}

func runExport(t *testing.T, deps Deps, opts Options) (*Result, string) {
	t.Helper()
	if opts.Output == "" {
		opts.Output = filepath.Join(t.TempDir(), "capsule")
	}
	res, err := Run(context.Background(), deps, opts)
	require.NoError(t, err)
	return res, opts.Output
}

func readCapsuleFile(t *testing.T, dir, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	require.NoError(t, err)
	return b
}

func TestRun_ManifestRoundTrips(t *testing.T) {
	deps, _ := newFixture(t)
	res, out := runExport(t, deps, Options{Generator: "ech0 test"})

	raw := readCapsuleFile(t, out, capsule.ManifestPath)
	var manifest capsule.Manifest
	unknown, err := capsule.DecodeYAML(raw, &manifest)
	require.NoError(t, err)
	assert.Empty(t, unknown)

	assert.Equal(t, capsule.SchemaVersion, manifest.SchemaVersion)
	assert.Equal(t, "ech0 test", manifest.Generator)
	_, err = capsule.ParseTime(manifest.ExportedAt)
	assert.NoError(t, err)
	assert.Equal(t, "linsnow", manifest.Owner.Username)
	assert.Equal(t, []capsule.Connect{{URL: "https://peer.example.com"}}, manifest.Connects)
	assert.Equal(t, 1, res.Connects)

	// site 逐字对齐库中设置（读回归一化后的值，避免与 locale 归一化耦合）。
	stored, err := coreSetting.Get(context.Background(), deps.KV, coreSetting.System)
	require.NoError(t, err)
	assert.Equal(t, capsule.Site{
		SiteTitle:     stored.SiteTitle,
		ServerLogo:    stored.ServerLogo,
		ServerName:    stored.ServerName,
		ServerURL:     stored.ServerURL,
		DefaultLocale: stored.DefaultLocale,
		ICPNumber:     stored.ICPNumber,
		FooterContent: stored.FooterContent,
		FooterLink:    stored.FooterLink,
		MetingAPI:     stored.MetingAPI,
		CustomCSS:     stored.CustomCSS,
		CustomJS:      stored.CustomJS,
	}, manifest.Site)
	// 行为开关不得入胶囊（spec §3）。
	assert.NotContains(t, string(raw), "allow_register")
}

func TestRun_EchoDocument(t *testing.T) {
	deps, _ := newFixture(t)
	_, out := runExport(t, deps, Options{})

	rel := capsule.EchoPath(publicEchoID, time.Unix(publicEchoAt, 0))
	assert.Equal(t, "echoes/2023/2023-11-14-00000001.md", rel)

	doc, unknown, err := capsule.DecodeEcho(readCapsuleFile(t, out, rel))
	require.NoError(t, err)
	assert.Empty(t, unknown)

	assert.Equal(t, publicEchoID, doc.ID)
	assert.Equal(t, capsule.FormatUnix(publicEchoAt), doc.CreatedAt)
	assert.Equal(t, "linsnow", doc.Username)
	assert.Equal(t, []string{"生活", "tech"}, doc.Tags)
	assert.Equal(t, "grid", doc.Layout)
	assert.False(t, doc.Private)
	assert.Equal(t, 3, doc.FavCount)
	assert.Equal(t, publicBody, doc.Content)
	require.NotNil(t, doc.Extension)
	assert.Equal(t, echoModel.Extension_WEBSITE, doc.Extension.Type)
	assert.Equal(t, map[string]any{"url": "https://example.org"}, doc.Extension.Payload)

	// 数组顺序 = sort_order 升序；托管条目只带 key，外链条目只带 url。
	require.Len(t, doc.Files, 3)
	assert.Equal(t, []string{"clip.mp4", "pic.png", ""}, []string{doc.Files[0].Key, doc.Files[1].Key, doc.Files[2].Key})
	assert.Empty(t, doc.Files[0].URL)
	assert.Equal(t, "https://cdn.example.com/remote.png", doc.Files[2].URL)
	assert.Equal(t, "f-pic", doc.Files[1].ID)
	assert.Equal(t, "image", doc.Files[1].Category)
	assert.Equal(t, int64(7), doc.Files[1].Size)
	assert.Equal(t, 4, doc.Files[1].Width)
}

func TestRun_MediaBytesAreSelfContained(t *testing.T) {
	deps, _ := newFixture(t)
	res, out := runExport(t, deps, Options{})

	assert.Equal(t, "PNGDATA", string(readCapsuleFile(t, out, "files/images/pic.png")))
	assert.Equal(t, "MP4!!", string(readCapsuleFile(t, out, "files/videos/clip.mp4")))
	// 悬空文件合法且照常导出（check 侧只告警）。
	assert.Equal(t, "%PDF-1", string(readCapsuleFile(t, out, "files/documents/orphan.pdf")))
	// 外链没有字节。
	assert.NoFileExists(t, filepath.Join(out, "files", "images", "remote.png"))

	assert.Equal(t, 4, res.Files) // pic / clip / orphan / external，secret 被排除
	assert.Equal(t, 1, res.ExternalFiles)
}

func TestRun_ExcludesPrivateByDefault(t *testing.T) {
	deps, _ := newFixture(t)
	res, out := runExport(t, deps, Options{})

	assert.NoFileExists(t, filepath.Join(out, filepath.FromSlash(
		capsule.EchoPath(privateEchoID, time.Unix(privateEchoAt, 0)))))
	// 仅被私密 Echo 引用的图片不得出门。
	assert.NoFileExists(t, filepath.Join(out, "files", "images", "secret.png"))
	assert.Equal(t, 1, res.Echoes)
	assert.Equal(t, 1, res.SkippedPrivate)

	// 私密 Echo 名下的评论会随之变成孤儿，不导出；待审核评论同样不导出。
	var doc capsule.CommentsDoc
	_, err := capsule.DecodeYAML(readCapsuleFile(t, out, capsule.CommentsPath), &doc)
	require.NoError(t, err)
	require.Len(t, doc.Comments, 1)
	assert.Equal(t, "cm-1", doc.Comments[0].ID)
	assert.Equal(t, 1, res.Comments)
	raw := string(readCapsuleFile(t, out, capsule.CommentsPath))
	for _, forbidden := range capsule.ForbiddenCommentFields {
		assert.NotContains(t, raw, forbidden)
	}
}

func TestRun_IncludePrivate(t *testing.T) {
	deps, _ := newFixture(t)
	res, out := runExport(t, deps, Options{IncludePrivate: true})

	doc, _, err := capsule.DecodeEcho(readCapsuleFile(t, out,
		capsule.EchoPath(privateEchoID, time.Unix(privateEchoAt, 0))))
	require.NoError(t, err)
	assert.True(t, doc.Private)
	assert.Equal(t, "私密内容", doc.Content)

	assert.Equal(t, "SECRET", string(readCapsuleFile(t, out, "files/images/secret.png")))
	assert.Equal(t, 2, res.Echoes)
	assert.Equal(t, 0, res.SkippedPrivate)
	assert.Equal(t, 5, res.Files)
	assert.Equal(t, 2, res.Comments)
}

func TestRun_Zip(t *testing.T) {
	deps, _ := newFixture(t)
	res, err := Run(context.Background(), deps, Options{
		Output: filepath.Join(t.TempDir(), "capsule"),
		Zip:    true,
	})
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(res.Path, ".zip"), res.Path)
	assert.FileExists(t, res.Path)

	src, err := capsule.Open(res.Path)
	require.NoError(t, err)
	defer func() { _ = src.Close() }()

	loaded, err := capsule.Load(context.Background(), src)
	require.NoError(t, err)
	require.NoError(t, loaded.ManifestErr)
	assert.Equal(t, "linsnow", loaded.Manifest.Owner.Username)
	require.Len(t, loaded.Echoes, 1)
	// zip 内布局与目录形态一致：无额外顶层目录。
	assert.Equal(t, int64(len("PNGDATA")), loaded.MediaPaths["files/images/pic.png"])
}

func TestRun_RefusesNonEmptyOutputDir(t *testing.T) {
	deps, _ := newFixture(t)
	out := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(out, "keep.txt"), []byte("x"), 0o600))

	_, err := Run(context.Background(), deps, Options{Output: out})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not empty")
	assert.FileExists(t, filepath.Join(out, "keep.txt"))
}

func TestRun_ReportsUnreadableManagedFiles(t *testing.T) {
	deps, dataRoot := newFixture(t)
	require.NoError(t, os.Remove(filepath.Join(dataRoot, "images", "pic.png")))
	require.NoError(t, os.Remove(filepath.Join(dataRoot, "videos", "clip.mp4")))

	_, err := Run(context.Background(), deps, Options{Output: filepath.Join(t.TempDir(), "capsule")})
	require.Error(t, err)
	// 一次跑完再报全清单，而不是撞上第一条就退出。
	assert.Contains(t, err.Error(), "pic.png")
	assert.Contains(t, err.Error(), "clip.mp4")
}

func TestUniquePath_DeduplicatesCollisions(t *testing.T) {
	used := make(map[string]struct{})
	base := capsule.EchoPath("0193f1a1-aaaa", time.Unix(publicEchoAt, 0))

	// 命名取 id 去掉横线后的末 8 位（UUIDv7 的随机段）。
	assert.Equal(t, "echoes/2023/2023-11-14-f1a1aaaa.md", uniquePath(used, base))
	assert.Equal(t, "echoes/2023/2023-11-14-f1a1aaaa-2.md", uniquePath(used, base))
	assert.Equal(t, "echoes/2023/2023-11-14-f1a1aaaa-3.md", uniquePath(used, base))
}
