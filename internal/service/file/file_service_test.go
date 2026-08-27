// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const fileTestUserID = "user-test-0001"

type fileFix struct {
	svc    *fileService.FileService
	common *commonmock.MockCommonRepository
	repo   *fileRepository.FileRepository
	db     *gorm.DB
	mgr    *storage.Manager
}

func newFileFix(t *testing.T) *fileFix {
	t.Helper()
	db := helpers.NewTestDB(t)
	repo := fileRepository.NewFileRepository(func() *gorm.DB { return db })
	tx := transaction.NewGormTransactor(func() *gorm.DB { return db })
	mgr := helpers.NewTestStorage(t)
	bus := helpers.NewTestBus(t)
	common := commonmock.NewMockCommonRepository(t)
	svc := fileService.NewFileService(tx, common, repo, mgr, func() *busen.Bus { return bus })
	return &fileFix{svc: svc, common: common, repo: repo, db: db, mgr: mgr}
}

func (f *fileFix) expectAdmin() {
	f.common.EXPECT().
		GetUserByUserId(mock.Anything, fileTestUserID).
		Return(helpers.NewUser(helpers.AsAdmin), nil)
}

func (f *fileFix) expectNonAdmin() {
	f.common.EXPECT().
		GetUserByUserId(mock.Anything, fileTestUserID).
		Return(helpers.NewUser(), nil)
}

func (f *fileFix) adminCtx() context.Context { return helpers.CtxAsUser(fileTestUserID) }

func makeFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(32<<20))
	headers := req.MultipartForm.File["file"]
	require.Len(t, headers, 1)
	return headers[0]
}

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func flacBytes() []byte {
	b := make([]byte, 64)
	copy(b, "fLaC")
	return b
}

func newGinCtx(t *testing.T, reqCtx context.Context) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if reqCtx != nil {
		req = req.WithContext(reqCtx)
	}
	c.Request = req
	return c, rec
}

func countFiles(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&fileModel.File{}).Count(&n).Error)
	return n
}

func countTemps(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Model(&fileModel.TempFile{}).Count(&n).Error)
	return n
}

func storedExists(t *testing.T, mgr *storage.Manager, key string) bool {
	t.Helper()
	rc, err := mgr.GetSelector().Get(context.Background(), storage.StorageTypeLocal, key)
	if err != nil {
		return false
	}
	_ = rc.Close()
	return true
}

func (f *fileFix) uploadPNG(t *testing.T, filename string, w, h int) commonModel.FileDto {
	t.Helper()
	f.expectAdmin()
	header := makeFileHeader(t, filename, pngBytes(t, w, h))
	dto, err := f.svc.UploadFile(f.adminCtx(), header, storage.CategoryImage, storage.StorageTypeLocal)
	require.NoError(t, err)
	return dto
}

