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
	PostEcho(ctx context.Context, newEcho *model.Echo) (*model.Echo, error)
	GetEchosByPage(ctx context.Context, pageQueryDto commonModel.PageQueryDto) (commonModel.List[model.Echo], error)
	DeleteEchoById(ctx context.Context, id string) error
	GetTodayEchos(ctx context.Context, timezone string) ([]model.Echo, error)
	UpdateEcho(ctx context.Context, echo *model.Echo) (*model.Echo, error)
	ReactToEcho(ctx context.Context, id, emoji string) ([]model.ReactionCount, error)
	UnreactEcho(ctx context.Context, id, emoji string) ([]model.ReactionCount, error)
	GetReactionCatalog() []model.ReactionEmoji
	GetEchoById(ctx context.Context, id string) (*model.Echo, error)
	GetAllTags() ([]model.Tag, error)
	CreateTag(ctx context.Context, name string) (*model.Tag, error)
	DeleteTag(ctx context.Context, id string) error
	GetEchosByTagId(ctx context.Context, tagId string, pageQueryDto commonModel.PageQueryDto) (commonModel.List[model.Echo], error)
	QueryEchos(ctx context.Context, queryDto commonModel.EchoQueryDto) (commonModel.List[model.Echo], error)
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
	GetEchosByPage(page, pageSize int, search string, scope model.PrivateScope) ([]model.Echo, int64)
	GetTodayEchos(scope model.PrivateScope, timezone string) []model.Echo
	GetEchosById(ctx context.Context, id string) (*model.Echo, error)
	UpdateEcho(ctx context.Context, echo *model.Echo) error
	DeleteEchoById(ctx context.Context, id string) error
	AddReaction(ctx context.Context, echoID, emoji, actor string) error
	RemoveReaction(ctx context.Context, echoID, emoji, actor string) error
	ListReactions(ctx context.Context, echoIDs []string, actor string) (map[string][]model.ReactionCount, error)
	InvalidateEchoCaches(echoIDs ...string)
	CreateTag(ctx context.Context, tag *model.Tag) error
	GetAllTags() ([]model.Tag, error)
	GetTagsByNames(ctx context.Context, names []string) ([]*model.Tag, error)
	IncrementTagUsageCount(ctx context.Context, tagID string) error
	DeleteTagById(ctx context.Context, id string) error
	GetEchosByTagId(tagID string, page, pageSize int, search string, scope model.PrivateScope) ([]model.Echo, int64, error)
	QueryEchos(queryDto commonModel.EchoQueryDto, scope model.PrivateScope) ([]model.Echo, int64, error)
	GetHotEchos(limit int, scope model.PrivateScope) ([]model.Echo, error)
	GetRandomEcho(scope model.PrivateScope) (*model.Echo, error)
	GetOnThisDayEchos(scope model.PrivateScope, timezone string) []model.Echo
}
