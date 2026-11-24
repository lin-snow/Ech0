// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	"github.com/lin-snow/ech0/internal/apperr"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	service "github.com/lin-snow/ech0/internal/service/echo"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
)

type EchoHandler struct {
	echoService service.Service
}

func NewEchoHandler(echoService service.Service) *EchoHandler {
	return &EchoHandler{
		echoService: echoService,
	}
}

type (
	EchoUpsertInput struct {
		Body model.EchoUpsertDto
	}
	EchoIDInput struct {
		ID string `path:"id" format:"uuid" doc:"Echo ID"`
	}
	QueryEchosInput struct {
		Body commonModel.EchoQueryDto
	}
	EchoPageGetInput struct {
		Page     int    `query:"page"`
		PageSize int    `query:"pageSize"`
		Search   string `query:"search"`
	}
	EchoPagePostInput struct {
		Body commonModel.PageQueryDto
	}
	GetEchosByTagIDInput struct {
		TagID    string `path:"tagid" format:"uuid" doc:"Tag ID"`
		Page     int    `query:"page"`
		PageSize int    `query:"pageSize"`
		Search   string `query:"search"`
	}
	TimezoneInput struct {
		Timezone string `header:"X-Timezone" doc:"客户端时区（IANA 名）"`
	}
	GetHotEchosInput struct {
		Limit int `query:"limit" default:"5" doc:"Number of Echoes to return; default 5, maximum 20"`
	}
	GetRandomEchoInput struct{}
	GetAllTagsInput    struct{}
	CreateTagInput     struct {
		Body model.CreateTagDto
	}
	TagIDInput struct {
		ID string `path:"id" format:"uuid" doc:"Tag ID"`
	}
	ReactionInput struct {
		ID    string `path:"id"    format:"uuid" doc:"Echo ID"`
		Emoji string `path:"emoji" doc:"Reaction key" enum:"like,heart,haha,wow,sad,cry,party"`
	}
	ReactionCatalogInput struct{}
)

type (
	EchoOutput = struct{ Body model.Echo }
	// EchoCreatedOutput adds the canonical URI of the freshly created Echo, so a
	// 201 can point at it (RFC 9110 6.4.2).
	EchoCreatedOutput struct {
		Location string `header:"Location"`
		Body     model.Echo
	}
	EchoListOutput = struct{ Body commonModel.List[model.Echo] }
	EchoPageOutput = struct{ Body commonModel.List[model.Echo] }
	TagOutput      = struct{ Body model.Tag }
	TagListOutput  = struct{ Body commonModel.List[model.Tag] }
	EmptyOutput    = struct{}

	ReactionListOutput = struct {
		Body commonModel.List[model.ReactionCount]
	}
	ReactionCatalogOutput = struct {
		Body commonModel.List[model.ReactionEmoji]
	}
)

func (echoHandler *EchoHandler) PostEcho(ctx context.Context, in *EchoUpsertInput) (*EchoCreatedOutput, error) {
	saved, err := echoHandler.echoService.PostEcho(ctx, in.Body.ToModel())
	if err != nil {
		return nil, err
	}
	return &EchoCreatedOutput{Location: "/api/echo/" + saved.ID, Body: *saved}, nil
}

func (echoHandler *EchoHandler) UpdateEcho(ctx context.Context, in *EchoUpsertInput) (*EchoOutput, error) {
	saved, err := echoHandler.echoService.UpdateEcho(ctx, in.Body.ToModel())
	if err != nil {
		return nil, err
	}
	return &EchoOutput{Body: *saved}, nil
}