func TestFileService_UploadFile(t *testing.T) {
	t.Run("image success persists and resolves local url", func(t *testing.T) {
		fix := newFileFix(t)
		content := pngBytes(t, 12, 8)
		fix.expectAdmin()

		dto, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "photo.png", content),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.NoError(t, err)

		assert.NotEmpty(t, dto.ID)
		assert.Equal(t, "photo.png", dto.Name)
		assert.Equal(t, "image", dto.Category)
		assert.Equal(t, "image/png", dto.ContentType)
		assert.Equal(t, "local", dto.StorageType)
		assert.Equal(t, int64(len(content)), dto.Size)
		assert.Equal(t, 12, dto.Width)
		assert.Equal(t, 8, dto.Height)
		assert.True(t, strings.HasSuffix(dto.Key, ".png"))
		assert.True(t, strings.HasPrefix(dto.URL, "/api/files/images/"))
		assert.True(t, strings.HasSuffix(dto.URL, dto.Key))

		assert.Equal(t, int64(1), countFiles(t, fix.db))
		assert.Equal(t, int64(1), countTemps(t, fix.db))
		assert.True(t, storedExists(t, fix.mgr, dto.Key))
	})

	t.Run("audio success uses octet-stream branch and audio route", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		dto, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "clip.flac", flacBytes()),
			storage.CategoryAudio,
			storage.StorageTypeLocal,
		)
		require.NoError(t, err)
		assert.Equal(t, "audio", dto.Category)
		assert.Equal(t, "audio/flac", dto.ContentType)
		assert.Equal(t, 0, dto.Width)
		assert.Equal(t, 0, dto.Height)
		assert.True(t, strings.HasPrefix(dto.URL, "/api/files/audios/"))
		assert.True(t, storedExists(t, fix.mgr, dto.Key))
	})

	t.Run("external storage type is downgraded to local", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		dto, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "photo.png", pngBytes(t, 4, 4)),
			storage.CategoryImage,
			storage.StorageTypeExternal,
		)
		require.NoError(t, err)
		assert.Equal(t, "local", dto.StorageType)
		assert.True(t, storedExists(t, fix.mgr, dto.Key))
	})

	t.Run("non-admin is denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()

		_, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "photo.png", pngBytes(t, 2, 2)),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
		assert.Equal(t, int64(0), countFiles(t, fix.db))
	})

	t.Run("user lookup error propagates", func(t *testing.T) {
		fix := newFileFix(t)
		sentinel := errors.New("db down")
		fix.common.EXPECT().
			GetUserByUserId(mock.Anything, fileTestUserID).
			Return(helpers.NewUser(), sentinel)

		_, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "photo.png", pngBytes(t, 2, 2)),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.ErrorIs(t, err, sentinel)
	})

	t.Run("dangerous extension rejected before storage write", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		_, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "evil.html", []byte("<!DOCTYPE html><html></html>")),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.Error(t, err)
		assert.Equal(t, commonModel.FILE_TYPE_NOT_ALLOWED, err.Error())
		assert.Equal(t, int64(0), countFiles(t, fix.db))
	})

	t.Run("empty file fails type validation via text sniff", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		_, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "empty.png", nil),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.Error(t, err)
		assert.Equal(t, commonModel.FILE_TYPE_NOT_ALLOWED, err.Error())
	})

	t.Run("size over limit rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		cfg := config.Config()
		prev := cfg.Upload.ImageMaxSize
		cfg.Upload.ImageMaxSize = 10
		t.Cleanup(func() { cfg.Upload.ImageMaxSize = prev })

		_, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "photo.png", pngBytes(t, 32, 32)),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.Error(t, err)
		assert.Equal(t, commonModel.FILE_SIZE_EXCEED_LIMIT, err.Error())
		assert.Equal(t, int64(0), countFiles(t, fix.db))
	})
}

func TestFileService_CreateExternalFile(t *testing.T) {
	t.Run("explicit image category then dedup returns same record", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		dto, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{
			URL:      "https://example.com/pic.png",
			Category: "image",
		})
		require.NoError(t, err)
		assert.Equal(t, "external", dto.StorageType)
		assert.Equal(t, "image", dto.Category)
		assert.Equal(t, "image/png", dto.ContentType)
		assert.Equal(t, "pic.png", dto.Name)
		assert.Equal(t, "https://example.com/pic.png", dto.URL)
		assert.True(t, strings.HasPrefix(dto.Key, "external/image/"))
		assert.Equal(t, int64(1), countFiles(t, fix.db))
		assert.Equal(t, int64(1), countTemps(t, fix.db))

		again, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{
			URL:      "https://example.com/pic.png",
			Category: "image",
		})
		require.NoError(t, err)
		assert.Equal(t, dto.ID, again.ID)
		assert.Equal(t, int64(1), countFiles(t, fix.db))
		assert.Equal(t, int64(1), countTemps(t, fix.db))
	})

	t.Run("omitted category normalizes to file", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		dto, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{
			URL: "https://example.com/blob",
		})
		require.NoError(t, err)
		assert.Equal(t, "file", dto.Category)
		assert.Equal(t, "application/octet-stream", dto.ContentType)
		assert.True(t, strings.HasPrefix(dto.Key, "external/file/"))
	})

	t.Run("explicit audio category retains declared content type", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()

		dto, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{
			URL:         "https://example.com/song",
			Category:    "audio",
			ContentType: "audio/mpeg",
		})
		require.NoError(t, err)
		assert.Equal(t, "audio", dto.Category)
		assert.Equal(t, "audio/mpeg", dto.ContentType)
	})

	t.Run("empty url rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{URL: ""})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("non-http scheme rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{
			URL: "ftp://example.com/x.png",
		})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("non-admin denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		_, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{
			URL: "https://example.com/pic.png",
		})
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})
}

