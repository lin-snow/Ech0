// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func registerEcho(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, public(), huma.Operation{
		OperationID: "echo-like",
		Method:      http.MethodPut,
		Path:        "/echo/like/{id}",
		Summary:     "点赞 Echo",
		Tags:        []string{"Echo"},
		Middlewares: huma.Middlewares{humares.Bridge(middleware.RateLimitWithIdempotency(2, 5, time.Hour, "id", func(c *gin.Context) {
			c.JSON(http.StatusOK, commonModel.OK[any](nil, commonModel.LIKE_ECHO_SUCCESS))
		}))},
	}, h.EchoHandler.LikeEcho)

	route(api, public(), huma.Operation{
		OperationID: "tag-list",
		Method:      http.MethodGet,
		Path:        "/tags",
		Summary:     "获取所有标签",
		Tags:        []string{"Tag"},
	}, h.EchoHandler.GetAllTags)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-query",
		Method:      http.MethodPost,
		Path:        "/echo/query",
		Summary:     "统一查询 Echo 列表",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.QueryEchos)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-page-get",
		Method:      http.MethodGet,
		Path:        "/echo/page",
		Summary:     "分页获取 Echo（Deprecated）",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetEchosByPageGet)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-page-post",
		Method:      http.MethodPost,
		Path:        "/echo/page",
		Summary:     "分页获取 Echo（Deprecated）",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetEchosByPagePost)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-by-tag",
		Method:      http.MethodGet,
		Path:        "/echo/tag/{tagid}",
		Summary:     "按标签获取 Echo（Deprecated）",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetEchosByTagId)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-today",
		Method:      http.MethodGet,
		Path:        "/echo/today",
		Summary:     "获取今天的 Echo",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetTodayEchos)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-hot",
		Method:      http.MethodGet,
		Path:        "/echo/hot",
		Summary:     "获取热门 Echo",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetHotEchos)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-random",
		Method:      http.MethodGet,
		Path:        "/echo/random",
		Summary:     "随机返回一篇 Echo",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetRandomEcho)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-onthisday",
		Method:      http.MethodGet,
		Path:        "/echo/onthisday",
		Summary:     "那年今日",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetOnThisDayEchos)

	route(api, optional(revoker), huma.Operation{
		OperationID: "echo-get",
		Method:      http.MethodGet,
		Path:        "/echo/{id}",
		Summary:     "获取指定 ID 的 Echo",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.GetEchoById)

	route(api, secured(revoker, authModel.ScopeEchoWrite), huma.Operation{
		OperationID: "echo-create",
		Method:      http.MethodPost,
		Path:        "/echo",
		Summary:     "创建新的 Echo",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.PostEcho)

	route(api, secured(revoker, authModel.ScopeEchoWrite), huma.Operation{
		OperationID: "echo-update",
		Method:      http.MethodPut,
		Path:        "/echo",
		Summary:     "更新 Echo",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.UpdateEcho)

	route(api, secured(revoker, authModel.ScopeEchoWrite), huma.Operation{
		OperationID: "echo-delete",
		Method:      http.MethodDelete,
		Path:        "/echo/{id}",
		Summary:     "删除 Echo",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.DeleteEcho)

	route(api, secured(revoker, authModel.ScopeEchoWrite), huma.Operation{
		OperationID: "tag-create",
		Method:      http.MethodPost,
		Path:        "/tag",
		Summary:     "创建标签",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.CreateTag)

	route(api, secured(revoker, authModel.ScopeEchoWrite), huma.Operation{
		OperationID: "tag-delete",
		Method:      http.MethodDelete,
		Path:        "/tag/{id}",
		Summary:     "删除标签",
		Tags:        []string{"Echo"},
	}, h.EchoHandler.DeleteTag)
}