func (echoHandler *EchoHandler) DeleteEcho(ctx context.Context, in *EchoIDInput) (*EmptyOutput, error) {
	if err := echoHandler.echoService.DeleteEchoById(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (echoHandler *EchoHandler) AddReaction(ctx context.Context, in *ReactionInput) (*ReactionListOutput, error) {
	reactions, err := echoHandler.echoService.ReactToEcho(ctx, in.ID, in.Emoji)
	if err != nil {
		return nil, err
	}
	return &ReactionListOutput{Body: commonModel.List[model.ReactionCount]{Items: reactions}}, nil
}

func (echoHandler *EchoHandler) RemoveReaction(ctx context.Context, in *ReactionInput) (*ReactionListOutput, error) {
	reactions, err := echoHandler.echoService.UnreactEcho(ctx, in.ID, in.Emoji)
	if err != nil {
		return nil, err
	}
	return &ReactionListOutput{Body: commonModel.List[model.ReactionCount]{Items: reactions}}, nil
}

func (echoHandler *EchoHandler) GetReactionCatalog(_ context.Context, _ *ReactionCatalogInput) (*ReactionCatalogOutput, error) {
	return &ReactionCatalogOutput{Body: commonModel.List[model.ReactionEmoji]{Items: echoHandler.echoService.GetReactionCatalog()}}, nil
}

func (echoHandler *EchoHandler) GetEchoById(ctx context.Context, in *EchoIDInput) (*EchoOutput, error) {
	echo, err := echoHandler.echoService.GetEchoById(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &EchoOutput{Body: *echo}, nil
}

func (echoHandler *EchoHandler) QueryEchos(ctx context.Context, in *QueryEchosInput) (*EchoPageOutput, error) {
	result, err := echoHandler.echoService.QueryEchos(ctx, in.Body)
	if err != nil {
		return nil, err
	}
	return &EchoPageOutput{Body: result}, nil
}

func (echoHandler *EchoHandler) GetEchosByPageGet(ctx context.Context, in *EchoPageGetInput) (*EchoPageOutput, error) {
	return echoHandler.getEchosByPage(ctx, commonModel.PageQueryDto{Page: in.Page, PageSize: in.PageSize, Search: in.Search})
}

func (echoHandler *EchoHandler) GetEchosByPagePost(ctx context.Context, in *EchoPagePostInput) (*EchoPageOutput, error) {
	return echoHandler.getEchosByPage(ctx, in.Body)
}

func (echoHandler *EchoHandler) getEchosByPage(ctx context.Context, page commonModel.PageQueryDto) (*EchoPageOutput, error) {
	result, err := echoHandler.echoService.GetEchosByPage(ctx, page)
	if err != nil {
		return nil, err
	}
	return &EchoPageOutput{Body: result}, nil
}

func (echoHandler *EchoHandler) GetEchosByTagId(ctx context.Context, in *GetEchosByTagIDInput) (*EchoPageOutput, error) {
	result, err := echoHandler.echoService.GetEchosByTagId(ctx, in.TagID, commonModel.PageQueryDto{Page: in.Page, PageSize: in.PageSize, Search: in.Search})
	if err != nil {
		return nil, err
	}
	return &EchoPageOutput{Body: result}, nil
}

func (echoHandler *EchoHandler) GetTodayEchos(ctx context.Context, in *TimezoneInput) (*EchoListOutput, error) {
	result, err := echoHandler.echoService.GetTodayEchos(ctx, timezoneUtil.NormalizeTimezone(in.Timezone))
	if err != nil {
		return nil, err
	}
	return &EchoListOutput{Body: commonModel.List[model.Echo]{Items: result}}, nil
}

func (echoHandler *EchoHandler) GetHotEchos(ctx context.Context, in *GetHotEchosInput) (*EchoListOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 5
	}
	result, err := echoHandler.echoService.GetHotEchos(ctx, limit)
	if err != nil {
		return nil, err
	}
	return &EchoListOutput{Body: commonModel.List[model.Echo]{Items: result}}, nil
}

func (echoHandler *EchoHandler) GetRandomEcho(ctx context.Context, _ *GetRandomEchoInput) (*EchoOutput, error) {
	echo, err := echoHandler.echoService.GetRandomEcho(ctx)
	if err != nil {
		return nil, err
	}
	if echo == nil {
		// The resource is optional: an empty library is a 404, never a null body.
		return nil, apperr.ErrEchoNotFound
	}
	return &EchoOutput{Body: *echo}, nil
}

func (echoHandler *EchoHandler) GetOnThisDayEchos(ctx context.Context, in *TimezoneInput) (*EchoListOutput, error) {
	result, err := echoHandler.echoService.GetOnThisDayEchos(ctx, timezoneUtil.NormalizeTimezone(in.Timezone))
	if err != nil {
		return nil, err
	}
	return &EchoListOutput{Body: commonModel.List[model.Echo]{Items: result}}, nil
}

func (echoHandler *EchoHandler) GetAllTags(ctx context.Context, _ *GetAllTagsInput) (*TagListOutput, error) {
	tags, err := echoHandler.echoService.GetAllTags()
	if err != nil {
		return nil, err
	}
	return &TagListOutput{Body: commonModel.List[model.Tag]{Items: tags}}, nil
}

func (echoHandler *EchoHandler) CreateTag(ctx context.Context, in *CreateTagInput) (*TagOutput, error) {
	tag, err := echoHandler.echoService.CreateTag(ctx, in.Body.Name)
	if err != nil {
		return nil, err
	}
	return &TagOutput{Body: *tag}, nil
}

func (echoHandler *EchoHandler) DeleteTag(ctx context.Context, in *TagIDInput) (*EmptyOutput, error) {
	if err := echoHandler.echoService.DeleteTag(ctx, in.ID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}