func TestFileService_GetFileByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fix := newFileFix(t)
		uploaded := fix.uploadPNG(t, "photo.png", 6, 6)
		fix.expectAdmin()

		dto, err := fix.svc.GetFileByID(fix.adminCtx(), uploaded.ID)
		require.NoError(t, err)
		assert.Equal(t, uploaded.ID, dto.ID)
		assert.Equal(t, uploaded.Key, dto.Key)
		assert.Equal(t, "image", dto.Category)
	})

	t.Run("non-admin denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		_, err := fix.svc.GetFileByID(fix.adminCtx(), "whatever")
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})

	t.Run("not found propagates repo error", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.GetFileByID(fix.adminCtx(), "missing-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestFileService_UpdateFileMeta(t *testing.T) {
	newObjectFile := func(t *testing.T, fix *fileFix) *fileModel.File {
		t.Helper()
		f := &fileModel.File{
			Key:         "obj_key.png",
			StorageType: "object",
			Provider:    "r2",
			Bucket:      "bucket",
			URL:         "https://cdn.example.com/obj_key.png",
			Name:        "obj.png",
			ContentType: "image/png",
			Category:    "image",
			UserID:      fileTestUserID,
		}
		require.NoError(t, fix.db.Create(f).Error)
		return f
	}

	t.Run("success updates object metadata", func(t *testing.T) {
		fix := newFileFix(t)
		f := newObjectFile(t, fix)
		fix.expectAdmin()

		w, h := 320, 240
		dto, err := fix.svc.UpdateFileMeta(fix.adminCtx(), f.ID, commonModel.UpdateFileMetaDto{
			Size:        4096,
			Width:       &w,
			Height:      &h,
			ContentType: "image/webp",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(4096), dto.Size)
		assert.Equal(t, 320, dto.Width)
		assert.Equal(t, 240, dto.Height)
		assert.Equal(t, "image/webp", dto.ContentType)
	})

	t.Run("empty id rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.UpdateFileMeta(fix.adminCtx(), "", commonModel.UpdateFileMetaDto{Size: 1})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("negative size rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.UpdateFileMeta(fix.adminCtx(), "id", commonModel.UpdateFileMetaDto{Size: -1})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("negative width rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		bad := -5
		_, err := fix.svc.UpdateFileMeta(fix.adminCtx(), "id", commonModel.UpdateFileMetaDto{Size: 1, Width: &bad})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("non-object storage rejected", func(t *testing.T) {
		fix := newFileFix(t)
		uploaded := fix.uploadPNG(t, "photo.png", 5, 5)
		fix.expectAdmin()
		_, err := fix.svc.UpdateFileMeta(fix.adminCtx(), uploaded.ID, commonModel.UpdateFileMetaDto{Size: 10})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("non-admin denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		_, err := fix.svc.UpdateFileMeta(fix.adminCtx(), "id", commonModel.UpdateFileMetaDto{Size: 1})
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})
}

func TestFileService_ListFiles(t *testing.T) {
	t.Run("returns uploaded files with default pagination", func(t *testing.T) {
		fix := newFileFix(t)
		fix.uploadPNG(t, "a.png", 2, 2)
		fix.uploadPNG(t, "b.png", 2, 2)
		fix.expectAdmin()

		res, err := fix.svc.ListFiles(fix.adminCtx(), commonModel.FileListQueryDto{Page: 0, PageSize: 0})
		require.NoError(t, err)
		assert.Equal(t, int64(2), res.Total)
		assert.Len(t, res.Items, 2)
	})

	t.Run("search with no match returns empty", func(t *testing.T) {
		fix := newFileFix(t)
		fix.uploadPNG(t, "a.png", 2, 2)
		fix.expectAdmin()

		res, err := fix.svc.ListFiles(fix.adminCtx(), commonModel.FileListQueryDto{Search: "no-such-name-xyz"})
		require.NoError(t, err)
		assert.Equal(t, int64(0), res.Total)
		assert.Empty(t, res.Items)
	})

	t.Run("invalid storage type rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.ListFiles(fix.adminCtx(), commonModel.FileListQueryDto{StorageType: "external"})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("non-admin denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		_, err := fix.svc.ListFiles(fix.adminCtx(), commonModel.FileListQueryDto{})
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})
}

func TestFileService_ListFileTree(t *testing.T) {
	t.Run("root lists category folders", func(t *testing.T) {
		fix := newFileFix(t)
		fix.uploadPNG(t, "a.png", 2, 2)
		fix.expectAdmin()

		fix.expectAdmin()
		_, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "clip.flac", flacBytes()),
			storage.CategoryAudio,
			storage.StorageTypeLocal,
		)
		require.NoError(t, err)

		res, err := fix.svc.ListFileTree(fix.adminCtx(), commonModel.FileTreeQueryDto{StorageType: "local"})
		require.NoError(t, err)
		require.Len(t, res.Items, 2)
		for _, item := range res.Items {
			assert.Equal(t, "folder", item.NodeType)
			assert.True(t, item.HasChildren)
		}
	})

	t.Run("prefix lists files with resolved file id", func(t *testing.T) {
		fix := newFileFix(t)
		uploaded := fix.uploadPNG(t, "a.png", 2, 2)

		fix.expectAdmin()
		res, err := fix.svc.ListFileTree(fix.adminCtx(), commonModel.FileTreeQueryDto{
			StorageType: "local",
			Prefix:      "images",
		})
		require.NoError(t, err)
		require.Len(t, res.Items, 1)
		assert.Equal(t, "file", res.Items[0].NodeType)
		assert.Equal(t, uploaded.ID, res.Items[0].FileID)
	})

	t.Run("empty storage type rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.ListFileTree(fix.adminCtx(), commonModel.FileTreeQueryDto{StorageType: ""})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("invalid storage type rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.ListFileTree(fix.adminCtx(), commonModel.FileTreeQueryDto{StorageType: "external"})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("non-admin denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		_, err := fix.svc.ListFileTree(fix.adminCtx(), commonModel.FileTreeQueryDto{StorageType: "local"})
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})
}

func TestFileService_StreamFileByID(t *testing.T) {
	t.Run("serves local content", func(t *testing.T) {
		fix := newFileFix(t)
		content := pngBytes(t, 10, 10)
		fix.expectAdmin()
		dto, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "photo.png", content),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.NoError(t, err)

		c, rec := newGinCtx(t, nil)
		fix.svc.StreamFileByID(c, dto.ID)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "image/png")
		assert.Equal(t, content, rec.Body.Bytes())
	})

	t.Run("missing file returns 404", func(t *testing.T) {
		fix := newFileFix(t)
		c, rec := newGinCtx(t, nil)
		fix.svc.StreamFileByID(c, "missing-id")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("external file redirects", func(t *testing.T) {
		fix := newFileFix(t)
		ext := &fileModel.File{
			Key:         "external/image/abc",
			StorageType: "external",
			Provider:    "external",
			URL:         "https://example.com/x.png",
			Name:        "x.png",
			ContentType: "image/png",
			Category:    "image",
			UserID:      fileTestUserID,
		}
		require.NoError(t, fix.db.Create(ext).Error)

		c, rec := newGinCtx(t, nil)
		fix.svc.StreamFileByID(c, ext.ID)
		c.Writer.WriteHeaderNow()
		assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
		assert.Equal(t, "https://example.com/x.png", rec.Header().Get("Location"))
	})
}

func TestFileService_StreamFileByPath(t *testing.T) {
	t.Run("serves by storage path", func(t *testing.T) {
		fix := newFileFix(t)
		content := pngBytes(t, 9, 9)
		fix.expectAdmin()
		dto, err := fix.svc.UploadFile(
			fix.adminCtx(),
			makeFileHeader(t, "photo.png", content),
			storage.CategoryImage,
			storage.StorageTypeLocal,
		)
		require.NoError(t, err)

		fix.expectAdmin()
		c, rec := newGinCtx(t, fix.adminCtx())
		fix.svc.StreamFileByPath(c, commonModel.FilePathStreamQueryDto{
			StorageType: "local",
			Path:        "images/" + dto.Key,
			ContentType: "image/png",
		})
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, content, rec.Body.Bytes())
	})

	t.Run("non-admin forbidden", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		c, rec := newGinCtx(t, fix.adminCtx())
		fix.svc.StreamFileByPath(c, commonModel.FilePathStreamQueryDto{
			StorageType: "local",
			Path:        "images/whatever.png",
		})
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("invalid storage type", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		c, rec := newGinCtx(t, fix.adminCtx())
		fix.svc.StreamFileByPath(c, commonModel.FilePathStreamQueryDto{
			StorageType: "external",
			Path:        "images/whatever.png",
		})
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("empty path", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		c, rec := newGinCtx(t, fix.adminCtx())
		fix.svc.StreamFileByPath(c, commonModel.FilePathStreamQueryDto{
			StorageType: "local",
			Path:        "   ",
		})
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing file 404", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		c, rec := newGinCtx(t, fix.adminCtx())
		fix.svc.StreamFileByPath(c, commonModel.FilePathStreamQueryDto{
			StorageType: "local",
			Path:        "images/does-not-exist.png",
		})
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestFileService_GetFilePresignURL(t *testing.T) {
	t.Run("object storage disabled surfaces error", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.GetFilePresignURL(fix.adminCtx(), &commonModel.GetPresignURLDto{
			FileName:    "pic.png",
			ContentType: "image/png",
		})
		require.Error(t, err)
	})

	t.Run("empty filename rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.GetFilePresignURL(fix.adminCtx(), &commonModel.GetPresignURLDto{FileName: ""})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("non-object storage type rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.GetFilePresignURL(fix.adminCtx(), &commonModel.GetPresignURLDto{
			FileName:    "pic.png",
			StorageType: "local",
		})
		require.Error(t, err)
		assert.Equal(t, commonModel.INVALID_PARAMS, err.Error())
	})

	t.Run("disallowed type rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		_, err := fix.svc.GetFilePresignURL(fix.adminCtx(), &commonModel.GetPresignURLDto{
			FileName:    "doc.pdf",
			ContentType: "application/pdf",
		})
		require.Error(t, err)
		assert.Equal(t, commonModel.FILE_TYPE_NOT_ALLOWED, err.Error())
	})

	t.Run("non-admin denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		_, err := fix.svc.GetFilePresignURL(fix.adminCtx(), &commonModel.GetPresignURLDto{FileName: "pic.png"})
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})
}

