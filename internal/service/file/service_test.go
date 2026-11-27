// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"testing"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/storage/storetest"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const pngBody = "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89"

type allowAll struct{}

func (allowAll) Kind() string { return fileModel.OwnerKindEcho }

func (allowAll) Authorize(context.Context, string, viewer.Context) (bool, error) {
	return true, nil
}

type ownerVisible struct{}

func (ownerVisible) Kind() string { return fileModel.OwnerKindEcho }

func (ownerVisible) Authorize(context.Context, string, viewer.Context) (bool, error) {
	return false, nil
}

type fixture struct {
	svc   *fileService.FileService
	store storage.Store
	db    *gorm.DB
}

func newFixture(t *testing.T, dbName string, authorizer fileService.Authorizer) *fixture {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", dbName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	database.SetDB(db)
	if err := database.MigrateDB(); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewFileResourceMigrator()),
	)

	ctx := context.Background()
	if err := db.Create(&fileModel.StorageConfig{
		ID: "mem", Driver: storage.DriverMem, Name: "mem", IsDefault: true,
	}).Error; err != nil {
		t.Fatalf("seed storage row: %v", err)
	}
	registry := storage.NewRegistry()
	authorizers := fileService.NewAuthorizerRegistry()
	if authorizer != nil {
		authorizers.Register(authorizer)
	}

	repo := fileRepository.NewFileRepository(func() *gorm.DB { return db })
	svc := fileService.NewFileService(transaction.NewGormTransactor(func() *gorm.DB { return db }), repo, registry, authorizers)
	if err := svc.ReloadStorages(ctx); err != nil {
		t.Fatalf("reload storages: %v", err)
	}

	return &fixture{svc: svc, store: mustStore(t, registry), db: db}
}

func mustStore(t *testing.T, registry *storage.Registry) storage.Store {
	t.Helper()
	_, store, err := registry.Default()
	if err != nil {
		t.Fatalf("default store: %v", err)
	}
	return store
}

// uploadHeader builds the multipart part the HTTP layer would hand to Create.
func uploadHeader(t *testing.T, name string, body string) *multipart.FileHeader {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(body)); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	form, err := multipart.NewReader(&buf, writer.Boundary()).ReadForm(int64(buf.Len()) + 1024)
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	return form.File["file"][0]
}

func TestCreate_OwnedIsReadyAndOnlyVisibleToItsUploader(t *testing.T) {
	f := newFixture(t, "file_owned", ownerVisible{})
	uploader := helpers.CtxAsUser("u-1")

	created, err := f.svc.Create(uploader, fileService.CreateInput{File: uploadHeader(t, "a.png", pngBody)})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got := created.File
	if got.Kind != string(fileModel.KindOwned) || got.Status != string(fileModel.StatusReady) || got.Access != string(fileModel.AccessOwner) {
		t.Fatalf("row = kind %q status %q access %q, want owned/ready/owner", got.Kind, got.Status, got.Access)
	}
	if got.Category != fileModel.CategoryImage || got.ContentType != "image/png" {
		t.Fatalf("category/content-type = %q/%q, want image/image/png", got.Category, got.ContentType)
	}
	if got.URL != fileModel.ContentPath(got.ID) {
		t.Fatalf("url = %q, want %q", got.URL, fileModel.ContentPath(got.ID))
	}

	if _, err := f.svc.Get(uploader, got.ID); err != nil {
		t.Fatalf("uploader get: %v", err)
	}
	if _, err := f.svc.Get(helpers.CtxAnonymous(), got.ID); !errors.Is(err, fileService.ErrNotFound) {
		t.Fatalf("anonymous get err = %v, want ErrNotFound", err)
	}
	if _, err := f.svc.Get(helpers.CtxAsUser("u-2"), got.ID); !errors.Is(err, fileService.ErrNotFound) {
		t.Fatalf("other user get err = %v, want ErrNotFound", err)
	}

	opened, err := f.svc.Open(uploader, got.ID, nil)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = opened.Reader.Close() }()
	if opened.CacheControl != "private, no-store" {
		t.Fatalf("cache-control = %q, want private, no-store", opened.CacheControl)
	}
	body := make([]byte, 64)
	n, _ := opened.Reader.Read(body)
	if string(body[:n]) != pngBody {
		t.Fatalf("content = %q, want the uploaded bytes", body[:n])
	}
}

