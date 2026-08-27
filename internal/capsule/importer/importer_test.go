// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package importer

import (
	"context"
	"io"
	"os"
	"path/filepath"
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
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	echoIDPublic  = "01890000-0000-7000-8000-000000000001"
	echoIDPrivate = "01890000-0000-7000-8000-000000000002"
	echoIDGhost   = "01890000-0000-7000-8000-000000000003"

	commentIDLive   = "01890000-0000-7000-8000-0000000000a1"
	commentIDOrphan = "01890000-0000-7000-8000-0000000000a2"

	publicCreatedAt = "2024-03-05T06:07:08Z"
	publicCreatedTS = int64(1709618828)
)

var pngBytes = []byte("\x89PNG\r\n\x1a\nfake-image-payload")

type fixture struct {
	db      *gorm.DB
	deps    Deps
	ownerID string
	aliceID string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	db := helpers.NewTestDB(t)
	dataRoot := t.TempDir()
	selector := storage.NewStorageManagerForTest(dataRoot).GetSelector()

	owner := userModel.User{Username: "owner", IsOwner: true, IsAdmin: true}
	require.NoError(t, db.Create(&owner).Error)
	alice := userModel.User{Username: "alice"}
	require.NoError(t, db.Create(&alice).Error)

	kv := kvstore.NewMemory()
	require.NoError(t, coreSetting.Set(context.Background(), kv, coreSetting.System,
		settingModel.SystemSetting{SiteTitle: "Existing Title"}))

	return &fixture{
		db:      db,
		ownerID: owner.ID,
		aliceID: alice.ID,
		deps: Deps{
			DB:       db,
			Tx:       transaction.NewGormTransactor(func() *gorm.DB { return db }),
			Selector: selector,
			KV:       kv,
		},
	}
}

func writeCapsule(
	t *testing.T,
	manifest *capsule.Manifest,
	docs []*capsule.EchoDoc,
	comments *capsule.CommentsDoc,
	media map[string][]byte,
) string {
	t.Helper()
	dir := t.TempDir()

	writeFile := func(rel string, data []byte) {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, data, 0o644))
	}

	raw, err := capsule.EncodeYAML(manifest)
	require.NoError(t, err)
	writeFile(capsule.ManifestPath, raw)

	require.NoError(t, os.MkdirAll(filepath.Join(dir, capsule.EchoesDir), 0o755))
	for _, doc := range docs {
		body, encErr := capsule.EncodeEcho(doc)
		require.NoError(t, encErr)
		ts, tErr := capsule.ParseTime(doc.CreatedAt)
		require.NoError(t, tErr)
		writeFile(capsule.EchoPath(doc.ID, time.Unix(ts, 0).UTC()), body)
	}

	if comments != nil {
		raw, err = capsule.EncodeYAML(comments)
		require.NoError(t, err)
		writeFile(capsule.CommentsPath, raw)
	}
	for key, data := range media {
		writeFile(capsule.MediaPath(key), data)
	}
	return dir
}

func loadCapsule(t *testing.T, dir string) *capsule.Loaded {
	t.Helper()
	src, err := capsule.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = src.Close() })

	loaded, err := capsule.Load(context.Background(), src)
	require.NoError(t, err)
	return loaded
}

func fullManifest() *capsule.Manifest {
	return &capsule.Manifest{
		SchemaVersion: capsule.SchemaVersion,
		Generator:     "ech0-test",
		Site: capsule.Site{
			SiteTitle:  "From Capsule",
			ServerName: "capsule-instance",
			FooterLink: "https://example.com/footer",
		},
		Owner:    capsule.Owner{Username: "owner"},
		Connects: []capsule.Connect{{URL: "https://peer.example.com"}},
	}
}

func fullDocs() []*capsule.EchoDoc {
	return []*capsule.EchoDoc{
		{
			ID:        echoIDPublic,
			CreatedAt: publicCreatedAt,
			Username:  "alice",
			Tags:      []string{"golang", "life"},
			Layout:    echoModel.LayoutGrid,
			FavCount:  42,
			Files: []capsule.FileRef{{
				Key:         "pic.png",
				Category:    string(storage.CategoryImage),
				Name:        "原图.png",
				ContentType: "image/png",
				Size:        int64(len(pngBytes)),
				Width:       3,
				Height:      4,
			}},
			Extension: &capsule.Extension{
				Type:    echoModel.Extension_WEBSITE,
				Payload: map[string]any{"url": "https://example.com"},
			},
			Content: "hello **capsule**\n",
		},
		{
			ID:        echoIDPrivate,
			CreatedAt: "2024-03-06T00:00:00Z",
			Username:  "alice",
			Private:   true,
			Content:   "secret\n",
		},
		{
			ID:        echoIDGhost,
			CreatedAt: "2024-03-07T00:00:00Z",
			Username:  "ghost",
			Content:   "unknown author\n",
		},
	}
}

