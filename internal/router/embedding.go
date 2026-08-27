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

func registerEmbedding(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	route(api, secured(revoker, authModel.ScopeAdminSettings), huma.Operation{
		OperationID: "embedding-reindex",
		Method:      http.MethodPost,
		Path:        "/embedding/reindex",
		Summary:     "触发全量向量索引重建",
		Description: "提交一次全量向量索引回填作业，起即返回（异步）。",
		Tags:        []string{"Embedding"},
	}, h.EmbeddingHandler.Reindex)

	route(api, secured(revoker, authModel.ScopeAdminSettings), huma.Operation{
		OperationID: "embedding-reindex-status",
		Method:      http.MethodGet,
		Path:        "/embedding/reindex/status",
		Summary:     "查询重建索引作业状态",
		Description: "前端按类型轮询；查无作业行时返回 status=idle。",
		Tags:        []string{"Embedding"},
	}, h.EmbeddingHandler.ReindexStatus)

	route(api, secured(revoker, authModel.ScopeAdminSettings), huma.Operation{
		OperationID: "embedding-reindex-cancel",
		Method:      http.MethodPost,
		Path:        "/embedding/reindex/cancel",
		Summary:     "取消进行中的重建索引作业",
		Description: "取消后返回最新状态（轮询收敛到 cancelled）。",
		Tags:        []string{"Embedding"},
	}, h.EmbeddingHandler.CancelReindex)
}
