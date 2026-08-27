// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	fileService "github.com/lin-snow/ech0/internal/service/file"
)

type Service interface {
	PostEcho(ctx context.Context, newEcho *model.Echo) error
	GetEchosByPage(ctx context.Context, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
	DeleteEchoById(ctx context.Context, id string) error
	GetTodayEchos(ctx context.Context, timezone string) ([]model.Echo, error)
	UpdateEcho(ctx context.Context, echo *model.Echo) error
	LikeEcho(ctx context.Context, id string) error
	GetEchoById(ctx context.Context, id string) (*model.Echo, error)
	GetAllTags() ([]model.Tag, error)
	CreateTag(ctx context.Context, name string) (*model.Tag, error)
	DeleteTag(ctx context.Context, id string) error
	GetEchosByTagId(ctx context.Context, tagId string, pageQueryDto commonModel.PageQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
	QueryEchos(ctx context.Context, queryDto commonModel.EchoQueryDto) (commonModel.PageQueryResult[[]model.Echo], error)
	GetHotEchos(ctx context.Context, limit int) ([]model.Echo, error)
	GetRandomEcho(ctx context.Context) (*model.Echo, error)
	GetOnThisDayEchos(ctx context.Context, timezone string) ([]model.Echo, error)
}

type (
	CommonService = commonService.Service
	FileService   = fileService.Service
)

type Repository interface {
	CreateEcho(ctx context.Context, newEcho *model.Echo) error
	GetEchosByPage(page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64)
	GetTodayEchos(showPrivate bool, timezone string) []model.Echo
	GetEchosById(ctx context.Context, id string) (*model.Echo, error)
	UpdateEcho(ctx context.Context, echo *model.Echo) error
	DeleteEchoById(ctx context.Context, id string) error
	LikeEcho(ctx context.Context, id string) error
	InvalidateEchoCaches(echoIDs ...string)
	CreateTag(ctx context.Context, tag *model.Tag) error
	GetAllTags() ([]model.Tag, error)
	GetTagsByNames(ctx context.Context, names []string) ([]*model.Tag, error)
	IncrementTagUsageCount(ctx context.Context, tagID string) error
	DeleteTagById(ctx context.Context, id string) error
	GetEchosByTagId(tagID string, page, pageSize int, search string, showPrivate bool) ([]model.Echo, int64, error)
	QueryEchos(queryDto commonModel.EchoQueryDto, showPrivate bool) ([]model.Echo, int64, error)
	GetHotEchos(limit int, showPrivate bool) ([]model.Echo, error)
	GetRandomEcho(showPrivate bool) (*model.Echo, error)
	GetOnThisDayEchos(showPrivate bool, timezone string) []model.Echo
}