func TestOpen_AnonymousOwnerFileIsNotImmutableCached(t *testing.T) {
	f := newFixture(t, "file_anon_owner", allowAll{})
	ctx := helpers.CtxAsUser("u-1")

	created, err := f.svc.Create(ctx, fileService.CreateInput{File: uploadHeader(t, "a.png", pngBody)})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.svc.Replace(ctx, []string{created.File.ID}, fileModel.OwnerKindEcho, "e-1", fileModel.RoleAttachment); err != nil {
		t.Fatalf("replace: %v", err)
	}

	opened, err := f.svc.Open(helpers.CtxAnonymous(), created.File.ID, nil)
	if err != nil {
		t.Fatalf("anonymous open: %v", err)
	}
	defer func() { _ = opened.Reader.Close() }()
	if opened.CacheControl != "public, max-age=0, must-revalidate" || !opened.VaryAuthorization {
		t.Fatalf("cache-control = %q vary = %v, want revalidate + Vary", opened.CacheControl, opened.VaryAuthorization)
	}
}

func TestCreate_RangeServesRequestedWindow(t *testing.T) {
	f := newFixture(t, "file_range", nil)
	ctx := helpers.CtxAsUser("u-1")

	created, err := f.svc.Create(ctx, fileService.CreateInput{File: uploadHeader(t, "a.png", pngBody)})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	opened, err := f.svc.Open(ctx, created.File.ID, &fileService.ByteRange{Start: 1, Length: 4})
	if err != nil {
		t.Fatalf("open range: %v", err)
	}
	defer func() { _ = opened.Reader.Close() }()
	got := make([]byte, 4)
	if _, err := opened.Reader.Read(got); err != nil {
		t.Fatalf("read range: %v", err)
	}
	if string(got) != pngBody[1:5] {
		t.Fatalf("range = %q, want %q", got, pngBody[1:5])
	}

	if _, err := f.svc.Open(ctx, created.File.ID, &fileService.ByteRange{Start: 4096}); !errors.Is(err, fileService.ErrRangeNotSatisfiable) {
		t.Fatalf("out-of-range err = %v, want ErrRangeNotSatisfiable", err)
	}
}

func TestCreate_InstanceProfileIsBlobAndCascades(t *testing.T) {
	f := newFixture(t, "file_blob", nil)
	admin := helpers.CtxAsUser("admin-1", role.Admin)

	created, err := f.svc.Create(admin, fileService.CreateInput{File: uploadHeader(t, "logo.png", pngBody), Profile: "instance"})
	if err != nil {
		t.Fatalf("create instance: %v", err)
	}
	if created.File.Kind != string(fileModel.KindBlob) || created.File.Access != string(fileModel.AccessPublic) {
		t.Fatalf("kind/access = %q/%q, want blob/public", created.File.Kind, created.File.Access)
	}
	var blob fileModel.Blob
	if err := f.db.Where("file_id = ?", created.File.ID).First(&blob).Error; err != nil {
		t.Fatalf("blob row missing: %v", err)
	}
	if !bytes.Equal(blob.Data, []byte(pngBody)) {
		t.Fatalf("blob bytes = %q, want the uploaded body", blob.Data)
	}

	opened, err := f.svc.Open(helpers.CtxAnonymous(), created.File.ID, nil)
	if err != nil {
		t.Fatalf("anonymous open blob: %v", err)
	}
	defer func() { _ = opened.Reader.Close() }()
	if opened.CacheControl != "public, max-age=31536000, immutable" {
		t.Fatalf("cache-control = %q, want immutable", opened.CacheControl)
	}

	if _, err := f.svc.Delete(admin, created.File.ID, false); err != nil {
		t.Fatalf("delete: %v", err)
	}
	var blobs int64
	if err := f.db.Model(&fileModel.Blob{}).Where("file_id = ?", created.File.ID).Count(&blobs).Error; err != nil {
		t.Fatalf("count blobs: %v", err)
	}
	if blobs != 0 {
		t.Fatalf("blob rows = %d, want the foreign key to cascade", blobs)
	}
}

