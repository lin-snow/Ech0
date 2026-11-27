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
		Body connectModel.AddConnectDto
	}
	DeleteConnectInput struct {
		ID string `path:"id" format:"uuid" doc:"Connection ID (UUID)"`
	}
)

type (
	ConnectOutput       = struct{ Body connectModel.Connect }
	ConnectedListOutput = struct {
		Body commonModel.List[connectModel.Connected]
	}
	ConnectListOutput = struct {
		Body commonModel.List[connectModel.Connect]
	}
	ConnectHealthOutput = struct {
		Body commonModel.List[connectModel.ConnectedHealth]
	}
	EmptyOutput = struct{}
)

func (connectHandler *ConnectHandler) GetConnect(_ context.Context, _ *GetConnectInput) (*ConnectOutput, error) {
	connect, err := connectHandler.connectService.GetConnect()
	if err != nil {
		return nil, err
	}
	return &ConnectOutput{Body: connect}, nil
}

func (connectHandler *ConnectHandler) GetConnects(_ context.Context, _ *GetConnectsInput) (*ConnectedListOutput, error) {
	connects, err := connectHandler.connectService.GetConnects()
	if err != nil {
		return nil, err
	}
	return &ConnectedListOutput{Body: commonModel.List[connectModel.Connected]{Items: connects}}, nil
}

func (connectHandler *ConnectHandler) GetConnectsInfo(_ context.Context, _ *GetConnectsInfoInput) (*ConnectListOutput, error) {
	connects, err := connectHandler.connectService.GetConnectsInfo()
	if err != nil {
		return nil, err
	}
	return &ConnectListOutput{Body: commonModel.List[connectModel.Connect]{Items: connects}}, nil
}

func (connectHandler *ConnectHandler) GetConnectsHealth(_ context.Context, _ *GetConnectsHealthIn) (*ConnectHealthOutput, error) {
	rows, err := connectHandler.connectService.GetConnectsHealth()
	if err != nil {
		return nil, err
	}
	return &ConnectHealthOutput{Body: commonModel.List[connectModel.ConnectedHealth]{Items: rows}}, nil
}

func (connectHandler *ConnectHandler) AddConnect(ctx context.Context, in *AddConnectInput) (*EmptyOutput, error) {
	connected := connectModel.Connected{ConnectURL: in.Body.ConnectURL}
	if err := connectHandler.connectService.AddConnect(ctx, connected); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (connectHandler *ConnectHandler) DeleteConnect(ctx context.Context, in *DeleteConnectInput) (*EmptyOutput, error) {
	if err := connectHandler.connectService.DeleteConnect(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}
