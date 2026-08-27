// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

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
		TagID    string `path:"tagid" format:"uuid" doc:"标签 ID"`
		Page     int    `query:"page"`
		PageSize int    `query:"pageSize"`
		Search   string `query:"search"`
	}
	TimezoneInput struct {
		Timezone string `header:"X-Timezone" doc:"客户端时区（IANA 名）"`
	}
	GetHotEchosInput struct {
		Limit int `query:"limit" default:"5" doc:"返回条数，默认 5，最大 20"`
	}
	GetRandomEchoInput struct{}
	GetAllTagsInput    struct{}
	CreateTagInput     struct {
		Body model.CreateTagDto
	}
	TagIDInput struct {
		ID string `path:"id" format:"uuid" doc:"标签 ID"`
	}
	LikeEchoInput struct {
		ID string `path:"id" format:"uuid" doc:"Echo ID"`
	}
)

type (
	EchoOutput     = commonModel.Result[*model.Echo]
	EchoListOutput = commonModel.Result[[]model.Echo]
	EchoPageOutput = commonModel.Result[commonModel.PageQueryResult[[]model.Echo]]
	TagOutput      = commonModel.Result[*model.Tag]
	TagListOutput  = commonModel.Result[[]model.Tag]
	EmptyOutput    = commonModel.Result[any]
)

func (echoHandler *EchoHandler) PostEcho(ctx context.Context, in *EchoUpsertInput) (EmptyOutput, error) {
	if err := echoHandler.echoService.PostEcho(ctx, in.Body.ToModel()); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.POST_ECHO_SUCCESS), nil
}

func (echoHandler *EchoHandler) UpdateEcho(ctx context.Context, in *EchoUpsertInput) (EmptyOutput, error) {
	if err := echoHandler.echoService.UpdateEcho(ctx, in.Body.ToModel()); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_ECHO_SUCCESS), nil
}

func (echoHandler *EchoHandler) DeleteEcho(ctx context.Context, in *EchoIDInput) (EmptyOutput, error) {
	if err := echoHandler.echoService.DeleteEchoById(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_ECHO_SUCCESS), nil
}

func (echoHandler *EchoHandler) LikeEcho(ctx context.Context, in *LikeEchoInput) (EmptyOutput, error) {
	if err := echoHandler.echoService.LikeEcho(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.LIKE_ECHO_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetEchoById(ctx context.Context, in *EchoIDInput) (EchoOutput, error) {
	echo, err := echoHandler.echoService.GetEchoById(ctx, in.ID)
	if err != nil {
		return EchoOutput{}, err
	}
	return commonModel.OK(echo, commonModel.GET_ECHO_BY_ID_SUCCESS), nil
}

func (echoHandler *EchoHandler) QueryEchos(ctx context.Context, in *QueryEchosInput) (EchoPageOutput, error) {
	result, err := echoHandler.echoService.QueryEchos(ctx, in.Body)
	if err != nil {
		return EchoPageOutput{}, err
	}
	return commonModel.OK(result, commonModel.QUERY_ECHOS_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetEchosByPageGet(ctx context.Context, in *EchoPageGetInput) (EchoPageOutput, error) {
	return echoHandler.getEchosByPage(ctx, commonModel.PageQueryDto{Page: in.Page, PageSize: in.PageSize, Search: in.Search})
}

func (echoHandler *EchoHandler) GetEchosByPagePost(ctx context.Context, in *EchoPagePostInput) (EchoPageOutput, error) {
	return echoHandler.getEchosByPage(ctx, in.Body)
}

func (echoHandler *EchoHandler) getEchosByPage(ctx context.Context, page commonModel.PageQueryDto) (EchoPageOutput, error) {
	result, err := echoHandler.echoService.GetEchosByPage(ctx, page)
	if err != nil {
		return EchoPageOutput{}, err
	}
	return commonModel.OK(result, commonModel.GET_ECHOS_BY_PAGE_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetEchosByTagId(ctx context.Context, in *GetEchosByTagIDInput) (EchoPageOutput, error) {
	result, err := echoHandler.echoService.GetEchosByTagId(ctx, in.TagID, commonModel.PageQueryDto{Page: in.Page, PageSize: in.PageSize, Search: in.Search})
	if err != nil {
		return EchoPageOutput{}, err
	}
	return commonModel.OK(result, commonModel.GET_ECHOS_BY_TAG_ID_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetTodayEchos(ctx context.Context, in *TimezoneInput) (EchoListOutput, error) {
	result, err := echoHandler.echoService.GetTodayEchos(ctx, timezoneUtil.NormalizeTimezone(in.Timezone))
	if err != nil {
		return EchoListOutput{}, err
	}
	return commonModel.OK(result, commonModel.GET_TODAY_ECHOS_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetHotEchos(ctx context.Context, in *GetHotEchosInput) (EchoListOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 5
	}
	result, err := echoHandler.echoService.GetHotEchos(ctx, limit)
	if err != nil {
		return EchoListOutput{}, err
	}
	return commonModel.OK(result, commonModel.GET_HOT_ECHOS_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetRandomEcho(ctx context.Context, _ *GetRandomEchoInput) (EchoOutput, error) {
	echo, err := echoHandler.echoService.GetRandomEcho(ctx)
	if err != nil {
		return EchoOutput{}, err
	}
	return commonModel.OK(echo, commonModel.GET_RANDOM_ECHO_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetOnThisDayEchos(ctx context.Context, in *TimezoneInput) (EchoListOutput, error) {
	result, err := echoHandler.echoService.GetOnThisDayEchos(ctx, timezoneUtil.NormalizeTimezone(in.Timezone))
	if err != nil {
		return EchoListOutput{}, err
	}
	return commonModel.OK(result, commonModel.GET_ON_THIS_DAY_ECHOS_SUCCESS), nil
}

func (echoHandler *EchoHandler) GetAllTags(ctx context.Context, _ *GetAllTagsInput) (TagListOutput, error) {
	tags, err := echoHandler.echoService.GetAllTags()
	if err != nil {
		return TagListOutput{}, err
	}
	return commonModel.OK(tags, commonModel.GET_ALL_TAGS_SUCCESS), nil
}

func (echoHandler *EchoHandler) CreateTag(ctx context.Context, in *CreateTagInput) (TagOutput, error) {
	tag, err := echoHandler.echoService.CreateTag(ctx, in.Body.Name)
	if err != nil {
		return TagOutput{}, err
	}
	return commonModel.OK(tag, commonModel.CREATE_TAG_SUCCESS), nil
}

func (echoHandler *EchoHandler) DeleteTag(ctx context.Context, in *TagIDInput) (EmptyOutput, error) {
	if err := echoHandler.echoService.DeleteTag(ctx, in.ID); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_TAG_SUCCESS), nil
}
