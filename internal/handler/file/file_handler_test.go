// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	filemock "github.com/lin-snow/ech0/internal/test/mocks/filemock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newGinCtx(t *testing.T, rawURL string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, rawURL, nil)
	return c, rec
}

func TestStreamFileByID_EmptyID_NoServiceCall(t *testing.T) {
	mockSvc := filemock.NewMockService(t)
	h := NewFileHandler(mockSvc)

	c, _ := newGinCtx(t, "/")
	h.StreamFileByID(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestStreamFileByID_DelegatesToService(t *testing.T) {
	cases := []struct {
		name         string
		simulate     func(*gin.Context)
		wantCode     int
		wantLocation string
	}{
		{"ok", func(c *gin.Context) { c.Status(http.StatusOK) }, http.StatusOK, ""},
		{"auth-rejected", func(c *gin.Context) { c.Status(http.StatusForbidden) }, http.StatusForbidden, ""},
		{"not-found", func(c *gin.Context) { c.Status(http.StatusNotFound) }, http.StatusNotFound, ""},
		{
			"external-redirect",
			func(c *gin.Context) { c.Redirect(http.StatusFound, "https://cdn.example/x.png") },
			http.StatusFound,
			"https://cdn.example/x.png",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := filemock.NewMockService(t)
			var gotID string
			var gotCtx *gin.Context
			mockSvc.EXPECT().
				StreamFileByID(mock.Anything, "file-1").
				Run(func(ctx *gin.Context, id string) {
					gotID = id
					gotCtx = ctx
					tc.simulate(ctx)
				}).
				Return().
				Once()

			h := NewFileHandler(mockSvc)
			c, rec := newGinCtx(t, "/files/file-1")
			c.Params = gin.Params{{Key: "id", Value: "file-1"}}

			h.StreamFileByID(c)

			assert.Equal(t, "file-1", gotID, "id 应原样透传")
			assert.Same(t, c, gotCtx, "应透传同一个 *gin.Context")
			assert.Equal(t, tc.wantCode, c.Writer.Status())
			if tc.wantLocation != "" {
				assert.Equal(t, tc.wantLocation, rec.Header().Get("Location"))
			}
		})
	}
}

func TestStreamFileByPath_BadQuery_NoServiceCall(t *testing.T) {
	cases := []struct {
		name   string
		rawURL string
	}{
		{"missing-all", "/files/stream"},
		{"missing-path", "/files/stream?storage_type=local"},
		{"missing-storage-type", "/files/stream?path=img%2Fa.png"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := filemock.NewMockService(t)
			h := NewFileHandler(mockSvc)

			c, rec := newGinCtx(t, tc.rawURL)
			h.StreamFileByPath(c)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, commonModel.INVALID_QUERY_PARAMS, rec.Body.String())
		})
	}
}

func TestStreamFileByPath_DelegatesParsedQuery(t *testing.T) {
	mockSvc := filemock.NewMockService(t)
	var got commonModel.FilePathStreamQueryDto
	var gotCtx *gin.Context
	mockSvc.EXPECT().
		StreamFileByPath(mock.Anything, mock.Anything).
		Run(func(ctx *gin.Context, q commonModel.FilePathStreamQueryDto) {
			got = q
			gotCtx = ctx
			ctx.Status(http.StatusOK)
		}).
		Return().
		Once()

	h := NewFileHandler(mockSvc)
	c, rec := newGinCtx(
		t,
		"/files/stream?storage_type=local&path=img%2Fa.png&name=a.png&content_type=image%2Fpng",
	)

	h.StreamFileByPath(c)

	assert.Equal(t, "local", got.StorageType)
	assert.Equal(t, "img/a.png", got.Path)
	assert.Equal(t, "a.png", got.Name)
	assert.Equal(t, "image/png", got.ContentType)
	assert.Same(t, c, gotCtx)
	assert.Equal(t, http.StatusOK, rec.Code)
}

var errBoom = errors.New("boom")

func TestListFiles(t *testing.T) {
	t.Run("success forwards query and wraps OK", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		var gotQuery commonModel.FileListQueryDto
		mockSvc.EXPECT().
			ListFiles(mock.Anything, mock.Anything).
			Run(func(_ context.Context, q commonModel.FileListQueryDto) { gotQuery = q }).
			Return(commonModel.FileListResultDto{Total: 7}, nil).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.ListFiles(context.Background(), &ListFilesInput{
			Page: 2, PageSize: 20, Search: "kw", StorageType: "s3",
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, int64(7), out.Data.Total)
		assert.Equal(t, commonModel.FileListQueryDto{Page: 2, PageSize: 20, Search: "kw", StorageType: "s3"}, gotQuery)
	})

	t.Run("error is propagated, zero envelope", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			ListFiles(mock.Anything, mock.Anything).
			Return(commonModel.FileListResultDto{}, errBoom).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.ListFiles(context.Background(), &ListFilesInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, FileListOutput{}, out)
	})
}

