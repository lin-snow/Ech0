// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/handler"
)

func registerConnect(api *apiRegistry, h *handler.Bundle) {
	api.route(public("peer directory metadata this instance shares with other instances"), huma.Operation{
		OperationID: "connect-self",
		Method:      http.MethodGet,
		Path:        "/connect",
		Summary:     "获取当前实例的连接信息",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnect)

	api.route(public("peer directory listing is public by design"), huma.Operation{
		OperationID: "connect-list",
		Method:      http.MethodGet,
		Path:        "/connect/list",
		Summary:     "获取当前实例添加的所有连接",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnects)

	api.route(public("peer directory details are public by design"), huma.Operation{
		OperationID: "connect-info",
		Method:      http.MethodGet,
		Path:        "/connects/info",
		Summary:     "获取所有已添加连接的详细信息",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnectsInfo)

	api.route(scoped(authz.ScopeConnectRead), huma.Operation{
		OperationID: "connect-health",
		Method:      http.MethodGet,
		Path:        "/connects/health",
		Summary:     "获取互联健康状态",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.GetConnectsHealth)

	api.route(scoped(authz.ScopeConnectWrite), huma.Operation{
		OperationID:   "connect-add",
		Method:        http.MethodPost,
		Path:          "/connects",
		Summary:       "添加连接",
		DefaultStatus: http.StatusCreated,
		Tags:          []string{"Connect"},
	}, h.ConnectHandler.AddConnect)

	api.route(scoped(authz.ScopeConnectDelete), huma.Operation{
		OperationID: "connect-delete",
		Method:      http.MethodDelete,
		Path:        "/connects/{id}",
		Summary:     "删除连接",
		Tags:        []string{"Connect"},
	}, h.ConnectHandler.DeleteConnect)
}