func TestFileService_DeleteFile(t *testing.T) {
	t.Run("local file removes record and blob", func(t *testing.T) {
		fix := newFileFix(t)
		dto := fix.uploadPNG(t, "photo.png", 3, 3)
		require.True(t, storedExists(t, fix.mgr, dto.Key))

		fix.expectAdmin()
		require.NoError(t, fix.svc.DeleteFile(fix.adminCtx(), dto.ID))
		assert.Equal(t, int64(0), countFiles(t, fix.db))
		assert.False(t, storedExists(t, fix.mgr, dto.Key))
	})

	t.Run("external file removes record only", func(t *testing.T) {
		fix := newFileFix(t)
		ext := &fileModel.File{
			Key:         "external/image/abc",
			StorageType: "external",
			Provider:    "external",
			URL:         "https://example.com/x.png",
			Name:        "x.png",
			Category:    "image",
			UserID:      fileTestUserID,
		}
		require.NoError(t, fix.db.Create(ext).Error)

		fix.expectAdmin()
		require.NoError(t, fix.svc.DeleteFile(fix.adminCtx(), ext.ID))
		assert.Equal(t, int64(0), countFiles(t, fix.db))
	})

	t.Run("empty id rejected", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		err := fix.svc.DeleteFile(fix.adminCtx(), "")
		require.Error(t, err)
		assert.Equal(t, commonModel.IMAGE_NOT_FOUND, err.Error())
	})

	t.Run("missing id propagates repo error", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		err := fix.svc.DeleteFile(fix.adminCtx(), "missing-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("non-admin denied", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectNonAdmin()
		err := fix.svc.DeleteFile(fix.adminCtx(), "id")
		require.Error(t, err)
		assert.Equal(t, commonModel.NO_PERMISSION_DENIED, err.Error())
	})
}