func TestListFileTree(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		var gotQuery commonModel.FileTreeQueryDto
		mockSvc.EXPECT().
			ListFileTree(mock.Anything, mock.Anything).
			Run(func(_ context.Context, q commonModel.FileTreeQueryDto) { gotQuery = q }).
			Return(commonModel.FileTreeResultDto{}, nil).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.ListFileTree(context.Background(), &ListFileTreeInput{StorageType: "local", Prefix: "img/"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.FileTreeQueryDto{StorageType: "local", Prefix: "img/"}, gotQuery)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			ListFileTree(mock.Anything, mock.Anything).
			Return(commonModel.FileTreeResultDto{}, errBoom).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.ListFileTree(context.Background(), &ListFileTreeInput{StorageType: "local"})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, FileTreeOutput{}, out)
	})
}

func TestGetFileByID(t *testing.T) {
	t.Run("success forwards id", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			GetFileByID(mock.Anything, "f-9").
			Return(commonModel.FileDto{ID: "f-9", Name: "a.png"}, nil).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.GetFileByID(context.Background(), &GetFileByIDInput{ID: "f-9"})

		require.NoError(t, err)
		assert.Equal(t, "f-9", out.Data.ID)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			GetFileByID(mock.Anything, "missing").
			Return(commonModel.FileDto{}, errBoom).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.GetFileByID(context.Background(), &GetFileByIDInput{ID: "missing"})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, FileOutput{}, out)
	})
}

func TestUpdateFileMeta(t *testing.T) {
	t.Run("success forwards id and body", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		var gotID string
		var gotBody commonModel.UpdateFileMetaDto
		mockSvc.EXPECT().
			UpdateFileMeta(mock.Anything, mock.Anything, mock.Anything).
			Run(func(_ context.Context, id string, dto commonModel.UpdateFileMetaDto) {
				gotID = id
				gotBody = dto
			}).
			Return(commonModel.FileDto{ID: "f-1"}, nil).
			Once()

		h := NewFileHandler(mockSvc)
		body := commonModel.UpdateFileMetaDto{Size: 123}
		out, err := h.UpdateFileMeta(context.Background(), &UpdateFileMetaInput{ID: "f-1", Body: body})

		require.NoError(t, err)
		assert.Equal(t, "f-1", gotID)
		assert.Equal(t, body, gotBody)
		assert.Equal(t, "f-1", out.Data.ID)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			UpdateFileMeta(mock.Anything, mock.Anything, mock.Anything).
			Return(commonModel.FileDto{}, errBoom).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.UpdateFileMeta(context.Background(), &UpdateFileMetaInput{ID: "f-1"})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, FileOutput{}, out)
	})
}

func TestCreateExternalFile(t *testing.T) {
	t.Run("success uses UPLOAD_SUCCESS message", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			CreateExternalFile(mock.Anything, mock.Anything).
			Return(commonModel.FileDto{ID: "ext-1"}, nil).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.CreateExternalFile(context.Background(), &CreateExternalFileInput{
			Body: commonModel.CreateExternalFileDto{URL: "https://x/y.png"},
		})

		require.NoError(t, err)
		assert.Equal(t, "ext-1", out.Data.ID)
		assert.Equal(t, commonModel.UPLOAD_SUCCESS, out.Message)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			CreateExternalFile(mock.Anything, mock.Anything).
			Return(commonModel.FileDto{}, errBoom).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.CreateExternalFile(context.Background(), &CreateExternalFileInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, FileOutput{}, out)
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("success uses DELETE_SUCCESS message", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().DeleteFile(mock.Anything, "f-1").Return(nil).Once()

		h := NewFileHandler(mockSvc)
		out, err := h.DeleteFile(context.Background(), &DeleteFileInput{ID: "f-1"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DELETE_SUCCESS, out.Message)
		assert.Nil(t, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().DeleteFile(mock.Anything, "f-1").Return(errBoom).Once()

		h := NewFileHandler(mockSvc)
		out, err := h.DeleteFile(context.Background(), &DeleteFileInput{ID: "f-1"})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, EmptyOutput{}, out)
	})
}

func TestGetFilePresignURL(t *testing.T) {
	t.Run("success uses presign message", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			GetFilePresignURL(mock.Anything, mock.Anything).
			Return(commonModel.PresignDto{ID: "p-1", PresignURL: "https://x/put"}, nil).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.GetFilePresignURL(context.Background(), &GetFilePresignURLInput{
			Body: commonModel.GetPresignURLDto{FileName: "a.png"},
		})

		require.NoError(t, err)
		assert.Equal(t, "p-1", out.Data.ID)
		assert.Equal(t, commonModel.GET_S3_PRESIGN_URL_SUCCESS, out.Message)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := filemock.NewMockService(t)
		mockSvc.EXPECT().
			GetFilePresignURL(mock.Anything, mock.Anything).
			Return(commonModel.PresignDto{}, errBoom).
			Once()

		h := NewFileHandler(mockSvc)
		out, err := h.GetFilePresignURL(context.Background(), &GetFilePresignURLInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, PresignOutput{}, out)
	})
}