func fullComments() *capsule.CommentsDoc {
	return &capsule.CommentsDoc{
		SchemaVersion: capsule.SchemaVersion,
		Comments: []capsule.Comment{
			{
				ID:        commentIDLive,
				EchoID:    echoIDPublic,
				Nickname:  "visitor",
				Content:   "nice",
				CreatedAt: "2024-03-05T07:00:00Z",
			},
			{
				ID:        commentIDOrphan,
				EchoID:    echoIDPrivate,
				Nickname:  "lurker",
				Content:   "hidden",
				CreatedAt: "2024-03-06T07:00:00Z",
			},
		},
	}
}

func TestRun_LandsCapsuleVerbatim(t *testing.T) {
	f := newFixture(t)
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{"pic.png": pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)

	require.Equal(t, 2, res.EchoesCreated)
	require.Equal(t, 1, res.SkippedPrivate)
	require.Equal(t, 0, res.EchoesSkipped)
	require.Equal(t, 2, res.TagsCreated)
	require.Equal(t, 1, res.FilesCreated)
	require.Equal(t, 1, res.CommentsCreated)
	require.Equal(t, 1, res.OrphanComments)
	require.Empty(t, res.Renames)

	var echo echoModel.Echo
	require.NoError(t, f.db.Where("id = ?", echoIDPublic).First(&echo).Error)
	require.Equal(t, publicCreatedTS, echo.CreatedAt)
	require.Equal(t, "alice", echo.Username)
	require.Equal(t, 42, echo.FavCount)
	require.Equal(t, "hello **capsule**\n", echo.Content)
	require.Equal(t, echoModel.LayoutGrid, echo.Layout)
	require.False(t, echo.Private)
	require.Equal(t, f.aliceID, echo.UserID)

	var ghost echoModel.Echo
	require.NoError(t, f.db.Where("id = ?", echoIDGhost).First(&ghost).Error)
	require.Equal(t, "ghost", ghost.Username, "username 逐字保留，不因查无此人而改写")
	require.Equal(t, f.ownerID, ghost.UserID, "查无此人则挂 owner")

	var private int64
	require.NoError(t, f.db.Model(&echoModel.Echo{}).Where("id = ?", echoIDPrivate).Count(&private).Error)
	require.Zero(t, private)

	var tags []echoModel.Tag
	require.NoError(t, f.db.Order("name asc").Find(&tags).Error)
	require.Len(t, tags, 2)
	require.Equal(t, "golang", tags[0].Name)
	require.Equal(t, 1, tags[0].UsageCount)
	require.Equal(t, 1, tags[1].UsageCount)

	var links int64
	require.NoError(t, f.db.Model(&echoModel.EchoTag{}).Where("echo_id = ?", echoIDPublic).Count(&links).Error)
	require.EqualValues(t, 2, links)

	var file fileModel.File
	require.NoError(t, f.db.Where("key = ?", "pic.png").First(&file).Error)
	require.Equal(t, string(storage.StorageTypeLocal), file.StorageType)
	require.Equal(t, "原图.png", file.Name)
	require.Equal(t, "image/png", file.ContentType)
	require.EqualValues(t, len(pngBytes), file.Size)
	require.Equal(t, 3, file.Width)
	require.Equal(t, 4, file.Height)
	require.Equal(t, string(storage.CategoryImage), file.Category)
	require.Equal(t, f.aliceID, file.UserID)
	require.Empty(t, file.URL)

	var link fileModel.EchoFile
	require.NoError(t, f.db.Where("echo_id = ?", echoIDPublic).First(&link).Error)
	require.Equal(t, file.ID, link.FileID)
	require.Equal(t, 0, link.SortOrder)

	require.Equal(t, pngBytes, readStoredBytes(t, f, "pic.png"))

	var ext echoModel.EchoExtension
	require.NoError(t, f.db.Where("echo_id = ?", echoIDPublic).First(&ext).Error)
	require.Equal(t, echoModel.Extension_WEBSITE, ext.Type)
	require.Equal(t, "https://example.com", ext.Payload["url"])

	var comment commentModel.Comment
	require.NoError(t, f.db.Where("id = ?", commentIDLive).First(&comment).Error)
	require.Equal(t, echoIDPublic, comment.EchoID)
	require.Equal(t, "visitor", comment.Nickname)
	require.Equal(t, commentModel.Status(capsule.DefaultCommentStatus), comment.Status)
	require.Equal(t, commentModel.SourceGuest, comment.Source)
	require.EqualValues(t, 1709622000, comment.CreatedAt)
	require.Empty(t, comment.Email)
	require.Empty(t, comment.IPHash)
	require.Empty(t, comment.UserAgent)
	require.Nil(t, comment.UserID)

	var orphan int64
	require.NoError(t, f.db.Model(&commentModel.Comment{}).Where("id = ?", commentIDOrphan).Count(&orphan).Error)
	require.Zero(t, orphan)

	require.ElementsMatch(t, []string{"server_name", "footer_link"}, res.SiteFieldsFilled)
	sys, err := coreSetting.Get(context.Background(), f.deps.KV, coreSetting.System)
	require.NoError(t, err)
	require.Equal(t, "Existing Title", sys.SiteTitle, "已配置项不被胶囊覆盖")
	require.Equal(t, "capsule-instance", sys.ServerName)
	require.Equal(t, "https://example.com/footer", sys.FooterLink)

	var connects []connectModel.Connected
	require.NoError(t, f.db.Find(&connects).Error)
	require.Len(t, connects, 1)
	require.Equal(t, "https://peer.example.com", connects[0].ConnectURL)
}

