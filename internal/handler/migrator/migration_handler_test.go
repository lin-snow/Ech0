// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	migratormock "github.com/lin-snow/ech0/internal/test/mocks/migratormock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func TestStartMigration(t *testing.T) {
	t.Run("success forwards body and returns the state", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		var gotReq migratorModel.StartGlobalMigrationRequest
		mockSvc.EXPECT().
			StartGlobalMigration(mock.Anything, mock.Anything).
			Run(func(_ context.Context, req migratorModel.StartGlobalMigrationRequest) { gotReq = req }).
			Return(migratorModel.GlobalMigrationStateDTO{Status: "running", SourceType: "ech0"}, nil).
			Once()

		h := NewMigrationHandler(mockSvc)
		req := migratorModel.StartGlobalMigrationRequest{SourceType: "ech0", SourcePayload: map[string]any{"tmp": "x"}}
		out, err := h.StartMigration(context.Background(), &StartMigrationInput{Body: req})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "running", out.Body.Status)
		assert.Equal(t, "ech0", out.Body.SourceType)
		assert.Equal(t, req, gotReq, "body 应原样透传给 service")
	})

	t.Run("error is propagated with a nil output", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			StartGlobalMigration(mock.Anything, mock.Anything).
			Return(migratorModel.GlobalMigrationStateDTO{}, errBoom).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.StartMigration(context.Background(), &StartMigrationInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Nil(t, out)
	})
}

func TestGetMigrationStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			GetGlobalMigrationStatus(mock.Anything).
			Return(migratorModel.GlobalMigrationStateDTO{Status: "idle"}, nil).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.GetMigrationStatus(context.Background(), &GetMigrationStatusInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "idle", out.Body.Status)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			GetGlobalMigrationStatus(mock.Anything).
			Return(migratorModel.GlobalMigrationStateDTO{}, errBoom).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.GetMigrationStatus(context.Background(), &GetMigrationStatusInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Nil(t, out)
	})
}

func TestCancelMigration(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			CancelGlobalMigration(mock.Anything).
			Return(migratorModel.GlobalMigrationStateDTO{Status: "cancelled"}, nil).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CancelMigration(context.Background(), &CancelMigrationInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "cancelled", out.Body.Status)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			CancelGlobalMigration(mock.Anything).
			Return(migratorModel.GlobalMigrationStateDTO{}, errBoom).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CancelMigration(context.Background(), &CancelMigrationInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Nil(t, out)
	})
}

func TestCleanupMigration(t *testing.T) {
	t.Run("success returns an empty output", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().CleanupGlobalMigration(mock.Anything).Return(nil).Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CleanupMigration(context.Background(), &CleanupMigrationInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().CleanupGlobalMigration(mock.Anything).Return(errBoom).Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CleanupMigration(context.Background(), &CleanupMigrationInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Nil(t, out)
	})
}

func TestStartExport(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			StartExport(mock.Anything, mock.Anything).
			Return(migratorModel.ExportStateDTO{Status: "running"}, nil).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.StartExport(context.Background(), &StartExportInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "running", out.Body.Status)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			StartExport(mock.Anything, mock.Anything).
			Return(migratorModel.ExportStateDTO{}, errBoom).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.StartExport(context.Background(), &StartExportInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Nil(t, out)
	})
}

func TestGetExportStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			GetExportStatus(mock.Anything).
			Return(migratorModel.ExportStateDTO{Status: "succeeded", FileName: "snap.zip", Size: 42}, nil).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.GetExportStatus(context.Background(), &GetExportStatusInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "succeeded", out.Body.Status)
		assert.Equal(t, "snap.zip", out.Body.FileName)
		assert.Equal(t, int64(42), out.Body.Size)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			GetExportStatus(mock.Anything).
			Return(migratorModel.ExportStateDTO{}, errBoom).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.GetExportStatus(context.Background(), &GetExportStatusInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Nil(t, out)
	})
}

