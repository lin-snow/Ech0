// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	model "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/comment"
	"github.com/lin-snow/ech0/pkg/viewer"
)

type CommentHandler struct {
	commentService service.Service
}

func NewCommentHandler(commentService service.Service) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

type commentMeta struct {
	clientIP  string
	userAgent string
	baseURL   string
}

type commentMetaKey struct{}

func metaFrom(ctx context.Context) commentMeta {
	m, _ := ctx.Value(commentMetaKey{}).(commentMeta)
	return m
}

func (h *CommentHandler) StashMeta() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		m := commentMeta{
			clientIP:  ctx.ClientIP(),
			userAgent: ctx.Request.UserAgent(),
			baseURL:   resolveRequestBaseURL(ctx.Request),
		}
		ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), commentMetaKey{}, m))
		ctx.Next()
	}
}

func (h *CommentHandler) OptionalViewer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h.attachOptionalViewer(ctx)
		ctx.Next()
	}
}

type (
	GetFormMetaInput        struct{}
	ListCommentsByEchoInput struct {
		EchoID string `query:"echo_id" required:"true" doc:"Echo ID"`
	}
	ListPublicCommentsInput struct {
		Limit int `query:"limit" default:"30" doc:"返回数量上限"`
	}
	CreateCommentInput struct {
		Body model.CreateCommentDto
	}
	CreateIntegrationCommentInput struct {
		Body model.CreateIntegrationCommentDto
	}
	ListPanelCommentsInput struct {
		Page     int    `query:"page" default:"1"`
		PageSize int    `query:"page_size" default:"20"`
		Keyword  string `query:"keyword"`
		Status   string `query:"status"`
		EchoID   string `query:"echo_id"`
		Hot      string `query:"hot" doc:"true/false 过滤置顶；缺省不过滤"`
	}
	GetCommentByIDInput struct {
		ID string `path:"id" doc:"评论 ID"`
	}
	UpdateCommentStatusInput struct {
		ID   string `path:"id" doc:"评论 ID"`
		Body model.UpdateCommentStatusDto
	}
	UpdateCommentHotInput struct {
		ID   string `path:"id" doc:"评论 ID"`
		Body model.UpdateCommentHotDto
	}
	DeleteCommentInput struct {
		ID string `path:"id" doc:"评论 ID"`
	}
	BatchActionInput struct {
		Body model.BatchCommentActionDto
	}
	GetCommentSettingInput    struct{}
	UpdateCommentSettingInput struct {
		Body model.SystemSetting
	}
	TestCommentEmailInput struct {
		Body model.TestEmailRequest
	}
)

type (
	FormMetaOutput       = commonModel.Result[model.FormMeta]
	PublicCommentsOutput = commonModel.Result[[]model.PublicComment]
	CreateCommentOutput  = commonModel.Result[model.CreateCommentResult]
	PanelCommentsOutput  = commonModel.Result[model.PageResult[model.Comment]]
	CommentOutput        = commonModel.Result[model.Comment]
	CommentSettingOutput = commonModel.Result[model.SystemSetting]
	EmptyOutput          = commonModel.Result[any]
)

