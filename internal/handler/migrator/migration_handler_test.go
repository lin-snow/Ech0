// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	migratormock "github.com/lin-snow/ech0/internal/test/mocks/migratormock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func TestStartMigration(t *testing.T) {
	t.Run("success forwards body and wraps OK", func(t *testing.T) {
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
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, "running", out.Data.Status)
		assert.Equal(t, "ech0", out.Data.SourceType)
		assert.Equal(t, req, gotReq, "body 应原样透传给 service")
	})

	t.Run("error is propagated, zero envelope", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().
			StartGlobalMigration(mock.Anything, mock.Anything).
			Return(migratorModel.GlobalMigrationStateDTO{}, errBoom).
			Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.StartMigration(context.Background(), &StartMigrationInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, GlobalMigrationOutput{}, out)
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
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, "idle", out.Data.Status)
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
		assert.Equal(t, GlobalMigrationOutput{}, out)
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
		assert.Equal(t, "cancelled", out.Data.Status)
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
		assert.Equal(t, GlobalMigrationOutput{}, out)
	})
}

func TestCleanupMigration(t *testing.T) {
	t.Run("success returns empty OK", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().CleanupGlobalMigration(mock.Anything).Return(nil).Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CleanupMigration(context.Background(), &CleanupMigrationInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Nil(t, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		mockSvc := migratormock.NewMockService(t)
		mockSvc.EXPECT().CleanupGlobalMigration(mock.Anything).Return(errBoom).Once()

		h := NewMigrationHandler(mockSvc)
		out, err := h.CleanupMigration(context.Background(), &CleanupMigrationInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, EmptyOutput{}, out)
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
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, "running", out.Data.Status)
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
		assert.Equal(t, ExportOutput{}, out)
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
		assert.Equal(t, "succeeded", out.Data.Status)
		assert.Equal(t, "snap.zip", out.Data.FileName)
		assert.Equal(t, int64(42), out.Data.Size)
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
		assert.Equal(t, ExportOutput{}, out)
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
		assert.Equal(t, "cancelled", out.Data.Status)
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
		assert.Equal(t, ExportOutput{}, out)
	})
}