func TestOpen_PublicOwnedFileRedirectsToCDN(t *testing.T) {
	f := newFixture(t, "file_cdn", nil)
	admin := helpers.CtxAsUser("admin-1", role.Admin)

	created, err := f.svc.Create(admin, fileService.CreateInput{File: uploadHeader(t, "a.png", pngBody), Profile: "avatar"})
	if err != nil {
		t.Fatalf("create avatar: %v", err)
	}

	// Without a CDN origin the gateway streams the bytes itself.
	opened, err := f.svc.Open(helpers.CtxAnonymous(), created.File.ID, nil)
	if err != nil {
		t.Fatalf("anonymous open avatar: %v", err)
	}
	_ = opened.Reader.Close()
	if opened.Redirect != "" {
		t.Fatalf("redirect without cdn = %q, want streaming", opened.Redirect)
	}

	if err := f.db.Model(&fileModel.StorageConfig{}).Where("id = ?", "mem").
		Update("config", `{"cdn_url":"https://cdn.example.com"}`).Error; err != nil {
		t.Fatalf("set cdn: %v", err)
	}
	if err := f.svc.ReloadStorages(context.Background()); err != nil {
		t.Fatalf("reload storages: %v", err)
	}

	opened, err = f.svc.Open(helpers.CtxAnonymous(), created.File.ID, nil)
	if err != nil {
		t.Fatalf("anonymous open with cdn: %v", err)
	}
	var row fileModel.File
	if err := f.db.Where("id = ?", created.File.ID).First(&row).Error; err != nil {
		t.Fatalf("load row: %v", err)
	}
	if want := "https://cdn.example.com/" + *row.ObjectKey; opened.Redirect != want {
		t.Fatalf("redirect = %q, want %q", opened.Redirect, want)
	}
	if opened.Reader != nil {
		_ = opened.Reader.Close()
		t.Fatal("a redirected public file must not also open a reader")
	}
}

func TestCreate_ExternalReferenceRedirectsOnlyWhenPublic(t *testing.T) {
	f := newFixture(t, "file_ref", nil)
	admin := helpers.CtxAsUser("admin-1", role.Admin)

	public, err := f.svc.Create(admin, fileService.CreateInput{
		Profile:     "avatar",
		ExternalURL: "https://cdn.example.com/a.png",
		ContentType: "image/png",
		Direct:      true,
	})
	if err != nil {
		t.Fatalf("create ref: %v", err)
	}
	if public.File.Kind != string(fileModel.KindRef) || public.File.Access != string(fileModel.AccessPublic) {
		t.Fatalf("kind/access = %q/%q, want ref/public", public.File.Kind, public.File.Access)
	}

	opened, err := f.svc.Open(helpers.CtxAnonymous(), public.File.ID, nil)
	if err != nil {
		t.Fatalf("anonymous open ref: %v", err)
	}
	if opened.Redirect != "https://cdn.example.com/a.png" {
		t.Fatalf("redirect = %q, want the stored URL", opened.Redirect)
	}

	owned, err := f.svc.Create(admin, fileService.CreateInput{ExternalURL: "https://cdn.example.com/b.png", ContentType: "image/png"})
	if err != nil {
		t.Fatalf("create owner ref: %v", err)
	}
	if _, err := f.svc.Open(helpers.CtxAnonymous(), owned.File.ID, nil); !errors.Is(err, fileService.ErrNotFound) {
		t.Fatalf("anonymous open owner ref err = %v, want ErrNotFound (no redirect)", err)
	}
}

func TestCreate_DirectIsRejectedForPublicProfiles(t *testing.T) {
	f := newFixture(t, "file_direct_profile", nil)
	admin := helpers.CtxAsUser("admin-1", role.Admin)

	if _, err := f.svc.Create(admin, fileService.CreateInput{Profile: "avatar", Direct: true, Name: "a.png"}); !errors.Is(err, fileService.ErrUnsupported) {
		t.Fatalf("direct avatar err = %v, want ErrUnsupported", err)
	}
}

func TestCreate_RejectsDangerousUploads(t *testing.T) {
	f := newFixture(t, "file_reject", nil)
	ctx := helpers.CtxAsUser("u-1")

	if _, err := f.svc.Create(ctx, fileService.CreateInput{File: uploadHeader(t, "evil.svg", "<svg/>")}); !errors.Is(err, fileService.ErrTypeNotAllowed) {
		t.Fatalf("svg upload err = %v, want ErrTypeNotAllowed", err)
	}
	if _, err := f.svc.Create(ctx, fileService.CreateInput{File: uploadHeader(t, "shell.sh", "#!/bin/sh\n")}); !errors.Is(err, fileService.ErrTypeNotAllowed) {
		t.Fatalf("script upload err = %v, want ErrTypeNotAllowed", err)
	}
}

