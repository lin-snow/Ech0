// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"testing"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Two echoes can point at one external file row because CreateExternalFile
// dedupes by URL. Deleting one echo must keep the row the other still uses.
func TestDeleteEchoKeepsSharedExternalFile(t *testing.T) {
	db := helpers.NewTestDB(t)
	echoRepo := echoRepository.NewEchoRepository(func() *gorm.DB { return db }, helpers.NewTestCache())
	fileRepo := fileRepository.NewFileRepository(func() *gorm.DB { return db })
	tx := transaction.NewGormTransactor(func() *gorm.DB { return db })
	bus := helpers.NewTestBus(t)
	fileSvc := fileService.NewFileService(tx, commonmock.NewMockCommonRepository(t), fileRepo, helpers.NewTestStorage(t), func() *busen.Bus { return bus })

	common := commonmock.NewMockService(t)
	common.EXPECT().CommonGetUserByUserId(mock.Anything, "user-test-0001").Return(helpers.NewUser(), nil).Once()
	svc := echoService.NewEchoService(tx, common, fileSvc, echoRepo, func() *busen.Bus { return bus })

	const userID = "user-test-0001"
	file := &fileModel.File{
		ID: "file-shared", Key: "external/image/shared", StorageType: "external",
		Provider: "external", URL: "https://example.com/pic.png", Name: "pic.png",
		ContentType: "image/png", Category: "image", UserID: userID,
	}
	require.NoError(t, db.Create(file).Error)

	for _, id := range []string{"echo-a", "echo-b"} {
		e := &echoModel.Echo{ID: id, Content: "shared image", UserID: userID, Layout: "waterfall"}
		require.NoError(t, db.Create(e).Error)
		require.NoError(t, db.Create(&fileModel.EchoFile{EchoID: id, FileID: file.ID}).Error)
	}

	require.NoError(t, svc.DeleteEchoById(helpers.CtxAsUser(userID), "echo-a"))

	var files int64
	require.NoError(t, db.Model(&fileModel.File{}).Where("id = ?", file.ID).Count(&files).Error)
	require.Equal(t, int64(1), files, "shared file row should survive")

	other, err := echoRepo.GetEchosById(context.Background(), "echo-b")
	require.NoError(t, err)
	require.NotNil(t, other)
	require.Len(t, other.EchoFiles, 1, "echo-b lost its attachment")
	require.Equal(t, file.ID, other.EchoFiles[0].File.ID)
}
