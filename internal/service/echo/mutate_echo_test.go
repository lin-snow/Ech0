// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/event"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	filemock "github.com/lin-snow/ech0/internal/test/mocks/filemock"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteEchoById_Guards(t *testing.T) {
	t.Run("member cannot delete another user's echo", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		foreign := helpers.NewEcho(func(e *echoModel.Echo) { e.UserID = adminID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&foreign, nil).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		require.ErrorIs(t, svc.DeleteEchoById(helpers.CtxAsUser(userID), echoID), authz.ErrDenied)
	})

	t.Run("member deletes their own echo", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		file := filemock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()
		own := helpers.NewEcho(func(e *echoModel.Echo) { e.UserID = userID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&own, nil).Once()
		file.EXPECT().ClearOwner(mock.Anything, fileModel.OwnerKindEcho, echoID).Return(nil).Once()
		repo.EXPECT().DeleteEchoById(mock.Anything, echoID).Return(nil).Once()
		repo.EXPECT().InvalidateEchoCaches(echoID).Return().Once()

		svc := echoService.NewEchoService(tx, common, file, repo, nilBus)
		require.NoError(t, svc.DeleteEchoById(helpers.CtxAsUser(userID), echoID))
	})

	t.Run("user lookup error propagates", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		boom := errors.New("lookup failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), boom).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		require.ErrorIs(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID, role.Admin), echoID), boom)
	})
}

func TestDeleteEchoById_NotFound(t *testing.T) {
	repo := newEchoRepoMock(t)
	common := commonmock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, nil).Once()

	svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
	require.EqualError(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID, role.Admin), echoID), apperr.ErrEchoNotFound.Error())
}

func TestDeleteEchoById_GetEchoError(t *testing.T) {
	repo := newEchoRepoMock(t)
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
	require.ErrorIs(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID, role.Admin), echoID), boom)
}

func TestDeleteEchoById_Success(t *testing.T) {
	repo := newEchoRepoMock(t)
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
		e.EchoFiles = []echoModel.EchoFile{
			{File: fileModel.File{ID: "f-local"}},
			{File: fileModel.File{ID: "f-ext"}},
		}
	})
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&stored, nil).Once()
	file.EXPECT().ClearOwner(mock.Anything, fileModel.OwnerKindEcho, echoID).Return(nil).Once()
	repo.EXPECT().DeleteEchoById(mock.Anything, echoID).Return(nil).Once()
	repo.EXPECT().InvalidateEchoCaches(echoID).Once()

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
	require.NoError(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID, role.Admin), echoID))

	require.Equal(t, 1, fired)
	assert.Equal(t, echoID, got.Echo.ID)
	assert.Equal(t, userModel.RoleAdmin, got.User.Role)
}

func TestDeleteEchoById_ClearOwnerError(t *testing.T) {
	repo := newEchoRepoMock(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	boom := errors.New("clear links failed")

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	stored := helpers.NewEcho(func(e *echoModel.Echo) {
		e.ID = echoID
		e.EchoFiles = []echoModel.EchoFile{{File: fileModel.File{ID: "f-1"}}}
	})
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&stored, nil).Once()
	file.EXPECT().ClearOwner(mock.Anything, fileModel.OwnerKindEcho, echoID).Return(boom).Once()

	svc := echoService.NewEchoService(tx, common, file, repo, nilBus)
	require.ErrorIs(t, svc.DeleteEchoById(helpers.CtxAsUser(adminID, role.Admin), echoID), boom)
}

func TestUpdateEcho_Guards(t *testing.T) {
	t.Run("member cannot edit another user's echo", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()
		foreign := helpers.NewEcho(func(e *echoModel.Echo) { e.UserID = adminID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&foreign, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		_, err := svc.UpdateEcho(helpers.CtxAsUser(userID), &echoModel.Echo{ID: echoID, Content: "hi"})
		require.ErrorIs(t, err, authz.ErrDenied)
	})

	t.Run("invalid extension rejected before transaction", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		existing := helpers.NewEcho(func(e *echoModel.Echo) { e.UserID = adminID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&existing, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		_, err := svc.UpdateEcho(helpers.CtxAsUser(adminID, role.Admin), &echoModel.Echo{
			ID:        echoID,
			Content:   "hi",
			Extension: &echoModel.EchoExtension{Type: echoModel.Extension_MUSIC},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "extension payload")
	})

	t.Run("attachment replacement failure surfaces from the transaction", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		file := filemock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		existing := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = echoID; e.UserID = adminID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&existing, nil).Once()
		file.EXPECT().
			Get(mock.Anything, "file-1").
			Return(fileModel.FileDTO{ID: "file-1", Category: fileModel.CategoryImage}, nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()
		repo.EXPECT().UpdateEcho(mock.Anything, mock.Anything).Return(nil).Once()
		boom := errors.New("replace failed")
		file.EXPECT().
			Replace(mock.Anything, []string{"file-1"}, fileModel.OwnerKindEcho, echoID, fileModel.RoleAttachment).
			Return(boom).
			Once()

		svc := echoService.NewEchoService(tx, common, file, repo, nilBus)
		_, err := svc.UpdateEcho(helpers.CtxAsUser(adminID, role.Admin), &echoModel.Echo{
			ID:        echoID,
			Content:   "updated",
			EchoFiles: []echoModel.EchoFile{{FileID: "file-1"}},
		})
		require.ErrorIs(t, err, boom)
	})

	t.Run("empty echo rejected", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		existing := helpers.NewEcho(func(e *echoModel.Echo) { e.UserID = adminID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&existing, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		_, err := svc.UpdateEcho(helpers.CtxAsUser(adminID, role.Admin), &echoModel.Echo{ID: echoID, Content: "   "})
		require.EqualError(t, err, apperr.ErrEchoEmpty.Error())
	})
}

func TestUpdateEcho_Success(t *testing.T) {
	repo := newEchoRepoMock(t)
	common := commonmock.NewMockService(t)
	file := filemock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	bus := helpers.NewTestBus(t)

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	existing := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = echoID; e.UserID = adminID })
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&existing, nil).Once()
	file.EXPECT().
		Get(mock.Anything, "file-1").
		Return(fileModel.FileDTO{ID: "file-1", Category: fileModel.CategoryImage}, nil).
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
	saved := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = echoID })
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&saved, nil).Once()
	file.EXPECT().
		Replace(mock.Anything, []string{"file-1"}, fileModel.OwnerKindEcho, echoID, fileModel.RoleAttachment).
		Return(nil).
		Once()

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
		EchoFiles: []echoModel.EchoFile{{FileID: "file-1"}},
	}
	svc := echoService.NewEchoService(tx, common, file, repo, func() *busen.Bus { return bus })
	_, err = svc.UpdateEcho(helpers.CtxAsUser(adminID, role.Admin), in)
	require.NoError(t, err)

	assert.Equal(t, echoModel.LayoutWaterfall, updated.Layout)
	require.Len(t, updated.EchoFiles, 1)
	assert.Equal(t, "file-1", updated.EchoFiles[0].FileID)
	require.Equal(t, 1, fired)
	assert.Equal(t, echoID, got.Echo.ID)
}

