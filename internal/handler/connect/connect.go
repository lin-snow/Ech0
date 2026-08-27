// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	service "github.com/lin-snow/ech0/internal/service/connect"
)

type ConnectHandler struct {
	connectService service.Service
}

func NewConnectHandler(connectService service.Service) *ConnectHandler {
	return &ConnectHandler{
		connectService: connectService,
	}
}

type (
	GetConnectInput      struct{}
	GetConnectsInput     struct{}
	GetConnectsInfoInput struct{}
	GetConnectsHealthIn  struct{}
	AddConnectInput      struct {
		Body connectModel.Connected
	}
	DeleteConnectInput struct {
		ID string `path:"id" format:"uuid" doc:"连接 ID（UUID）"`
	}
)

type (
	ConnectOutput       = commonModel.Result[connectModel.Connect]
	ConnectedListOutput = commonModel.Result[[]connectModel.Connected]
	ConnectListOutput   = commonModel.Result[[]connectModel.Connect]
	ConnectHealthOutput = commonModel.Result[[]connectModel.ConnectedHealth]
	EmptyOutput         = commonModel.Result[any]
)

func (connectHandler *ConnectHandler) GetConnect(ctx context.Context, _ *GetConnectInput) (ConnectOutput, error) {
	connect, err := connectHandler.connectService.GetConnect()
	if err != nil {
		return ConnectOutput{}, err
	}
	return commonModel.OK(connect, commonModel.CONNECT_SUCCESS), nil
}

func (connectHandler *ConnectHandler) GetConnects(ctx context.Context, _ *GetConnectsInput) (ConnectedListOutput, error) {
	connects, err := connectHandler.connectService.GetConnects()
	if err != nil {
		return ConnectedListOutput{}, err
	}
	return commonModel.OK(connects, commonModel.GET_CONNECTED_LIST_SUCCESS), nil
}

func (connectHandler *ConnectHandler) GetConnectsInfo(ctx context.Context, _ *GetConnectsInfoInput) (ConnectListOutput, error) {
	connects, err := connectHandler.connectService.GetConnectsInfo()
	if err != nil {
		return ConnectListOutput{}, err
	}
	return commonModel.OK(connects, commonModel.GET_CONNECT_INFO_SUCCESS), nil
}

func (connectHandler *ConnectHandler) GetConnectsHealth(ctx context.Context, _ *GetConnectsHealthIn) (ConnectHealthOutput, error) {
	rows, err := connectHandler.connectService.GetConnectsHealth()
	if err != nil {
		return ConnectHealthOutput{}, err
	}
	return commonModel.OK(rows, commonModel.GET_CONNECT_HEALTH_SUCCESS), nil
}

func (connectHandler *ConnectHandler) AddConnect(ctx context.Context, in *AddConnectInput) (EmptyOutput, error) {
	if err := connectHandler.connectService.AddConnect(ctx, in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.ADD_CONNECT_SUCCESS), nil
}

func (connectHandler *ConnectHandler) DeleteConnect(ctx context.Context, in *DeleteConnectInput) (EmptyOutput, error) {
	if err := connectHandler.connectService.DeleteConnect(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_CONNECT_SUCCESS), nil
}