func TestFileService_DeleteStoredFile(t *testing.T) {
	t.Run("empty key is no-op", func(t *testing.T) {
		fix := newFileFix(t)
		require.NoError(t, fix.svc.DeleteStoredFile("local", ""))
	})

	t.Run("external is no-op", func(t *testing.T) {
		fix := newFileFix(t)
		require.NoError(t, fix.svc.DeleteStoredFile("external", "some/key"))
	})

	t.Run("local removes blob", func(t *testing.T) {
		fix := newFileFix(t)
		dto := fix.uploadPNG(t, "photo.png", 3, 3)
		require.True(t, storedExists(t, fix.mgr, dto.Key))
		require.NoError(t, fix.svc.DeleteStoredFile("local", dto.Key))
		assert.False(t, storedExists(t, fix.mgr, dto.Key))
	})
}

func TestFileService_DeleteFileRecord(t *testing.T) {
	fix := newFileFix(t)
	dto := fix.uploadPNG(t, "photo.png", 3, 3)
	require.NoError(t, fix.svc.DeleteFileRecord(context.Background(), dto.ID))
	assert.Equal(t, int64(0), countFiles(t, fix.db))
	assert.True(t, storedExists(t, fix.mgr, dto.Key))
}

func TestFileService_ConfirmTempFiles(t *testing.T) {
	t.Run("removes temp tracking but keeps file", func(t *testing.T) {
		fix := newFileFix(t)
		dto := fix.uploadPNG(t, "photo.png", 3, 3)
		require.Equal(t, int64(1), countTemps(t, fix.db))

		require.NoError(t, fix.svc.ConfirmTempFiles(context.Background(), []string{dto.ID}))
		assert.Equal(t, int64(0), countTemps(t, fix.db))
		assert.Equal(t, int64(1), countFiles(t, fix.db))
	})

	t.Run("blank and duplicate ids are skipped", func(t *testing.T) {
		fix := newFileFix(t)
		dto := fix.uploadPNG(t, "photo.png", 3, 3)
		require.NoError(t, fix.svc.ConfirmTempFiles(context.Background(), []string{"", "  ", dto.ID, dto.ID}))
		assert.Equal(t, int64(0), countTemps(t, fix.db))
	})
}

