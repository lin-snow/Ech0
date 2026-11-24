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
	FormMetaOutput       = struct{ Body model.FormMeta }
	PublicCommentsOutput = struct {
		Body commonModel.List[model.PublicComment]
	}
	CreateCommentOutput = struct{ Body model.Comment }
	PanelCommentsOutput = struct {
		Body commonModel.List[model.Comment]
	}
	CommentOutput        = struct{ Body model.Comment }
	CommentSettingOutput = struct{ Body model.SystemSetting }
	EmptyOutput          = struct{}
)

func (h *CommentHandler) GetFormMeta(ctx context.Context, _ *GetFormMetaInput) (*FormMetaOutput, error) {
	m := metaFrom(ctx)
	data, err := h.commentService.GetFormMeta(ctx, m.clientIP, m.baseURL)
	if err != nil {
		return nil, err
	}
	return &FormMetaOutput{Body: data}, nil
}

func (h *CommentHandler) ListCommentsByEchoID(ctx context.Context, in *ListCommentsByEchoInput) (*PublicCommentsOutput, error) {
	comments, err := h.commentService.ListPublicByEchoID(ctx, strings.TrimSpace(in.EchoID))
	if err != nil {
		return nil, err
	}
	return &PublicCommentsOutput{Body: commonModel.List[model.PublicComment]{Items: comments}}, nil
}

func (h *CommentHandler) ListPublicComments(ctx context.Context, in *ListPublicCommentsInput) (*PublicCommentsOutput, error) {
	comments, err := h.commentService.ListPublicComments(ctx, in.Limit)
	if err != nil {
		return nil, err
	}
	return &PublicCommentsOutput{Body: commonModel.List[model.PublicComment]{Items: comments}}, nil
}

func (h *CommentHandler) CreateComment(ctx context.Context, in *CreateCommentInput) (*CreateCommentOutput, error) {
	m := metaFrom(ctx)
	comment, err := h.commentService.CreateComment(ctx, m.clientIP, m.userAgent, &in.Body)
	if err != nil {
		return nil, err
	}
	return &CreateCommentOutput{Body: *comment}, nil
}

func (h *CommentHandler) CreateIntegrationComment(ctx context.Context, in *CreateIntegrationCommentInput) (*CreateCommentOutput, error) {
	m := metaFrom(ctx)
	comment, err := h.commentService.CreateIntegrationComment(ctx, m.clientIP, m.userAgent, &in.Body)
	if err != nil {
		return nil, err
	}
	return &CreateCommentOutput{Body: *comment}, nil
}

func (h *CommentHandler) ListPanelComments(ctx context.Context, in *ListPanelCommentsInput) (*PanelCommentsOutput, error) {
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
		return nil, err
	}
	return &PanelCommentsOutput{Body: data}, nil
}

func (h *CommentHandler) GetCommentByID(ctx context.Context, in *GetCommentByIDInput) (*CommentOutput, error) {
	data, err := h.commentService.GetCommentByID(ctx, strings.TrimSpace(in.ID))
	if err != nil {
		return nil, err
	}
	return &CommentOutput{Body: data}, nil
}

func (h *CommentHandler) UpdateCommentStatus(ctx context.Context, in *UpdateCommentStatusInput) (*EmptyOutput, error) {
	if err := h.commentService.UpdateCommentStatus(ctx, strings.TrimSpace(in.ID), in.Body.Status); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *CommentHandler) UpdateCommentHot(ctx context.Context, in *UpdateCommentHotInput) (*EmptyOutput, error) {
	if err := h.commentService.UpdateCommentHot(ctx, strings.TrimSpace(in.ID), in.Body.Hot); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *CommentHandler) DeleteComment(ctx context.Context, in *DeleteCommentInput) (*EmptyOutput, error) {
	if err := h.commentService.DeleteComment(ctx, strings.TrimSpace(in.ID)); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *CommentHandler) BatchAction(ctx context.Context, in *BatchActionInput) (*EmptyOutput, error) {
	if err := h.commentService.BatchAction(ctx, in.Body.Action, in.Body.IDs); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *CommentHandler) GetCommentSetting(ctx context.Context, _ *GetCommentSettingInput) (*CommentSettingOutput, error) {
	data, err := h.commentService.GetSystemSettingForAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &CommentSettingOutput{Body: data}, nil
}

func (h *CommentHandler) UpdateCommentSetting(ctx context.Context, in *UpdateCommentSettingInput) (*EmptyOutput, error) {
	if err := h.commentService.UpdateSystemSetting(ctx, in.Body); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *CommentHandler) TestCommentEmail(ctx context.Context, in *TestCommentEmailInput) (*EmptyOutput, error) {
	if err := h.commentService.SendTestEmail(ctx, in.Body.Setting); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
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
