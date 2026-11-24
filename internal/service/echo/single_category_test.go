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
	filemock "github.com/lin-snow/ech0/internal/test/mocks/filemock"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func echoWithFileIDs(ids ...string) *echoModel.Echo {
	files := make([]echoModel.EchoFile, 0, len(ids))
	for i, id := range ids {
		files = append(files, echoModel.EchoFile{FileID: id, SortOrder: i})
	}
	e := helpers.NewEcho(func(e *echoModel.Echo) {
		e.ID = echoID
		e.Content = "hi"
		e.EchoFiles = files
	})
	return &e
}

func fileDto(id, category string) fileModel.FileDTO {
	return fileModel.FileDTO{ID: id, Category: category}
}

func expectFiles(file *filemock.MockService, files []fileModel.FileDTO) {
	for _, f := range files {
		file.EXPECT().Get(mock.Anything, f.ID).Return(f, nil).Once()
	}
}

func TestValidateSingleFileCategory_Reject(t *testing.T) {
	cases := []struct {
		name  string
		files []fileModel.FileDTO
	}{
		{
			name:  "image mixed with audio rejected",
			files: []fileModel.FileDTO{fileDto("f1", fileModel.CategoryImage), fileDto("f2", fileModel.CategoryAudio)},
		},
		{
			name:  "image mixed with video rejected",
			files: []fileModel.FileDTO{fileDto("f1", fileModel.CategoryImage), fileDto("f2", fileModel.CategoryVideo)},
		},
		{
			name:  "two audios rejected",
			files: []fileModel.FileDTO{fileDto("f1", fileModel.CategoryAudio), fileDto("f2", fileModel.CategoryAudio)},
		},
		{
			name:  "two videos rejected",
			files: []fileModel.FileDTO{fileDto("f1", fileModel.CategoryVideo), fileDto("f2", fileModel.CategoryVideo)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newEchoRepoMock(t)
			common := commonmock.NewMockService(t)
			file := filemock.NewMockService(t)
			common.EXPECT().
				CommonGetUserByUserId(mock.Anything, adminID).
				Return(helpers.NewUser(helpers.AsAdmin), nil).
				Once()
			expectFiles(file, tc.files)

			svc := echoService.NewEchoService(nil, common, file, repo, nilBus)
			ids := make([]string, len(tc.files))
			for i, f := range tc.files {
				ids[i] = f.ID
			}
			_, err := svc.PostEcho(helpers.CtxAsUser(adminID), echoWithFileIDs(ids...))

			require.EqualError(t, err, commonModel.ECHO_MIXED_FILE_CATEGORIES)
		})
	}
}

func TestValidateSingleFileCategory_MultipleImagesAllowed(t *testing.T) {
	repo := newEchoRepoMock(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	bus := helpers.NewTestBus(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	expectFiles(file, []fileModel.FileDTO{
		fileDto("f1", fileModel.CategoryImage),
		fileDto("f2", fileModel.CategoryImage),
	})

	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()

	var created echoModel.Echo
	repo.EXPECT().
		CreateEcho(mock.Anything, mock.Anything).
		Run(func(_ context.Context, e *echoModel.Echo) { e.ID = "saved-1"; created = *e }).
		Return(nil).
		Once()
	repo.EXPECT().InvalidateEchoCaches().Once()
	saved := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = "saved-1" })
	repo.EXPECT().GetEchosById(mock.Anything, mock.Anything).Return(&saved, nil).Once()
	file.EXPECT().
		Replace(mock.Anything, []string{"f1", "f2"}, fileModel.OwnerKindEcho, "saved-1", fileModel.RoleAttachment).
		Return(nil).
		Once()

	svc := echoService.NewEchoService(tx, common, file, repo, func() *busen.Bus { return bus })
	_, err := svc.PostEcho(helpers.CtxAsUser(adminID), echoWithFileIDs("f1", "f2"))
	require.NoError(t, err)

	assert.Len(t, created.EchoFiles, 2)
}

func TestUpdateEcho_MixedFileCategoriesRejected(t *testing.T) {
	repo := newEchoRepoMock(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	existing := echoWithFileIDs("f1", "f2")
	existing.UserID = adminID
	repo.EXPECT().GetEchosById(mock.Anything, mock.Anything).Return(existing, nil).Once()
	expectFiles(file, []fileModel.FileDTO{
		fileDto("f1", fileModel.CategoryImage),
		fileDto("f2", fileModel.CategoryAudio),
	})

	svc := echoService.NewEchoService(nil, common, file, repo, nilBus)
	_, err := svc.UpdateEcho(helpers.CtxAsUser(adminID), echoWithFileIDs("f1", "f2"))

	require.EqualError(t, err, commonModel.ECHO_MIXED_FILE_CATEGORIES)
}