func TestDelete_ConflictUntilForced(t *testing.T) {
	f := newFixture(t, "file_delete", allowAll{})
	uploader := helpers.CtxAsUser("u-1")
	admin := helpers.CtxAsUser("admin-1", role.Admin)

	created, err := f.svc.Create(uploader, fileService.CreateInput{File: uploadHeader(t, "a.png", pngBody)})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.svc.Replace(uploader, []string{created.File.ID}, fileModel.OwnerKindEcho, "e-1", fileModel.RoleAttachment); err != nil {
		t.Fatalf("replace: %v", err)
	}

	refs, err := f.svc.Delete(uploader, created.File.ID, false)
	var conflict *fileService.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("delete err = %v, want ConflictError", err)
	}
	if len(refs) != 1 || refs[0].OwnerID != "e-1" || refs[0].Role != fileModel.RoleAttachment {
		t.Fatalf("references = %+v, want the echo attachment", refs)
	}

	var row fileModel.File
	if err := f.db.Where("id = ?", created.File.ID).First(&row).Error; err != nil {
		t.Fatalf("load row: %v", err)
	}
	key := *row.ObjectKey
	if _, err := f.svc.Delete(uploader, created.File.ID, true); !errors.As(err, &conflict) {
		t.Fatalf("force delete by non-admin err = %v, want ConflictError", err)
	}

	if _, err := f.svc.Delete(admin, created.File.ID, true); err != nil {
		t.Fatalf("admin force delete: %v", err)
	}
	if _, err := f.svc.Get(admin, created.File.ID); !errors.Is(err, fileService.ErrNotFound) {
		t.Fatalf("row survived delete: %v", err)
	}
	// The row goes first; the bytes are reclaimed by the gc queue, never by the
	// request that deleted the row.
	if _, err := f.store.Stat(context.Background(), key); err != nil {
		t.Fatalf("bytes must outlive the row until gc runs: %v", err)
	}
	if _, err := f.svc.DrainGC(context.Background(), 10); err != nil {
		t.Fatalf("drain gc: %v", err)
	}
	if _, err := f.store.Stat(context.Background(), key); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("object not collected: %v", err)
	}
}

func TestSweepPending_DropsStaleUploadsAndTheirBytes(t *testing.T) {
	f := newFixture(t, "file_pending", nil)
	ctx := helpers.CtxAsUser("u-1")
	key, err := storage.KeyFor("0192f0c4-0000-7000-8000-0000000000aa", ".png")
	if err != nil {
		t.Fatalf("key: %v", err)
	}

	if _, err := f.store.Put(ctx, key, bytes.NewReader([]byte(pngBody)), int64(len(pngBody))); err != nil {
		t.Fatalf("put: %v", err)
	}
	storageID := "mem"
	row := fileModel.File{
		ID:          "0192f0c4-0000-7000-8000-0000000000aa",
		Kind:        string(fileModel.KindOwned),
		Status:      string(fileModel.StatusPending),
		Access:      string(fileModel.AccessOwner),
		StorageID:   &storageID,
		ObjectKey:   &key,
		ContentType: "image/png",
		Size:        int64(len(pngBody)),
		UploaderID:  "u-1",
		CreatedAt:   time.Now().Add(-2 * time.Hour).Unix(),
	}
	if err := f.db.Create(&row).Error; err != nil {
		t.Fatalf("insert pending: %v", err)
	}
	// A pending direct upload is swept only after its TTL; nothing is collected
	// here, which is what keeps an in-flight presigned upload alive.
	if _, err := f.svc.SweepPending(ctx, 10); err != nil {
		t.Fatalf("sweep before ttl: %v", err)
	}
	if err := f.db.Model(&fileModel.File{}).Where("id = ?", row.ID).Update("created_at", time.Now().Add(-25*time.Hour).Unix()).Error; err != nil {
		t.Fatalf("backdate: %v", err)
	}
	if _, err := f.svc.SweepPending(ctx, 10); err != nil {
		t.Fatalf("sweep: %v", err)
	}
	var rows int64
	if err := f.db.Model(&fileModel.File{}).Where("id = ?", row.ID).Count(&rows).Error; err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if rows != 0 {
		t.Fatalf("pending row survived sweep")
	}
	if _, err := f.svc.DrainGC(ctx, 10); err != nil {
		t.Fatalf("drain gc: %v", err)
	}
	if _, err := f.store.Stat(ctx, key); !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("swept object still present: %v", err)
	}
}

