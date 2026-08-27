// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	echomock "github.com/lin-snow/ech0/internal/test/mocks/echomock"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTag(t *testing.T) {
	t.Run("non-admin is denied", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(userID), "golang")

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
		assert.Nil(t, got)
	})

	t.Run("user lookup error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		boom := errors.New("user lookup failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), boom).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID), "golang")

		require.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})

	t.Run("blank or hash-only name is invalid", func(t *testing.T) {
		cases := []string{"", "   ", "#"}
		for _, name := range cases {
			t.Run("name="+name, func(t *testing.T) {
				repo := echomock.NewMockRepository(t)
				common := commonmock.NewMockService(t)
				common.EXPECT().
					CommonGetUserByUserId(mock.Anything, adminID).
					Return(helpers.NewUser(helpers.AsAdmin), nil).
					Once()

				svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
				got, err := svc.CreateTag(helpers.CtxAsUser(adminID), name)

				require.EqualError(t, err, commonModel.INVALID_PARAMS)
				assert.Nil(t, got)
			})
		}
	})

	t.Run("unsafe name is invalid", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID), "<script>")

		require.EqualError(t, err, commonModel.INVALID_PARAMS)
		assert.Nil(t, got)
	})

	t.Run("returns existing tag without creating", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		existing := &echoModel.Tag{ID: "tag-go", Name: "golang", UsageCount: 7}
		repo.EXPECT().
			GetTagsByNames(mock.Anything, []string{"golang"}).
			Return([]*echoModel.Tag{existing}, nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID), "#golang  ")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, existing, got)
	})

	t.Run("creates new tag in transaction", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		repo.EXPECT().
			GetTagsByNames(mock.Anything, []string{"vue"}).
			Return([]*echoModel.Tag{}, nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()

		var created echoModel.Tag
		repo.EXPECT().
			CreateTag(mock.Anything, mock.Anything).
			Run(func(_ context.Context, tag *echoModel.Tag) { created = *tag }).
			Return(nil).
			Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID), "#vue")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "vue", got.Name)
		assert.Equal(t, 0, got.UsageCount)
		assert.Equal(t, "vue", created.Name)
	})

	t.Run("GetTagsByNames error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		boom := errors.New("query failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return(nil, boom).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID), "golang")

		require.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})

	t.Run("CreateTag error inside transaction propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		boom := errors.New("insert failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().CreateTag(mock.Anything, mock.Anything).Return(boom).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID), "golang")

		require.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})
}

func TestDeleteTag(t *testing.T) {
	const tagID = "tag-to-delete"

	t.Run("non-admin is denied", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.DeleteTag(helpers.CtxAsUser(userID), tagID)

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("user lookup error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		boom := errors.New("user lookup failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), boom).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.DeleteTag(helpers.CtxAsUser(adminID), tagID)

		require.ErrorIs(t, err, boom)
	})

	t.Run("admin deletes in transaction", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().DeleteTagById(mock.Anything, tagID).Return(nil).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		require.NoError(t, svc.DeleteTag(helpers.CtxAsUser(adminID), tagID))
	})

	t.Run("DeleteTagById error propagates", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		boom := errors.New("delete failed")
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().DeleteTagById(mock.Anything, tagID).Return(boom).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		require.ErrorIs(t, svc.DeleteTag(helpers.CtxAsUser(adminID), tagID), boom)
	})
}