func TestFileService_CleanupOrphanFiles(t *testing.T) {
	expireTemp := func(t *testing.T, fix *fileFix, fileID string) {
		t.Helper()
		past := time.Now().UTC().Add(-time.Hour).Unix()
		require.NoError(t, fix.db.Model(&fileModel.TempFile{}).
			Where("file_id = ?", fileID).
			Update("expire_at", past).Error)
	}

	t.Run("no expired temps is a no-op", func(t *testing.T) {
		fix := newFileFix(t)
		require.NoError(t, fix.svc.CleanupOrphanFiles())
	})

	t.Run("deletes expired temp file and blob", func(t *testing.T) {
		fix := newFileFix(t)
		dto := fix.uploadPNG(t, "photo.png", 3, 3)
		expireTemp(t, fix, dto.ID)

		require.NoError(t, fix.svc.CleanupOrphanFiles())
		assert.Equal(t, int64(0), countFiles(t, fix.db))
		assert.Equal(t, int64(0), countTemps(t, fix.db))
		assert.False(t, storedExists(t, fix.mgr, dto.Key))
	})

	t.Run("dry-run keeps everything", func(t *testing.T) {
		fix := newFileFix(t)
		dto := fix.uploadPNG(t, "photo.png", 3, 3)
		expireTemp(t, fix, dto.ID)

		t.Setenv("ECH0_FILE_TEMP_CLEANUP_DRY_RUN", "true")
		require.NoError(t, fix.svc.CleanupOrphanFiles())
		assert.Equal(t, int64(1), countFiles(t, fix.db))
		assert.Equal(t, int64(1), countTemps(t, fix.db))
		assert.True(t, storedExists(t, fix.mgr, dto.Key))
	})

	t.Run("deletes expired unconfirmed external record", func(t *testing.T) {
		fix := newFileFix(t)
		fix.expectAdmin()
		dto, err := fix.svc.CreateExternalFile(fix.adminCtx(), commonModel.CreateExternalFileDto{
			URL:      "https://example.com/orphan.png",
			Category: "image",
		})
		require.NoError(t, err)
		expireTemp(t, fix, dto.ID)

		require.NoError(t, fix.svc.CleanupOrphanFiles())
		assert.Equal(t, int64(0), countFiles(t, fix.db))
		assert.Equal(t, int64(0), countTemps(t, fix.db))
	})
}