func TestRun_IsIdempotent(t *testing.T) {
	f := newFixture(t)
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{"pic.png": pngBytes})

	_, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)
	before := snapshotCounts(t, f.db)

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)

	require.Equal(t, 0, res.EchoesCreated)
	require.Equal(t, 2, res.EchoesSkipped)
	require.Equal(t, 1, res.SkippedPrivate)
	require.Equal(t, 0, res.TagsCreated)
	require.Equal(t, 0, res.FilesCreated, "幂等跳过的 Echo 其关联文件一并跳过")
	require.Equal(t, 0, res.FilesRenamed)
	require.Equal(t, 0, res.CommentsCreated)
	require.Equal(t, 1, res.CommentsSkipped)
	require.Equal(t, 1, res.OrphanComments)
	require.Empty(t, res.SiteFieldsFilled, "站点设置已被首轮填满，第二轮无空位可填")

	require.Equal(t, before, snapshotCounts(t, f.db))
}

func TestRun_DryRunWritesNothing(t *testing.T) {
	f := newFixture(t)
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{"pic.png": pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{DryRun: true})
	require.NoError(t, err)

	require.Equal(t, 2, res.EchoesCreated)
	require.Equal(t, 1, res.FilesCreated)
	require.Equal(t, 2, res.TagsCreated)
	require.Equal(t, 1, res.CommentsCreated)
	require.Equal(t, 1, res.OrphanComments)
	require.NotEmpty(t, res.SiteFieldsFilled)

	require.Equal(t, map[string]int64{}, nonZero(snapshotCounts(t, f.db)))

	_, err = f.deps.Selector.Get(context.Background(), storage.StorageTypeLocal, "pic.png")
	require.Error(t, err)

	sys, err := coreSetting.Get(context.Background(), f.deps.KV, coreSetting.System)
	require.NoError(t, err)
	require.Equal(t, "Existing Title", sys.SiteTitle)
	require.Empty(t, sys.ServerName, "dry-run 不得写 KV：仓储写完会刷进程内缓存，回滚兜不住")
}

func TestRun_RenamesOnKeyCollisionWithDifferentBytes(t *testing.T) {
	f := newFixture(t)
	first := writeCapsule(t, fullManifest(), []*capsule.EchoDoc{{
		ID:        echoIDPublic,
		CreatedAt: publicCreatedAt,
		Username:  "alice",
		Files:     []capsule.FileRef{{Key: "dup.png", Category: string(storage.CategoryImage)}},
		Content:   "first\n",
	}}, nil, map[string][]byte{"dup.png": pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, first), Options{})
	require.NoError(t, err)
	require.Equal(t, 1, res.FilesCreated)
	require.Empty(t, res.Renames)

	other := []byte("\x89PNG\r\n\x1a\na-totally-different-payload")
	second := writeCapsule(t, fullManifest(), []*capsule.EchoDoc{{
		ID:        echoIDGhost,
		CreatedAt: "2024-03-07T00:00:00Z",
		Username:  "alice",
		Files:     []capsule.FileRef{{Key: "dup.png", Category: string(storage.CategoryImage)}},
		Content:   "second\n",
	}}, nil, map[string][]byte{"dup.png": other})

	res, err = Run(context.Background(), f.deps, loadCapsule(t, second), Options{})
	require.NoError(t, err)
	require.Equal(t, 1, res.FilesRenamed)
	require.Equal(t, 0, res.FilesCreated)
	require.Len(t, res.Renames, 1)
	require.Contains(t, res.Renames[0], "dup.png -> ")

	var files []fileModel.File
	require.NoError(t, f.db.Order("key asc").Find(&files).Error)
	require.Len(t, files, 2)
	require.Equal(t, pngBytes, readStoredBytes(t, f, "dup.png"), "既有 key 的字节没被覆盖")

	newKey := files[0].Key
	if newKey == "dup.png" {
		newKey = files[1].Key
	}
	require.Equal(t, other, readStoredBytes(t, f, newKey))
}

