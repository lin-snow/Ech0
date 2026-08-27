// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	echomock "github.com/lin-snow/ech0/internal/test/mocks/echomock"
	filemock "github.com/lin-snow/ech0/internal/test/mocks/filemock"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func echoWithFileIDs(ids ...string) *echoModel.Echo {
	files := make([]fileModel.EchoFile, 0, len(ids))
	for i, id := range ids {
		files = append(files, fileModel.EchoFile{FileID: id, SortOrder: i})
	}
	e := helpers.NewEcho(func(e *echoModel.Echo) {
		e.ID = echoID
		e.Content = "hi"
		e.EchoFiles = files
	})
	return &e
}

func fileDto(id, category string) commonModel.FileDto {
	return commonModel.FileDto{ID: id, Category: category}
}

func TestValidateSingleFileCategory_Reject(t *testing.T) {
	cases := []struct {
		name  string
		files []commonModel.FileDto
	}{
		{
			name:  "image mixed with audio rejected",
			files: []commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "audio")},
		},
		{
			name:  "image mixed with video rejected",
			files: []commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "video")},
		},
		{
			name:  "two audios rejected",
			files: []commonModel.FileDto{fileDto("f1", "audio"), fileDto("f2", "audio")},
		},
		{
			name:  "two videos rejected",
			files: []commonModel.FileDto{fileDto("f1", "video"), fileDto("f2", "video")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := echomock.NewMockRepository(t)
			common := commonmock.NewMockService(t)
			file := filemock.NewMockService(t)
			common.EXPECT().
				CommonGetUserByUserId(mock.Anything, adminID).
				Return(helpers.NewUser(helpers.AsAdmin), nil).
				Once()
			file.EXPECT().
				GetFilesByIDs(mock.Anything, mock.Anything).
				Return(tc.files, nil).
				Once()

			svc := echoService.NewEchoService(nil, common, file, repo, nilBus)
			ids := make([]string, len(tc.files))
			for i, f := range tc.files {
				ids[i] = f.ID
			}
			err := svc.PostEcho(helpers.CtxAsUser(adminID), echoWithFileIDs(ids...))

			require.EqualError(t, err, commonModel.ECHO_MIXED_FILE_CATEGORIES)
		})
	}
}

func TestValidateSingleFileCategory_MultipleImagesAllowed(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	bus := helpers.NewTestBus(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	file.EXPECT().
		GetFilesByIDs(mock.Anything, mock.Anything).
		Return([]commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "image")}, nil).
		Once()

	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()

	var created echoModel.Echo
	repo.EXPECT().
		CreateEcho(mock.Anything, mock.Anything).
		Run(func(_ context.Context, e *echoModel.Echo) { created = *e }).
		Return(nil).
		Once()
	repo.EXPECT().InvalidateEchoCaches().Once()
	saved := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = "saved-1" })
	repo.EXPECT().GetEchosById(mock.Anything, mock.Anything).Return(&saved, nil).Once()
	file.EXPECT().ConfirmTempFiles(mock.Anything, mock.Anything).Return(nil).Once()

	svc := echoService.NewEchoService(tx, common, file, repo, func() *busen.Bus { return bus })
	require.NoError(t, svc.PostEcho(helpers.CtxAsUser(adminID), echoWithFileIDs("f1", "f2")))

	assert.Len(t, created.EchoFiles, 2)
}

func TestUpdateEcho_MixedFileCategoriesRejected(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	file.EXPECT().
		GetFilesByIDs(mock.Anything, mock.Anything).
		Return([]commonModel.FileDto{fileDto("f1", "image"), fileDto("f2", "audio")}, nil).
		Once()

	svc := echoService.NewEchoService(nil, common, file, repo, nilBus)
	err := svc.UpdateEcho(helpers.CtxAsUser(adminID), echoWithFileIDs("f1", "f2"))

	require.EqualError(t, err, commonModel.ECHO_MIXED_FILE_CATEGORIES)
}
