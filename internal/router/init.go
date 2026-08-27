// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/handler"
)

func registerInit(api huma.API, h *handler.Bundle) {
	route(api, public(), huma.Operation{
		OperationID: "init-status",
		Method:      http.MethodGet,
		Path:        "/init/status",
		Summary:     "获取系统初始化状态",
		Description: "返回站点是否已初始化、是否已存在 Owner。",
		Tags:        []string{"Init"},
	}, h.InitHandler.GetInitStatus)

	route(api, public(), huma.Operation{
		OperationID: "init-owner",
		Method:      http.MethodPost,
		Path:        "/init/owner",
		Summary:     "初始化站点 Owner",
		Description: "创建首个 Owner 账号（仅在未初始化时可用）。",
		Tags:        []string{"Init"},
	}, h.InitHandler.InitOwner)
}