func TestRun_ReusesFileWithIdenticalBytes(t *testing.T) {
	f := newFixture(t)
	docs := []*capsule.EchoDoc{{
		ID:        echoIDPublic,
		CreatedAt: publicCreatedAt,
		Files:     []capsule.FileRef{{Key: "dup.png", Category: string(storage.CategoryImage)}},
		Content:   "first\n",
	}}
	first := writeCapsule(t, fullManifest(), docs, nil, map[string][]byte{"dup.png": pngBytes})
	_, err := Run(context.Background(), f.deps, loadCapsule(t, first), Options{})
	require.NoError(t, err)

	docs[0].ID = echoIDGhost
	second := writeCapsule(t, fullManifest(), docs, nil, map[string][]byte{"dup.png": pngBytes})
	res, err := Run(context.Background(), f.deps, loadCapsule(t, second), Options{})
	require.NoError(t, err)

	require.Equal(t, 1, res.FilesReused)
	require.Equal(t, 0, res.FilesCreated)
	require.Empty(t, res.Renames)

	var files int64
	require.NoError(t, f.db.Model(&fileModel.File{}).Count(&files).Error)
	require.EqualValues(t, 1, files, "同 key 同内容只该有一行")
}

func TestRun_ExternalRefKeepsURLVerbatim(t *testing.T) {
	f := newFixture(t)
	const raw = "https://cdn.example.com/a.png?v=1"
	docs := []*capsule.EchoDoc{{
		ID:        echoIDPublic,
		CreatedAt: publicCreatedAt,
		Files:     []capsule.FileRef{{URL: raw}},
		Content:   "external\n",
	}}
	dir := writeCapsule(t, fullManifest(), docs, nil, nil)

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)
	require.Equal(t, 1, res.FilesCreated)

	var file fileModel.File
	require.NoError(t, f.db.Where("storage_type = ?", string(storage.StorageTypeExternal)).First(&file).Error)
	require.Equal(t, raw, file.URL, "外链 URL 原样透传，不本地化")
	require.Equal(t, string(storage.CategoryImage), file.Category, "缺省 category 由扩展名推导，query 不干扰")
	require.Equal(t, string(storage.StorageTypeExternal), file.Provider)
}

func TestRun_IncludePrivateLandsPrivateEchoAndItsComment(t *testing.T) {
	f := newFixture(t)
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{"pic.png": pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{IncludePrivate: true})
	require.NoError(t, err)
	require.Equal(t, 3, res.EchoesCreated)
	require.Equal(t, 0, res.SkippedPrivate)
	require.Equal(t, 2, res.CommentsCreated)
	require.Equal(t, 0, res.OrphanComments)
}

func TestRun_FreshInstanceTakesCapsuleSiteIdentity(t *testing.T) {
	f := newFixture(t)
	f.deps.KV = kvstore.NewMemory()

	pristine := coreSetting.System.Default()
	require.NotEmpty(t, pristine.SiteTitle, "前提：默认站点标题非空，否则本用例守不住任何东西")

	manifest := fullManifest()
	manifest.Site.SiteTitle = "搬家后的站点"
	dir := writeCapsule(t, manifest, fullDocs(), fullComments(), map[string][]byte{"pic.png": pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)
	require.Contains(t, res.SiteFieldsFilled, "site_title")

	sys, err := coreSetting.Get(context.Background(), f.deps.KV, coreSetting.System)
	require.NoError(t, err)
	require.Equal(t, "搬家后的站点", sys.SiteTitle)
}

func readStoredBytes(t *testing.T, f *fixture, key string) []byte {
	t.Helper()
	rc, err := f.deps.Selector.Get(context.Background(), storage.StorageTypeLocal, key)
	require.NoError(t, err)
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	return data
}

func snapshotCounts(t *testing.T, db *gorm.DB) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	for _, table := range []string{"echos", "echo_extensions", "tags", "echo_tags", "files", "echo_files", "comments", "connecteds"} {
		var n int64
		require.NoError(t, db.Table(table).Count(&n).Error)
		out[table] = n
	}
	return out
}

func nonZero(in map[string]int64) map[string]int64 {
	out := map[string]int64{}
	for k, v := range in {
		if v != 0 {
			out[k] = v
		}
	}
	return out
}
