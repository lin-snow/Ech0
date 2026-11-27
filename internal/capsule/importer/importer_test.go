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

	fileIDPublic = "01890000-0000-7000-8000-0000000000f1"
	fileIDOther  = "01890000-0000-7000-8000-0000000000f2"

	mediaKey = "media/pic.png"

	publicCreatedAt = "2024-03-05T06:07:08Z"
	publicCreatedTS = int64(1709618828)
)

var pngBytes = []byte("\x89PNG\r\n\x1a\nfake-image-payload")

type fixture struct {
	db      *gorm.DB
	deps    Deps
	store   storage.Store
	ownerID string
	aliceID string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	db := helpers.NewTestDB(t)
	registry := storage.NewRegistry()
	require.NoError(t, registry.Reload(context.Background(), []storage.Config{
		{ID: "mem", Driver: storage.DriverMem, Default: true},
	}))
	_, store, err := registry.Default()
	require.NoError(t, err)

	owner := userModel.User{Username: "owner", Role: userModel.RoleOwner}
	require.NoError(t, db.Create(&owner).Error)
	alice := userModel.User{Username: "alice"}
	require.NoError(t, db.Create(&alice).Error)

	kv := kvstore.NewMemory()
	require.NoError(t, coreSetting.System.Set(context.Background(), kv, settingModel.SystemSetting{SiteTitle: "Existing Title"}))

	return &fixture{
		db:      db,
		store:   store,
		ownerID: owner.ID,
		aliceID: alice.ID,
		deps: Deps{
			DB:      db,
			Tx:      transaction.NewGormTransactor(func() *gorm.DB { return db }),
			Storage: registry,
			KV:      kv,
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
				ID:          fileIDPublic,
				Key:         mediaKey,
				Category:    fileModel.CategoryImage,
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
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{mediaKey: pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)

	require.Equal(t, 2, res.EchoesCreated)
	require.Equal(t, 1, res.SkippedPrivate)
	require.Equal(t, 0, res.EchoesSkipped)
	require.Equal(t, 2, res.TagsCreated)
	require.Equal(t, 1, res.FilesCreated)
	require.Equal(t, 0, res.FilesReused)
	require.Equal(t, 1, res.CommentsCreated)
	require.Equal(t, 1, res.OrphanComments)

	var echo echoModel.Echo
	require.NoError(t, f.db.Where("id = ?", echoIDPublic).First(&echo).Error)
	require.Equal(t, publicCreatedTS, echo.CreatedAt)
	require.Equal(t, "alice", echo.Username)
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
	require.NoError(t, f.db.Where("id = ?", fileIDPublic).First(&file).Error)
	require.Equal(t, string(fileModel.KindOwned), file.Kind)
	require.Equal(t, string(fileModel.StatusReady), file.Status)
	require.Equal(t, string(fileModel.AccessOwner), file.Access)
	require.NotNil(t, file.StorageID)
	require.Equal(t, "mem", *file.StorageID)
	require.NotNil(t, file.ObjectKey)

	expectedKey, err := storage.KeyFor(fileIDPublic, ".png")
	require.NoError(t, err)
	require.Equal(t, expectedKey, *file.ObjectKey)
	require.Equal(t, "原图.png", file.Name)
	require.Equal(t, "image/png", file.ContentType)
	require.EqualValues(t, len(pngBytes), file.Size)
	require.Equal(t, 3, file.Width)
	require.Equal(t, 4, file.Height)
	require.Equal(t, f.aliceID, file.UploaderID)
	require.Nil(t, file.RefURL)
	require.Equal(t, fileModel.CategoryImage, fileModel.CategoryOf(file.ContentType))

	var link fileModel.Link
	require.NoError(t, f.db.Where("file_id = ? AND owner_id = ?", file.ID, echoIDPublic).First(&link).Error)
	require.Equal(t, fileModel.OwnerKindEcho, link.OwnerKind)
	require.Equal(t, fileModel.RoleAttachment, link.Role)
	require.Equal(t, 0, link.SortOrder)

	require.Equal(t, pngBytes, readStoredBytes(t, f, *file.ObjectKey))

	var blobs int64
	require.NoError(t, f.db.Model(&fileModel.Blob{}).Count(&blobs).Error)
	require.Zero(t, blobs)

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
	sys, err := coreSetting.System.Get(context.Background(), f.deps.KV)
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
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{mediaKey: pngBytes})

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
	require.Equal(t, 0, res.FilesReused)
	require.Equal(t, 0, res.CommentsCreated)
	require.Equal(t, 1, res.CommentsSkipped)
	require.Equal(t, 1, res.OrphanComments)
	require.Empty(t, res.SiteFieldsFilled, "站点设置已被首轮填满，第二轮无空位可填")

	require.Equal(t, before, snapshotCounts(t, f.db))
}

func TestRun_DryRunWritesNothing(t *testing.T) {
	f := newFixture(t)
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{mediaKey: pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{DryRun: true})
	require.NoError(t, err)

	require.Equal(t, 2, res.EchoesCreated)
	require.Equal(t, 1, res.FilesCreated)
	require.Equal(t, 2, res.TagsCreated)
	require.Equal(t, 1, res.CommentsCreated)
	require.Equal(t, 1, res.OrphanComments)
	require.NotEmpty(t, res.SiteFieldsFilled)

	require.Equal(t, map[string]int64{}, nonZero(snapshotCounts(t, f.db)))

	key, err := storage.KeyFor(fileIDPublic, ".png")
	require.NoError(t, err)
	_, err = f.store.Stat(context.Background(), key)
	require.Error(t, err)

	sys, err := coreSetting.System.Get(context.Background(), f.deps.KV)
	require.NoError(t, err)
	require.Equal(t, "Existing Title", sys.SiteTitle)
	require.Empty(t, sys.ServerName, "dry-run 不得写 KV：仓储写完会刷进程内缓存，回滚兜不住")
}

func TestRun_DedupesByObjectKey(t *testing.T) {
	f := newFixture(t)
	key, err := storage.KeyFor(fileIDPublic, ".png")
	require.NoError(t, err)
	require.NoError(t, f.db.Create(&fileModel.File{
		ID:          fileIDOther,
		Kind:        string(fileModel.KindOwned),
		Status:      string(fileModel.StatusReady),
		Access:      string(fileModel.AccessOwner),
		StorageID:   new("mem"),
		ObjectKey:   &key,
		Name:        "already-there.png",
		ContentType: "image/png",
		Size:        int64(len(pngBytes)),
		UploaderID:  f.ownerID,
	}).Error)

	docs := []*capsule.EchoDoc{{
		ID:        echoIDPublic,
		CreatedAt: publicCreatedAt,
		Username:  "alice",
		Files:     []capsule.FileRef{{ID: fileIDPublic, Key: mediaKey}},
		Content:   "dedupe\n",
	}}
	dir := writeCapsule(t, fullManifest(), docs, nil, map[string][]byte{mediaKey: pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)
	require.Equal(t, 0, res.FilesCreated)
	require.Equal(t, 1, res.FilesReused)

	var files int64
	require.NoError(t, f.db.Model(&fileModel.File{}).Count(&files).Error)
	require.EqualValues(t, 1, files)

	var link fileModel.Link
	require.NoError(t, f.db.Where("owner_id = ?", echoIDPublic).First(&link).Error)
	require.Equal(t, fileIDOther, link.FileID)
}

func TestRun_ReusesFileByID(t *testing.T) {
	f := newFixture(t)
	docs := []*capsule.EchoDoc{{
		ID:        echoIDPublic,
		CreatedAt: publicCreatedAt,
		Files:     []capsule.FileRef{{ID: fileIDPublic, Key: mediaKey}},
		Content:   "first\n",
	}}
	first := writeCapsule(t, fullManifest(), docs, nil, map[string][]byte{mediaKey: pngBytes})
	_, err := Run(context.Background(), f.deps, loadCapsule(t, first), Options{})
	require.NoError(t, err)

	docs[0].ID = echoIDGhost
	second := writeCapsule(t, fullManifest(), docs, nil, map[string][]byte{mediaKey: pngBytes})
	res, err := Run(context.Background(), f.deps, loadCapsule(t, second), Options{})
	require.NoError(t, err)

	require.Equal(t, 1, res.FilesReused)
	require.Equal(t, 0, res.FilesCreated)

	var files int64
	require.NoError(t, f.db.Model(&fileModel.File{}).Count(&files).Error)
	require.EqualValues(t, 1, files, "同 ID 只该有一行")

	var links int64
	require.NoError(t, f.db.Model(&fileModel.Link{}).Count(&links).Error)
	require.EqualValues(t, 2, links)
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
	require.NoError(t, f.db.Where("kind = ?", string(fileModel.KindRef)).First(&file).Error)
	require.NotNil(t, file.RefURL)
	require.Equal(t, raw, *file.RefURL, "外链 URL 原样透传，不本地化")
	require.Nil(t, file.ObjectKey)
	require.Nil(t, file.StorageID)
	require.Equal(t, fileModel.CategoryImage, fileModel.CategoryOf(file.ContentType), "缺省 contentType 由扩展名推导")
	require.Equal(t, "image/png", file.ContentType)
}

func TestRun_IncludePrivateLandsPrivateEchoAndItsComment(t *testing.T) {
	f := newFixture(t)
	dir := writeCapsule(t, fullManifest(), fullDocs(), fullComments(), map[string][]byte{mediaKey: pngBytes})

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

	pristine := coreSetting.System.DefaultValue()
	require.NotEmpty(t, pristine.SiteTitle, "前提：默认站点标题非空，否则本用例守不住任何东西")

	manifest := fullManifest()
	manifest.Site.SiteTitle = "搬家后的站点"
	dir := writeCapsule(t, manifest, fullDocs(), fullComments(), map[string][]byte{mediaKey: pngBytes})

	res, err := Run(context.Background(), f.deps, loadCapsule(t, dir), Options{})
	require.NoError(t, err)
	require.Contains(t, res.SiteFieldsFilled, "site_title")

	sys, err := coreSetting.System.Get(context.Background(), f.deps.KV)
	require.NoError(t, err)
	require.Equal(t, "搬家后的站点", sys.SiteTitle)
}

func readStoredBytes(t *testing.T, f *fixture, key string) []byte {
	t.Helper()
	rc, _, err := f.store.Open(context.Background(), key)
	require.NoError(t, err)
	defer func() { _ = rc.Close() }()
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	return data
}

func snapshotCounts(t *testing.T, db *gorm.DB) map[string]int64 {
	t.Helper()
	out := map[string]int64{}
	for _, table := range []string{"echos", "echo_extensions", "tags", "echo_tags", "files", "file_links", "file_blobs", "comments", "connecteds"} {
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