func (h *CommentHandler) GetFormMeta(ctx context.Context, _ *GetFormMetaInput) (FormMetaOutput, error) {
	m := metaFrom(ctx)
	data, err := h.commentService.GetFormMeta(ctx, m.clientIP, m.baseURL)
	if err != nil {
		return FormMetaOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *CommentHandler) ListCommentsByEchoID(ctx context.Context, in *ListCommentsByEchoInput) (PublicCommentsOutput, error) {
	comments, err := h.commentService.ListPublicByEchoID(ctx, strings.TrimSpace(in.EchoID))
	if err != nil {
		return PublicCommentsOutput{}, err
	}
	return commonModel.OK(comments), nil
}

func (h *CommentHandler) ListPublicComments(ctx context.Context, in *ListPublicCommentsInput) (PublicCommentsOutput, error) {
	comments, err := h.commentService.ListPublicComments(ctx, in.Limit)
	if err != nil {
		return PublicCommentsOutput{}, err
	}
	return commonModel.OK(comments), nil
}

func (h *CommentHandler) CreateComment(ctx context.Context, in *CreateCommentInput) (CreateCommentOutput, error) {
	m := metaFrom(ctx)
	result, err := h.commentService.CreateComment(ctx, m.clientIP, m.userAgent, &in.Body)
	if err != nil {
		return CreateCommentOutput{}, err
	}
	return commonModel.OK(result), nil
}

func (h *CommentHandler) CreateIntegrationComment(ctx context.Context, in *CreateIntegrationCommentInput) (CreateCommentOutput, error) {
	m := metaFrom(ctx)
	result, err := h.commentService.CreateIntegrationComment(ctx, m.clientIP, m.userAgent, &in.Body)
	if err != nil {
		return CreateCommentOutput{}, err
	}
	return commonModel.OK(result), nil
}

func (h *CommentHandler) ListPanelComments(ctx context.Context, in *ListPanelCommentsInput) (PanelCommentsOutput, error) {
	var hot *bool
	if raw := strings.TrimSpace(in.Hot); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			hot = &v
		}
	}
	data, err := h.commentService.ListPanelComments(ctx, model.ListCommentQuery{
		Page:     in.Page,
		PageSize: in.PageSize,
		Keyword:  in.Keyword,
		Status:   in.Status,
		EchoID:   in.EchoID,
		Hot:      hot,
	})
	if err != nil {
		return PanelCommentsOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *CommentHandler) GetCommentByID(ctx context.Context, in *GetCommentByIDInput) (CommentOutput, error) {
	data, err := h.commentService.GetCommentByID(ctx, strings.TrimSpace(in.ID))
	if err != nil {
		return CommentOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *CommentHandler) UpdateCommentStatus(ctx context.Context, in *UpdateCommentStatusInput) (EmptyOutput, error) {
	if err := h.commentService.UpdateCommentStatus(ctx, strings.TrimSpace(in.ID), in.Body.Status); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil), nil
}

func (h *CommentHandler) UpdateCommentHot(ctx context.Context, in *UpdateCommentHotInput) (EmptyOutput, error) {
	if err := h.commentService.UpdateCommentHot(ctx, strings.TrimSpace(in.ID), in.Body.Hot); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil), nil
}

func (h *CommentHandler) DeleteComment(ctx context.Context, in *DeleteCommentInput) (EmptyOutput, error) {
	if err := h.commentService.DeleteComment(ctx, strings.TrimSpace(in.ID)); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.DELETE_SUCCESS), nil
}

func (h *CommentHandler) BatchAction(ctx context.Context, in *BatchActionInput) (EmptyOutput, error) {
	if err := h.commentService.BatchAction(ctx, in.Body.Action, in.Body.IDs); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil), nil
}

func (h *CommentHandler) GetCommentSetting(ctx context.Context, _ *GetCommentSettingInput) (CommentSettingOutput, error) {
	data, err := h.commentService.GetSystemSetting(ctx)
	if err != nil {
		return CommentSettingOutput{}, err
	}
	return commonModel.OK(data), nil
}

func (h *CommentHandler) UpdateCommentSetting(ctx context.Context, in *UpdateCommentSettingInput) (EmptyOutput, error) {
	if err := h.commentService.UpdateSystemSetting(ctx, in.Body); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.UPDATE_SETTINGS_SUCCESS), nil
}

func (h *CommentHandler) TestCommentEmail(ctx context.Context, in *TestCommentEmailInput) (EmptyOutput, error) {
	if err := h.commentService.SendTestEmail(ctx, in.Body.Setting); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil), nil
}

func (h *CommentHandler) attachOptionalViewer(ctx *gin.Context) {
	auth := strings.TrimSpace(ctx.GetHeader("Authorization"))
	userID := service.ParseOptionalUserIDFromAuthHeader(auth)
	if userID != "" {
		viewer.AttachToRequest(&ctx.Request, viewer.NewUserViewer(userID))
		return
	}
	viewer.AttachToRequest(&ctx.Request, viewer.NewNoopViewer())
}

func resolveRequestBaseURL(r *http.Request) string {
	if r == nil {
		return ""
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return ""
	}
	scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + host
}
