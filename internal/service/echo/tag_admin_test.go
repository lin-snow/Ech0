// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/authz"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	txmock "github.com/lin-snow/ech0/internal/test/mocks/txmock"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTag(t *testing.T) {
	t.Run("member creates a tag", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		tx := txmock.NewMockTransactor(t)
		repo.EXPECT().
			GetTagsByNames(mock.Anything, []string{"golang"}).
			Return([]*echoModel.Tag{}, nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().CreateTag(mock.Anything, mock.Anything).Return(nil).Once()

		svc := echoService.NewEchoService(tx, nil, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(userID), "golang")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "golang", got.Name)
	})

	t.Run("blank or hash-only name is invalid", func(t *testing.T) {
		cases := []string{"", "   ", "#"}
		for _, name := range cases {
			t.Run("name="+name, func(t *testing.T) {
				repo := newEchoRepoMock(t)

				svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
				got, err := svc.CreateTag(helpers.CtxAsUser(adminID, role.Admin), name)

				require.EqualError(t, err, commonModel.INVALID_PARAMS)
				assert.Nil(t, got)
			})
		}
	})

	t.Run("unsafe name is invalid", func(t *testing.T) {
		repo := newEchoRepoMock(t)

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID, role.Admin), "<script>")

		require.EqualError(t, err, commonModel.INVALID_PARAMS)
		assert.Nil(t, got)
	})

	t.Run("returns existing tag without creating", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		existing := &echoModel.Tag{ID: "tag-go", Name: "golang", UsageCount: 7}
		repo.EXPECT().
			GetTagsByNames(mock.Anything, []string{"golang"}).
			Return([]*echoModel.Tag{existing}, nil).
			Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID, role.Admin), "#golang  ")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, existing, got)
	})

	t.Run("creates new tag in transaction", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		tx := txmock.NewMockTransactor(t)
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

		svc := echoService.NewEchoService(tx, nil, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID, role.Admin), "#vue")

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "vue", got.Name)
		assert.Equal(t, 0, got.UsageCount)
		assert.Equal(t, "vue", created.Name)
	})

	t.Run("GetTagsByNames error propagates", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		boom := errors.New("query failed")
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return(nil, boom).Once()

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID, role.Admin), "golang")

		require.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})

	t.Run("CreateTag error inside transaction propagates", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		tx := txmock.NewMockTransactor(t)
		boom := errors.New("insert failed")
		repo.EXPECT().GetTagsByNames(mock.Anything, mock.Anything).Return([]*echoModel.Tag{}, nil).Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().CreateTag(mock.Anything, mock.Anything).Return(boom).Once()

		svc := echoService.NewEchoService(tx, nil, nil, repo, nilBus)
		got, err := svc.CreateTag(helpers.CtxAsUser(adminID, role.Admin), "golang")

		require.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})
}

func TestDeleteTag(t *testing.T) {
	const tagID = "tag-to-delete"

	t.Run("non-admin is denied", func(t *testing.T) {
		repo := newEchoRepoMock(t)

		svc := echoService.NewEchoService(nil, nil, nil, repo, nilBus)
		err := svc.DeleteTag(helpers.CtxAsUser(userID), tagID)

		require.ErrorIs(t, err, authz.ErrDenied)
	})

	t.Run("admin deletes in transaction", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		tx := txmock.NewMockTransactor(t)
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().DeleteTagById(mock.Anything, tagID).Return(nil).Once()

		svc := echoService.NewEchoService(tx, nil, nil, repo, nilBus)
		require.NoError(t, svc.DeleteTag(helpers.CtxAsUser(adminID, role.Admin), tagID))
	})

	t.Run("DeleteTagById error propagates", func(t *testing.T) {
		repo := newEchoRepoMock(t)
		tx := txmock.NewMockTransactor(t)
		boom := errors.New("delete failed")
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().DeleteTagById(mock.Anything, tagID).Return(boom).Once()

		svc := echoService.NewEchoService(tx, nil, nil, repo, nilBus)
		require.ErrorIs(t, svc.DeleteTag(helpers.CtxAsUser(adminID, role.Admin), tagID), boom)
	})
}
