// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	AddConnect(ctx context.Context, connected model.Connected) error
	DeleteConnect(ctx context.Context, id string) error
	GetConnect() (model.Connect, error)
	GetConnectsInfo() ([]model.Connect, error)
	GetConnects() ([]model.Connected, error)
	GetConnectsHealth() ([]model.ConnectedHealth, error)
}

type Repository interface {
	GetAllConnects(ctx context.Context) ([]model.Connected, error)
	CreateConnect(ctx context.Context, connected *model.Connected) error
	DeleteConnect(ctx context.Context, id string) error
}

type EchoRepository interface {
	GetTodayEchos(showPrivate bool, timezone string) []echoModel.Echo
	GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]echoModel.Echo, int64)
}

type CommonService = commonService.Service
