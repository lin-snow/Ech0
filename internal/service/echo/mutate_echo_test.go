// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/event"
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

func TestDeleteEchoById_Guards(t *testing.T) {
	t.Run("non-admin is denied", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		require.EqualError(t, svc.DeleteEchoById(helpers.CtxAsUser(userID), echoID), commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("user lookup error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		boom := errors.New("lookup failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), boom).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		require.ErrorIs(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID), echoID), boom)
	})
}

func TestDeleteEchoById_NotFound(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, nil).Once()

	svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
	require.EqualError(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID), echoID), commonModel.ECHO_NOT_FOUND)
}

func TestDeleteEchoById_GetEchoError(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	boom := errors.New("fetch failed")
	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, boom).Once()

	svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
	require.ErrorIs(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID), echoID), boom)
}

func TestDeleteEchoById_Success(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	bus := helpers.NewTestBus(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()

	stored := helpers.NewEcho(func(e *echoModel.Echo) {
		e.ID = echoID
		e.EchoFiles = []fileModel.EchoFile{
			{File: fileModel.File{ID: "f-local", Key: "k-local", StorageType: "local"}},
			{File: fileModel.File{ID: "f-ext", Key: "k-ext", StorageType: "external"}},
		}
	})
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&stored, nil).Once()
	file.EXPECT().DeleteFileRecord(mock.Anything, "f-local").Return(nil).Once()
	file.EXPECT().DeleteFileRecord(mock.Anything, "f-ext").Return(nil).Once()
	repo.EXPECT().DeleteEchoById(mock.Anything, echoID).Return(nil).Once()
	repo.EXPECT().InvalidateEchoCaches(echoID).Once()
	file.EXPECT().DeleteStoredFile("local", "k-local").Return(errors.New("ignored")).Once()

	var got event.EchoDeleted
	var fired int
	unsub, err := bus.Subscribe(func(_ context.Context, e busen.Event[event.EchoDeleted]) error {
		got = e.Value
		fired++
		return nil
	})
	require.NoError(t, err)
	defer unsub()

	svc := echoService.NewEchoService(tx, common, file, repo, func() *busen.Bus { return bus })
	require.NoError(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID), echoID))

	require.Equal(t, 1, fired)
	assert.Equal(t, echoID, got.Echo.ID)
	assert.True(t, got.User.IsAdmin)
}

func TestDeleteEchoById_DeleteFileRecordError(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	boom := errors.New("delete record failed")

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	stored := helpers.NewEcho(func(e *echoModel.Echo) {
		e.ID = echoID
		e.EchoFiles = []fileModel.EchoFile{{File: fileModel.File{ID: "f-1", Key: "k-1", StorageType: "local"}}}
	})
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&stored, nil).Once()
	file.EXPECT().DeleteFileRecord(mock.Anything, "f-1").Return(boom).Once()

	svc := echoService.NewEchoService(tx, common, file, repo, nilBus)
	require.ErrorIs(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID), echoID), boom)
}

func TestUpdateEcho_Guards(t *testing.T) {
	t.Run("non-admin is denied", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		require.EqualError(t,
			svc.UpdateEcho(helpers.CtxAsUser(userID), &echoModel.Echo{ID: echoID, Content: "hi"}),
			commonModel.NO_PERMISSION_DENIED,
		)
	})

	t.Run("invalid extension rejected before transaction", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.UpdateEcho(helpers.CtxAsUser(adminID), &echoModel.Echo{
			ID:        echoID,
			Content:   "hi",
			Extension: &echoModel.EchoExtension{Type: echoModel.Extension_MUSIC},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "extension payload")
	})

	t.Run("empty echo rejected", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		require.EqualError(t,
			svc.UpdateEcho(helpers.CtxAsUser(adminID), &echoModel.Echo{ID: echoID, Content: "   "}),
			commonModel.ECHO_CAN_NOT_BE_EMPTY,
		)
	})
}

func TestUpdateEcho_Success(t *testing.T) {
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
		GetFilesByIDs(mock.Anything, []string{"file-1"}).
		Return([]commonModel.FileDto{{ID: "file-1", Category: "image"}}, nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()

	var updated echoModel.Echo
	repo.EXPECT().
		UpdateEcho(mock.Anything, mock.Anything).
		Run(func(_ context.Context, e *echoModel.Echo) { updated = *e }).
		Return(nil).
		Once()
	repo.EXPECT().InvalidateEchoCaches(echoID).Once()
	file.EXPECT().ConfirmTempFiles(mock.Anything, []string{"file-1"}).Return(nil).Once()

	var got event.EchoUpdated
	var fired int
	unsub, err := bus.Subscribe(func(_ context.Context, e busen.Event[event.EchoUpdated]) error {
		got = e.Value
		fired++
		return nil
	})
	require.NoError(t, err)
	defer unsub()

	in := &echoModel.Echo{
		ID:        echoID,
		Content:   "updated",
		Layout:    "bogus-layout",
		EchoFiles: []fileModel.EchoFile{{FileID: "file-1"}},
	}
	svc := echoService.NewEchoService(tx, common, file, repo, func() *busen.Bus { return bus })
	require.NoError(t, svc.UpdateEcho(helpers.CtxAsUser(adminID), in))

	assert.Equal(t, echoModel.LayoutWaterfall, updated.Layout)
	require.Len(t, updated.EchoFiles, 1)
	assert.Equal(t, echoID, updated.EchoFiles[0].EchoID)
	require.Equal(t, 1, fired)
	assert.Equal(t, echoID, got.Echo.ID)
}

func TestUpdateEcho_TransactionError(t *testing.T) {
	repo := echomock.NewMockRepository(t)
	common := commonmock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	boom := errors.New("update failed")

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()
	repo.EXPECT().UpdateEcho(mock.Anything, mock.Anything).Return(boom).Once()

	svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
	require.ErrorIs(t,
		svc.UpdateEcho(helpers.CtxAsUser(adminID), &echoModel.Echo{ID: echoID, Content: "x", Layout: echoModel.LayoutGrid}),
		boom,
	)
}
