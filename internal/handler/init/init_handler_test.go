// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/lin-snow/ech0/internal/apperr"
	initHandler "github.com/lin-snow/ech0/internal/handler/init"
	"github.com/lin-snow/ech0/internal/httperr"
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
			require.NotNil(t, out)
			assert.Equal(t, tc.want, out.Body)
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
	assert.Nil(t, out)
}

func TestInitOwner(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := initmock.NewMockService(t)
		dto := initModel.InitOwnerDto{Username: "owner", Password: "s3cret", Email: "o@example.com", SiteTitle: "My notes"}
		svc.EXPECT().
			InitOwner(mock.MatchedBy(func(d *initModel.InitOwnerDto) bool {
				return d != nil && d.Username == "owner" && d.Password == "s3cret" && d.SiteTitle == "My notes"
			})).
			Return(nil).Once()

		h := initHandler.NewInitHandler(svc)
		out, err := h.InitOwner(context.Background(), &initHandler.InitOwnerInput{Body: dto})

		require.NoError(t, err)
		assert.Equal(t, &initHandler.EmptyOutput{}, out)
	})

	t.Run("owner-already-exists", func(t *testing.T) {
		svc := initmock.NewMockService(t)
		failure := apperr.ErrOwnerExists
		svc.EXPECT().InitOwner(mock.Anything).Return(failure).Once()

		h := initHandler.NewInitHandler(svc)
		out, err := h.InitOwner(context.Background(), &initHandler.InitOwnerInput{Body: initModel.InitOwnerDto{Username: "x", Password: "y"}})

		require.ErrorIs(t, err, apperr.ErrOwnerExists)
		assert.Equal(t, http.StatusConflict, httperr.From(err).Status)
		assert.Nil(t, out)
	})
}