func TestSchema_EnforcesLocatorSlotsAndLinks(t *testing.T) {
	f := newFixture(t, "file_schema", nil)
	store := "mem"
	key := "2026/11/x.png"
	fileID := "0192f0c4-0000-7000-8000-0000000000bb"

	if err := f.db.Create(&fileModel.File{
		ID: fileID, Kind: "nonsense", Status: string(fileModel.StatusReady), Access: string(fileModel.AccessOwner),
		StorageID: &store, ObjectKey: &key, ContentType: "image/png", UploaderID: "u-1",
	}).Error; err == nil {
		t.Fatal("CHECK on files.kind did not reject an unknown kind")
	}
	// A owned row without a locator is not a row.
	if err := f.db.Create(&fileModel.File{
		ID: fileID, Kind: string(fileModel.KindOwned), Status: string(fileModel.StatusReady),
		Access: string(fileModel.AccessOwner), ContentType: "image/png", UploaderID: "u-1",
	}).Error; err == nil {
		t.Fatal("CHECK on files locator did not reject owned without storage/key")
	}

	first := fileModel.File{ID: "0192f0c4-0000-7000-8000-0000000000c1", Kind: string(fileModel.KindOwned), Status: string(fileModel.StatusReady), Access: string(fileModel.AccessOwner), StorageID: &store, ObjectKey: strPtr("k1"), ContentType: "image/png", UploaderID: "u-1"}
	second := fileModel.File{ID: "0192f0c4-0000-7000-8000-0000000000c2", Kind: string(fileModel.KindOwned), Status: string(fileModel.StatusReady), Access: string(fileModel.AccessOwner), StorageID: &store, ObjectKey: strPtr("k2"), ContentType: "image/png", UploaderID: "u-1"}
	for _, row := range []fileModel.File{first, second} {
		copy := row
		if err := f.db.Create(&copy).Error; err != nil {
			t.Fatalf("insert %s: %v", row.ID, err)
		}
	}

	avatar := func(fileID string) fileModel.Link {
		return fileModel.Link{FileID: fileID, OwnerKind: fileModel.OwnerKindUser, OwnerID: "u-1", Role: fileModel.RoleAvatar, CreatedAt: time.Now().Unix()}
	}
	if err := f.db.Create(&[]fileModel.Link{avatar(first.ID), avatar(second.ID)}).Error; err == nil {
		t.Fatal("uq_file_links_slot did not reject a second avatar for one owner")
	}

	single := avatar(first.ID)
	if err := f.db.Create(&single).Error; err != nil {
		t.Fatalf("insert avatar link: %v", err)
	}
	if err := f.db.Delete(&fileModel.File{}, "id = ?", first.ID).Error; err != nil {
		t.Fatalf("delete file: %v", err)
	}
	var links int64
	if err := f.db.Model(&fileModel.Link{}).Where("file_id = ?", first.ID).Count(&links).Error; err != nil {
		t.Fatalf("count links: %v", err)
	}
	if links != 0 {
		t.Fatalf("links = %d, want the foreign key to cascade", links)
	}
}

func TestStores_Conform(t *testing.T) {
	t.Run("mem", func(t *testing.T) {
		storetest.Run(t, func(t *testing.T) storage.Store { return storage.NewMemStore() })
	})
	t.Run("local", func(t *testing.T) {
		storetest.Run(t, func(t *testing.T) storage.Store {
			store, err := storage.OpenDriver(context.Background(), storage.Config{
				ID: "local", Driver: storage.DriverLocal, Config: map[string]string{"root": t.TempDir()},
			})
			if err != nil {
				t.Fatalf("open local store: %v", err)
			}
			return store
		})
	})
}

func strPtr(v string) *string { return &v }
