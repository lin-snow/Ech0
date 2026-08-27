// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"testing"

	initHandler "github.com/lin-snow/ech0/internal/handler/init"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	initModel "github.com/lin-snow/ech0/internal/model/init"
	initmock "github.com/lin-snow/ech0/internal/test/mocks/initmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetInitStatus(t *testing.T) {
	cases := []struct {
		name string
		want initModel.Status
	}{
		{name: "fresh-system", want: initModel.Status{Initialized: false, OwnerExists: false}},
		{name: "owner-exists", want: initModel.Status{Initialized: true, OwnerExists: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := initmock.NewMockService(t)
			svc.EXPECT().GetStatus().Return(tc.want, nil).Once()

			h := initHandler.NewInitHandler(svc)
			out, err := h.GetInitStatus(context.Background(), &initHandler.GetInitStatusInput{})

			require.NoError(t, err)
			assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
			assert.Equal(t, tc.want, out.Data)
		})
	}
}

func TestGetInitStatus_ServiceError(t *testing.T) {
	svc := initmock.NewMockService(t)
	sentinel := errors.New("db unavailable")
	svc.EXPECT().GetStatus().Return(initModel.Status{}, sentinel).Once()

	h := initHandler.NewInitHandler(svc)
	out, err := h.GetInitStatus(context.Background(), &initHandler.GetInitStatusInput{})

	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, initHandler.StatusOutput{}, out)
}

func TestInitOwner(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := initmock.NewMockService(t)
		dto := authModel.RegisterDto{Username: "owner", Password: "s3cret", Email: "o@example.com"}
		svc.EXPECT().
			InitOwner(mock.MatchedBy(func(d *authModel.RegisterDto) bool {
				return d != nil && d.Username == "owner" && d.Password == "s3cret"
			})).
			Return(nil).Once()

		h := initHandler.NewInitHandler(svc)
		out, err := h.InitOwner(context.Background(), &initHandler.InitOwnerInput{Body: dto})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.INIT_OWNER_SUCCESS, out.Message)
		assert.Nil(t, out.Data)
	})

	t.Run("owner-already-exists", func(t *testing.T) {
		svc := initmock.NewMockService(t)
		bizErr := commonModel.NewBizError(commonModel.ErrCodeInitOwnerExists, commonModel.OWNER_ALREADY_EXISTS)
		svc.EXPECT().InitOwner(mock.Anything).Return(bizErr).Once()

		h := initHandler.NewInitHandler(svc)
		out, err := h.InitOwner(context.Background(), &initHandler.InitOwnerInput{Body: authModel.RegisterDto{Username: "x", Password: "y"}})

		require.Error(t, err)
		var got *commonModel.BizError
		require.ErrorAs(t, err, &got)
		assert.Equal(t, commonModel.ErrCodeInitOwnerExists, got.Code)
		assert.Equal(t, initHandler.EmptyOutput{}, out)
	})
}
