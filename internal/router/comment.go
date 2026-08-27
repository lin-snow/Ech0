// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/captcha"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/handler/humares"
	"github.com/lin-snow/ech0/internal/middleware"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

func setupCommentRoutes(appRouterGroup *AppRouterGroup, _ *handler.Bundle) {
	captchaHandler, err := captcha.NewHTTPHandler("/api")
	if err != nil {
		panic(err)
	}
	appRouterGroup.PublicRouterGroup.Any("/cap/*any", gin.WrapH(captchaHandler))
}

func registerComment(api huma.API, h *handler.Bundle, revoker authService.TokenRevoker) {
	nc := humares.Bridge(middleware.NoCache())
	stash := humares.Bridge(h.CommentHandler.StashMeta())
	optViewer := humares.Bridge(h.CommentHandler.OptionalViewer())
	moderate := secured(revoker, authModel.ScopeCommentMod)

	route(api, public(), huma.Operation{
		OperationID: "comment-form-meta",
		Method:      http.MethodGet,
		Path:        "/comments/form",
		Summary:     "获取评论表单元信息",
		Tags:        []string{"Comment"},
		Middlewares: huma.Middlewares{nc, stash, optViewer},
	}, h.CommentHandler.GetFormMeta)

	route(api, public(), huma.Operation{
		OperationID: "comment-list-by-echo",
		Method:      http.MethodGet,
		Path:        "/comments",
		Summary:     "按 echo 列出公开评论",
		Tags:        []string{"Comment"},
		Middlewares: huma.Middlewares{nc},
	}, h.CommentHandler.ListCommentsByEchoID)

	route(api, public(), huma.Operation{
		OperationID: "comment-list-public",
		Method:      http.MethodGet,
		Path:        "/comments/public",
		Summary:     "列出最新公开评论",
		Tags:        []string{"Comment"},
		Middlewares: huma.Middlewares{nc},
	}, h.CommentHandler.ListPublicComments)

	route(api, public(), huma.Operation{
		OperationID: "comment-create",
		Method:      http.MethodPost,
		Path:        "/comments",
		Summary:     "创建公开评论",
		Tags:        []string{"Comment"},
		Middlewares: huma.Middlewares{stash, optViewer},
	}, h.CommentHandler.CreateComment)

	route(api, secured(revoker, authModel.ScopeCommentWrite).audience(authModel.AudienceIntegration, authModel.AudienceMCPRemote), huma.Operation{
		OperationID: "comment-create-integration",
		Method:      http.MethodPost,
		Path:        "/comments/integration",
		Summary:     "经访问令牌创建评论（集成）",
		Tags:        []string{"Comment"},
		Middlewares: huma.Middlewares{stash},
	}, h.CommentHandler.CreateIntegrationComment)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-list",
		Method:      http.MethodGet,
		Path:        "/panel/comments",
		Summary:     "管理面板列出评论",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.ListPanelComments)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-get",
		Method:      http.MethodGet,
		Path:        "/panel/comments/{id}",
		Summary:     "获取单条评论",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.GetCommentByID)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-status",
		Method:      http.MethodPatch,
		Path:        "/panel/comments/{id}/status",
		Summary:     "更新评论审核状态",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.UpdateCommentStatus)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-hot",
		Method:      http.MethodPatch,
		Path:        "/panel/comments/{id}/hot",
		Summary:     "置顶/取消置顶评论",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.UpdateCommentHot)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-delete",
		Method:      http.MethodDelete,
		Path:        "/panel/comments/{id}",
		Summary:     "删除评论",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.DeleteComment)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-batch",
		Method:      http.MethodPost,
		Path:        "/panel/comments/batch",
		Summary:     "批量操作评论",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.BatchAction)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-settings-get",
		Method:      http.MethodGet,
		Path:        "/panel/comments/settings",
		Summary:     "获取评论系统设置",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.GetCommentSetting)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-settings-update",
		Method:      http.MethodPut,
		Path:        "/panel/comments/settings",
		Summary:     "更新评论系统设置",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.UpdateCommentSetting)

	route(api, moderate, huma.Operation{
		OperationID: "comment-panel-test-email",
		Method:      http.MethodPost,
		Path:        "/panel/comments/settings/test-email",
		Summary:     "发送评论通知测试邮件",
		Tags:        []string{"Comment"},
	}, h.CommentHandler.TestCommentEmail)
}
