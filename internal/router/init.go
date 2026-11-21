// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
)

func registerInit(api *apiRegistry, h *handler.Bundle) {
	api.route(public(), huma.Operation{
		OperationID: "init-status",
		Method:      http.MethodGet,
		Path:        "/init/status",
		Summary:     "获取系统初始化状态",
		Description: "返回站点是否已初始化、是否已存在管理员。",
		Tags:        []string{"Init"},
	}, h.InitHandler.GetInitStatus)

	api.route(public(), huma.Operation{
		OperationID: "init-owner",
		Method:      http.MethodPost,
		Path:        "/init/owner",
		Summary:     "初始化站点管理员",
		Description: "创建首个管理员账号（仅在未初始化时可用）。",
		Tags:        []string{"Init"},
		Middlewares: huma.Middlewares{humares.Bridge(middleware.RateLimitN(5, time.Minute, 5))},
	}, h.InitHandler.InitOwner)
}
