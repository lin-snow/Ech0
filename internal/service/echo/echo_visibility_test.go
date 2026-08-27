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
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func nilBus() *busen.Bus { return nil }

const (
	adminID = "admin-0001"
	userID  = "user-0002"
	echoID  = "echo-0001"
)

func TestGetEchoById_Visibility(t *testing.T) {
	t.Run("anonymous cannot read private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAnonymous(), echoID)

		require.Error(t, err)
		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
		assert.Nil(t, got)
	})

	t.Run("anonymous can read public echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		public := helpers.NewEcho()
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&public, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAnonymous(), echoID)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.False(t, got.Private)
	})

	t.Run("non-admin user cannot read private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAsUser(userID), echoID)

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
		assert.Nil(t, got)
	})

	t.Run("admin can read private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAsUser(adminID), echoID)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.True(t, got.Private)
	})

	t.Run("not found returns ECHO_NOT_FOUND", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAnonymous(), echoID)

		require.EqualError(t, err, commonModel.ECHO_NOT_FOUND)
		assert.Nil(t, got)
	})

	t.Run("repository error is propagated", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repoErr := errors.New("db down")
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, repoErr).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		got, err := svc.GetEchoById(helpers.CtxAsUser(adminID), echoID)

		require.ErrorIs(t, err, repoErr)
		assert.Nil(t, got)
	})
}

func runTx(_ context.Context, fn func(ctx context.Context) error) error {
	return fn(context.Background())
}

func TestLikeEcho_Visibility(t *testing.T) {
	t.Run("anonymous cannot like private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAnonymous(), echoID)

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("non-admin user cannot like private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, userID).
			Return(helpers.NewUser(), nil).
			Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAsUser(userID), echoID)

		require.EqualError(t, err, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("admin can like private echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		private := helpers.NewEcho(helpers.AsPrivate)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&private, nil).Once()
		common.EXPECT().
			CommonGetUserByUserId(mock.Anything, adminID).
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().LikeEcho(mock.Anything, echoID).Return(nil).Once()
		repo.EXPECT().InvalidateEchoCaches(echoID).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAsUser(adminID), echoID)

		require.NoError(t, err)
	})

	t.Run("anonymous can like public echo", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		tx := txmock.NewMockTransactor(t)
		public := helpers.NewEcho()
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(&public, nil).Once()
		tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(runTx).Once()
		repo.EXPECT().LikeEcho(mock.Anything, echoID).Return(nil).Once()
		repo.EXPECT().InvalidateEchoCaches(echoID).Once()

		svc := echoService.NewEchoService(tx, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAnonymous(), echoID)

		require.NoError(t, err)
	})

	t.Run("not found returns ECHO_NOT_FOUND", func(t *testing.T) {
		repo := echomock.NewMockRepository(t)
		common := commonmock.NewMockService(t)
		repo.EXPECT().GetEchosById(mock.Anything, echoID).Return(nil, nil).Once()

		svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
		err := svc.LikeEcho(helpers.CtxAnonymous(), echoID)

		require.EqualError(t, err, commonModel.ECHO_NOT_FOUND)
	})
}

func TestQueryEchos_PageSizeClamp(t *testing.T) {
	cases := []struct {
		name         string
		inPage       int
		inPageSize   int
		wantPage     int
		wantPageSize int
	}{
		{"zero pagesize falls back to 10", 1, 0, 1, 10},
		{"negative pagesize falls back to 10", 1, -5, 1, 10},
		{"oversized pagesize clamps to 100", 1, 500, 1, 100},
		{"exactly 100 is kept", 1, 100, 1, 100},
		{"in-range pagesize is kept", 2, 50, 2, 50},
		{"zero page clamps to 1", 0, 20, 1, 20},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := echomock.NewMockRepository(t)
			common := commonmock.NewMockService(t)

			var captured commonModel.EchoQueryDto
			repo.EXPECT().
				QueryEchos(mock.Anything, false).
				Run(func(dto commonModel.EchoQueryDto, _ bool) { captured = dto }).
				Return([]echoModel.Echo{}, int64(0), nil).
				Once()

			svc := echoService.NewEchoService(nil, common, nil, repo, nilBus)
			_, err := svc.QueryEchos(helpers.CtxAnonymous(), commonModel.EchoQueryDto{
				Page:     tc.inPage,
				PageSize: tc.inPageSize,
			})

			require.NoError(t, err)
			assert.Equal(t, tc.wantPageSize, captured.PageSize, "pageSize clamp")
			assert.Equal(t, tc.wantPage, captured.Page, "page clamp")
			assert.Equal(t, "created_at", captured.SortBy)
			assert.Equal(t, "desc", captured.SortOrder)
		})
	}
}