func TestUpdateEcho_EchoFilesSemantics(t *testing.T) {
	t.Run("omitted echo_files keeps existing attachments", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		file := filemock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)

		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		existing := helpers.NewEcho(func(e *echoModel.Echo) {
			e.ID = echoID
			e.UserID = adminID
			e.EchoFiles = []echoModel.EchoFile{{FileID: "file-1", SortOrder: 2}}
		})
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&existing, nil).Once()
		file.EXPECT().
			Get(mock.Anything, "file-1").
			Return(fileModel.FileDTO{ID: "file-1", Category: fileModel.CategoryImage}, nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()

		var updated echoModel.Echo
		repo.EXPECT().UpdateEcho(mock.Anything, mock.Anything).
			Run(func(_ context.Context, e *echoModel.Echo) { updated = *e }).
			Return(nil).
			Once()
		repo.EXPECT().InvalidateEchoCaches(echoID).Once()
		saved := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = echoID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&saved, nil).Once()

		svc := echoService.NewEchoService(tx, common, file, repo, nilBus)
		_, err := svc.UpdateEcho(helpers.CtxAsUser(adminID, role.Admin), &echoModel.Echo{
			ID:      echoID,
			Content: "new body",
		})
		require.NoError(t, err)
		require.Len(t, updated.EchoFiles, 1)
		assert.Equal(t, "file-1", updated.EchoFiles[0].FileID)
		assert.Equal(t, 2, updated.EchoFiles[0].SortOrder)
	})

	t.Run("explicit empty echo_files clears attachments", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		common := commonmock.NewMockService(t)
		file := filemock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)

		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		existing := helpers.NewEcho(func(e *echoModel.Echo) {
			e.ID = echoID
			e.UserID = adminID
			e.EchoFiles = []echoModel.EchoFile{{FileID: "file-1"}}
		})
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&existing, nil).Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()
		file.EXPECT().
			Replace(mock.Anything, []string{}, fileModel.OwnerKindEcho, echoID, fileModel.RoleAttachment).
			Return(nil).
			Once()

		var updated echoModel.Echo
		repo.EXPECT().UpdateEcho(mock.Anything, mock.Anything).
			Run(func(_ context.Context, e *echoModel.Echo) { updated = *e }).
			Return(nil).
			Once()
		repo.EXPECT().InvalidateEchoCaches(echoID).Once()
		saved := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = echoID })
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&saved, nil).Once()

		empty := []echoModel.EchoFile{}
		svc := echoService.NewEchoService(tx, common, file, repo, nilBus)
		_, err := svc.UpdateEcho(helpers.CtxAsUser(adminID, role.Admin), &echoModel.Echo{
			ID:        echoID,
			Content:   "new body",
			EchoFiles: empty,
		})
		require.NoError(t, err)
		assert.Empty(t, updated.EchoFiles)
	})
}

func TestUpdateEcho_TransactionError(t *testing.T) {
	repo := newEchoRepoMock(t)
	common := commonmock.NewMockService(t)
	tx := txmock.NewMockTransactor(t)
	boom := errors.New("update failed")

	common.EXPECT().
		CommonGetUserByUserId(mock.Anything, adminID).
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	existing := helpers.NewEcho(func(e *echoModel.Echo) { e.ID = echoID; e.UserID = adminID })
	repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&existing, nil).Once()
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
	repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()
	repo.EXPECT().UpdateEcho(mock.Anything, mock.Anything).Return(boom).Once()

	svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
	_, err := svc.UpdateEcho(helpers.CtxAsUser(adminID, role.Admin), &echoModel.Echo{ID: echoID, Content: "x", Layout: echoModel.LayoutGrid})
	require.ErrorIs(t, err, boom)
}
