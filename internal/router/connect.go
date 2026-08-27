// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func registerConnect(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, public(), huma.Operation{
		OperationID: "connect-self",
		Method:      http.MethodGet,
		Path:        "/connect",
		Summary:     "获取当前实例的连接信息",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnect)

	route(api, public(), huma.Operation{
		OperationID: "connect-list",
		Method:      http.MethodGet,
		Path:        "/connect/list",
		Summary:     "获取当前实例添加的所有连接",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnects)

	route(api, public(), huma.Operation{
		OperationID: "connect-info",
		Method:      http.MethodGet,
		Path:        "/connects/info",
		Summary:     "获取所有已添加连接的详细信息",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnectsInfo)

	route(api, secured(revoker, authModel.ScopeConnectRead), huma.Operation{
		OperationID: "connect-health",
		Method:      http.MethodGet,
		Path:        "/connects/health",
		Summary:     "获取互联健康状态",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnectsHealth)

	route(api, secured(revoker, authModel.ScopeConnectWrite), huma.Operation{
		OperationID: "connect-add",
		Method:      http.MethodPost,
		Path:        "/connects",
		Summary:     "添加连接",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.AddConnect)

	route(api, secured(revoker, authModel.ScopeConnectWrite), huma.Operation{
		OperationID: "connect-delete",
		Method:      http.MethodDelete,
		Path:        "/connects/{id}",
		Summary:     "删除连接",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.DeleteConnect)
}
