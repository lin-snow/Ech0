// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

func setupCopilotRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.AuthRouterGroup.POST(
		"/chat",
		middleware.RequireScopes(authz.ScopeAdminAgent),
		h.CopilotHandler.Ask(),
	)
}

func registerCopilot(api *apiRegistry, h *handler.Bundle) {
	api.route(public("public digest of the author's recent posts"), huma.Operation{
		OperationID: "copilot-recent",
		Method:      http.MethodGet,
		Path:        "/agent/recent",
		Summary:     "获取作者近况的 AI 总结",
		Tags:        []string{"Copilot"},
	}, h.CopilotHandler.GetRecent)

	api.route(scoped(authz.ScopeAdminAgent), huma.Operation{
		OperationID: "copilot-session-get",
		Method:      http.MethodGet,
		Path:        "/chat/session",
		Summary:     "获取持久化 Chat 会话",
		Tags:        []string{"Copilot"},
	}, h.CopilotHandler.GetSession)

	api.route(scoped(authz.ScopeAdminAgent), huma.Operation{
		OperationID: "copilot-session-clear",
		Method:      http.MethodDelete,
		Path:        "/chat/session",
		Summary:     "清除持久化 Chat 会话",
		Tags:        []string{"Copilot"},
	}, h.CopilotHandler.ClearSession)

	api.route(scoped(authz.ScopeAdminAgent), huma.Operation{
		OperationID: "copilot-chat-answer",
		Method:      http.MethodPost,
		Path:        "/chat/answer",
		Summary:     "回答 Copilot 正在等待的提问/确认",
		Tags:        []string{"Copilot"},
	}, h.CopilotHandler.AnswerAsk)
}