func TestCancelExport(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			CancelExport(mock.Anything).
			Return(migratorModel.ExportStateDTO{Status: "cancelled"}, nil).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CancelExport(context.Background(), &CancelExportInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, "cancelled", out.Body.Status)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			CancelExport(mock.Anything).
			Return(migratorModel.ExportStateDTO{}, errBoom).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CancelExport(context.Background(), &CancelExportInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Nil(t, out)
	})
}

type formBody struct {
	contentType string
	payload     string
}

// multipartBody builds a one-field multipart request with a source zip attached.
func multipartBody(t *testing.T, field, value string) formBody {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	require.NoError(t, w.WriteField(field, value))
	fw, err := w.CreateFormFile("file", "src.zip")
	require.NoError(t, err)
	_, err = fw.Write([]byte("PK-src"))
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return formBody{contentType: w.FormDataContentType(), payload: buf.String()}
}

// newGinCtx builds a request context good enough for the raw routes.
func newGinCtx(method, target, contentType, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		c.Request.Header.Set("Content-Type", contentType)
	}
	return c, rec
}

func TestUploadSourceZipRejectsMissingFields(t *testing.T) {
	mockSvc := migratormock.NewMockService(t)
	h := NewMigrationHandler(mockSvc)

	c, rec := newGinCtx(http.MethodPost, "/migration/upload", "application/x-www-form-urlencoded", "")
	h.UploadSourceZip()(c)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	var problem struct {
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
	assert.Equal(t, http.StatusBadRequest, problem.Status)
	assert.NotEmpty(t, problem.Detail)
}

func TestUploadSourceZipReturnsTheSource(t *testing.T) {
	mockSvc := migratormock.NewMockService(t)
	want := migratorModel.UploadMigrationSourceZipResponse{}
	mockSvc.EXPECT().
		UploadSourceZip(mock.Anything, "ech0", mock.Anything).
		Return(want, nil).
		Once()
	h := NewMigrationHandler(mockSvc)

	body := multipartBody(t, "source_type", "ech0")
	c, rec := newGinCtx(http.MethodPost, "/migration/upload", body.contentType, body.payload)
	h.UploadSourceZip()(c)

	require.Equal(t, http.StatusOK, rec.Code)
	var got migratorModel.UploadMigrationSourceZipResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, want, got)
}

func TestDownloadExportMapsErrorsToProblemDocuments(t *testing.T) {
	mockSvc := migratormock.NewMockService(t)
	mockSvc.EXPECT().
		DownloadExport(mock.Anything, mock.Anything, "").
		Return(errBoom).
		Once()
	h := NewMigrationHandler(mockSvc)

	c, rec := newGinCtx(http.MethodGet, "/migration/export/download", "", "")
	h.DownloadExport()(c)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	var problem struct {
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem))
	assert.Equal(t, http.StatusInternalServerError, problem.Status)
	assert.NotContains(t, rec.Body.String(), "boom")
	assert.Equal(t, "internal server error", problem.Detail)
}

func TestDownloadExportKeepsTheStreamedBody(t *testing.T) {
	mockSvc := migratormock.NewMockService(t)
	mockSvc.EXPECT().
		DownloadExport(mock.Anything, mock.Anything, "capsule").
		Run(func(ctx *gin.Context, _ context.Context, _ string) {
			ctx.Writer.Header().Set("Content-Type", "application/zip")
			ctx.Writer.WriteHeader(http.StatusOK)
			_, _ = ctx.Writer.WriteString("PK-bytes")
		}).
		Return(nil).
		Once()
	h := NewMigrationHandler(mockSvc)

	c, rec := newGinCtx(http.MethodGet, "/migration/export/download?format=capsule", "", "")
	h.DownloadExport()(c)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/zip", rec.Header().Get("Content-Type"))
	assert.Equal(t, "PK-bytes", rec.Body.String())
}
