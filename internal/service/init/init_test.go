// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"errors"
	"testing"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	initModel "github.com/lin-snow/ech0/internal/model/init"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	initService "github.com/lin-snow/ech0/internal/service/init"
	"github.com/lin-snow/ech0/internal/test/mocks/settingmock"
	"github.com/lin-snow/ech0/internal/test/mocks/usermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeInitRepo struct {
	initialized    bool
	initializedErr error
	owner          userModel.User
	getOwnerErr    error

	isInitCalls   int
	getOwnerCalls int
}

func (f *fakeInitRepo) IsInitialized() (bool, error) {
	f.isInitCalls++
	return f.initialized, f.initializedErr
}

func (f *fakeInitRepo) GetOwner() (userModel.User, error) {
	f.getOwnerCalls++
	return f.owner, f.getOwnerErr
}

func asBizError(t *testing.T, err error) *commonModel.BizError {
	t.Helper()
	var biz *commonModel.BizError
	require.True(t, errors.As(err, &biz), "expected *commonModel.BizError, got %T (%v)", err, err)
	return biz
}

func TestInitService_GetStatus(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("db down")

	cases := []struct {
		name        string
		repo        *fakeInitRepo
		wantErr     error
		wantStatus  initModel.Status
		wantOwnerHi bool
	}{
		{
			name:       "IsInitialized error short-circuits",
			repo:       &fakeInitRepo{initializedErr: sentinel},
			wantErr:    sentinel,
			wantStatus: initModel.Status{},
		},
		{
			name:        "initialized with owner",
			repo:        &fakeInitRepo{initialized: true},
			wantStatus:  initModel.Status{Initialized: true, OwnerExists: true},
			wantOwnerHi: true,
		},
		{
			name:        "initialized but owner lookup fails -> OwnerExists false",
			repo:        &fakeInitRepo{initialized: true, getOwnerErr: errors.New("no owner")},
			wantStatus:  initModel.Status{Initialized: true, OwnerExists: false},
			wantOwnerHi: true,
		},
		{
			name:        "not initialized and no owner",
			repo:        &fakeInitRepo{initialized: false, getOwnerErr: errors.New("no owner")},
			wantStatus:  initModel.Status{Initialized: false, OwnerExists: false},
			wantOwnerHi: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			us := usermock.NewMockService(t)
			ss := settingmock.NewMockService(t)
			svc := initService.NewInitService(tc.repo, us, ss)

			got, err := svc.GetStatus()

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, initModel.Status{}, got)
				assert.Equal(t, 0, tc.repo.getOwnerCalls)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, got)
			assert.Equal(t, 1, tc.repo.isInitCalls)
			if tc.wantOwnerHi {
				assert.Equal(t, 1, tc.repo.getOwnerCalls)
			}
		})
	}
}

func TestInitService_InitOwner(t *testing.T) {
	t.Parallel()

	dto := &authModel.RegisterDto{Username: "owner", Password: "pw", Locale: "ja"}

	t.Run("IsInitialized error short-circuits before user service", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("db down")
		repo := &fakeInitRepo{initializedErr: sentinel}
		us := usermock.NewMockService(t)
		ss := settingmock.NewMockService(t)
		svc := initService.NewInitService(repo, us, ss)

		err := svc.InitOwner(dto)

		require.Error(t, err)
		assert.ErrorIs(t, err, sentinel)
		assert.Equal(t, 1, repo.isInitCalls)
	})

	t.Run("already initialized returns INIT_ALREADY_DONE biz error", func(t *testing.T) {
		t.Parallel()

		repo := &fakeInitRepo{initialized: true}
		us := usermock.NewMockService(t)
		ss := settingmock.NewMockService(t)
		svc := initService.NewInitService(repo, us, ss)

		err := svc.InitOwner(dto)

		biz := asBizError(t, err)
		assert.Equal(t, commonModel.ErrCodeInitAlreadyDone, biz.Code)
		assert.Equal(t, commonModel.SYSTEM_ALREADY_INITED, biz.Msg)
	})

	t.Run("user service InitOwner error is propagated, locale not bootstrapped", func(t *testing.T) {
		t.Parallel()

		sentinel := errors.New("create owner failed")
		repo := &fakeInitRepo{initialized: false}
		us := usermock.NewMockService(t)
		us.EXPECT().InitOwner(dto).Return(sentinel).Once()
		ss := settingmock.NewMockService(t)
		svc := initService.NewInitService(repo, us, ss)

		err := svc.InitOwner(dto)

		require.Error(t, err)
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("happy path bootstraps locale and returns nil", func(t *testing.T) {
		t.Parallel()

		repo := &fakeInitRepo{initialized: false}
		us := usermock.NewMockService(t)
		us.EXPECT().InitOwner(dto).Return(nil).Once()
		ss := settingmock.NewMockService(t)
		ss.EXPECT().BootstrapDefaultLocale(mock.Anything, dto.Locale).Return(nil).Once()
		svc := initService.NewInitService(repo, us, ss)

		err := svc.InitOwner(dto)

		require.NoError(t, err)
	})

	t.Run("locale bootstrap failure is best-effort and does not fail init", func(t *testing.T) {
		t.Parallel()

		repo := &fakeInitRepo{initialized: false}
		us := usermock.NewMockService(t)
		us.EXPECT().InitOwner(dto).Return(nil).Once()
		ss := settingmock.NewMockService(t)
		ss.EXPECT().
			BootstrapDefaultLocale(mock.Anything, dto.Locale).
			Return(errors.New("locale write failed")).
			Once()
		svc := initService.NewInitService(repo, us, ss)

		err := svc.InitOwner(dto)

		require.NoError(t, err)
	})
}
