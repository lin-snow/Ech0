// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/lin-snow/ech0/internal/apperr"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	"github.com/lin-snow/ech0/internal/httperr"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	connectmock "github.com/lin-snow/ech0/internal/test/mocks/connectmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func assertStatusErr(t *testing.T, err error, wantStatus int) {
	t.Helper()
	require.Error(t, err)
	assert.Equal(t, wantStatus, httperr.From(err).Status)
}

func internalErr() error {
	return errors.New("boom")
}

func TestConnectHandler_GetConnect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		want := connectModel.Connect{ServerName: "ech0", TotalEchos: 7, TodayEchos: 2}
		svc.EXPECT().GetConnect().Return(want, nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnect(context.Background(), &connectHandler.GetConnectInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnect().Return(connectModel.Connect{}, internalErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnect(context.Background(), &connectHandler.GetConnectInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})
}

func TestConnectHandler_GetConnects(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		want := []connectModel.Connected{{ID: "c1", ConnectURL: "https://a.example"}}
		svc.EXPECT().GetConnects().Return(want, nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnects(context.Background(), &connectHandler.GetConnectsInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body.Items)
		assert.Nil(t, out.Body.Total)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnects().Return(nil, internalErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnects(context.Background(), &connectHandler.GetConnectsInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})
}

func TestConnectHandler_GetConnectsInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		want := []connectModel.Connect{{ServerName: "peer", ServerURL: "https://peer.example"}}
		svc.EXPECT().GetConnectsInfo().Return(want, nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnectsInfo(context.Background(), &connectHandler.GetConnectsInfoInput{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body.Items)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnectsInfo().Return(nil, internalErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnectsInfo(context.Background(), &connectHandler.GetConnectsInfoInput{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})
}

func TestConnectHandler_GetConnectsHealth(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		want := []connectModel.ConnectedHealth{{ID: "c1", Status: "online", Version: "v1"}}
		svc.EXPECT().GetConnectsHealth().Return(want, nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnectsHealth(context.Background(), &connectHandler.GetConnectsHealthIn{})

		require.NoError(t, err)
		require.NotNil(t, out)
		assert.Equal(t, want, out.Body.Items)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnectsHealth().Return(nil, internalErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnectsHealth(context.Background(), &connectHandler.GetConnectsHealthIn{})

		assertStatusErr(t, err, http.StatusInternalServerError)
		assert.Nil(t, out)
	})
}

func TestConnectHandler_AddConnect(t *testing.T) {
	t.Run("success passes body through", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		body := connectHandler.AddConnectInput{Body: connectModel.AddConnectDto{ConnectURL: "https://new.example"}}
		svc.EXPECT().
			AddConnect(mock.Anything, connectModel.Connected{ConnectURL: "https://new.example"}).
			Return(nil).
			Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.AddConnect(context.Background(), &body)

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().AddConnect(mock.Anything, mock.Anything).
			Return(apperr.ErrConnectURL).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.AddConnect(context.Background(), &connectHandler.AddConnectInput{})

		assertStatusErr(t, err, http.StatusBadRequest)
		assert.Nil(t, out)
	})
}

func TestConnectHandler_DeleteConnect(t *testing.T) {
	t.Run("success passes id through", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().DeleteConnect(mock.Anything, "id-123").Return(nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.DeleteConnect(context.Background(), &connectHandler.DeleteConnectInput{ID: "id-123"})

		require.NoError(t, err)
		require.NotNil(t, out)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().DeleteConnect(mock.Anything, mock.Anything).Return(errors.New("nope")).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.DeleteConnect(context.Background(), &connectHandler.DeleteConnectInput{ID: "x"})

		require.Error(t, err)
		assert.Nil(t, out)
	})
}
