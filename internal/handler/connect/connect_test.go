// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"testing"

	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	connectmock "github.com/lin-snow/ech0/internal/test/mocks/connectmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func assertBizErr(t *testing.T, err error, wantCode string) {
	t.Helper()
	require.Error(t, err)
	var be *commonModel.BizError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, wantCode, be.Code)
}

func bizErr() *commonModel.BizError {
	return commonModel.NewBizError(commonModel.ErrCodeInternal, "boom")
}

func TestConnectHandler_GetConnect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		want := connectModel.Connect{ServerName: "ech0", TotalEchos: 7, TodayEchos: 2}
		svc.EXPECT().GetConnect().Return(want, nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnect(context.Background(), &connectHandler.GetConnectInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.CONNECT_SUCCESS, out.Message)
		assert.Equal(t, want, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnect().Return(connectModel.Connect{}, bizErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnect(context.Background(), &connectHandler.GetConnectInput{})

		assertBizErr(t, err, commonModel.ErrCodeInternal)
		assert.Equal(t, 0, out.Code)
		assert.Empty(t, out.Data)
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
		assert.Equal(t, commonModel.GET_CONNECTED_LIST_SUCCESS, out.Message)
		assert.Equal(t, want, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnects().Return(nil, bizErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnects(context.Background(), &connectHandler.GetConnectsInput{})

		assertBizErr(t, err, commonModel.ErrCodeInternal)
		assert.Nil(t, out.Data)
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
		assert.Equal(t, commonModel.GET_CONNECT_INFO_SUCCESS, out.Message)
		assert.Equal(t, want, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnectsInfo().Return(nil, bizErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnectsInfo(context.Background(), &connectHandler.GetConnectsInfoInput{})

		assertBizErr(t, err, commonModel.ErrCodeInternal)
		assert.Nil(t, out.Data)
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
		assert.Equal(t, commonModel.GET_CONNECT_HEALTH_SUCCESS, out.Message)
		assert.Equal(t, want, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().GetConnectsHealth().Return(nil, bizErr()).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.GetConnectsHealth(context.Background(), &connectHandler.GetConnectsHealthIn{})

		assertBizErr(t, err, commonModel.ErrCodeInternal)
		assert.Nil(t, out.Data)
	})
}

func TestConnectHandler_AddConnect(t *testing.T) {
	t.Run("success passes body through", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		body := connectModel.Connected{ConnectURL: "https://new.example"}
		svc.EXPECT().AddConnect(mock.Anything, body).Return(nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.AddConnect(context.Background(), &connectHandler.AddConnectInput{Body: body})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.ADD_CONNECT_SUCCESS, out.Message)
		assert.Nil(t, out.Data)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		be := commonModel.NewBizError(commonModel.ErrCodeInvalidRequest, commonModel.INVALID_CONNECTION_URL)
		svc.EXPECT().AddConnect(mock.Anything, mock.Anything).Return(be).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.AddConnect(context.Background(), &connectHandler.AddConnectInput{})

		assertBizErr(t, err, commonModel.ErrCodeInvalidRequest)
		assert.Equal(t, 0, out.Code)
	})
}

func TestConnectHandler_DeleteConnect(t *testing.T) {
	t.Run("success passes id through", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().DeleteConnect(mock.Anything, "id-123").Return(nil).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.DeleteConnect(context.Background(), &connectHandler.DeleteConnectInput{ID: "id-123"})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DELETE_CONNECT_SUCCESS, out.Message)
	})

	t.Run("error", func(t *testing.T) {
		svc := connectmock.NewMockService(t)
		svc.EXPECT().DeleteConnect(mock.Anything, mock.Anything).Return(errors.New("nope")).Once()

		h := connectHandler.NewConnectHandler(svc)
		out, err := h.DeleteConnect(context.Background(), &connectHandler.DeleteConnectInput{ID: "x"})

		require.Error(t, err)
		assert.Equal(t, 0, out.Code)
	})
}
