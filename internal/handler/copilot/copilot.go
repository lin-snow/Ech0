// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	i18n "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
)

type CopilotHandler struct {
	summaryService copilotService.SummaryService
	chatService    copilotService.ChatService
}

func NewCopilotHandler(
	summaryService copilotService.SummaryService,
	chatService copilotService.ChatService,
) *CopilotHandler {
	return &CopilotHandler{
		summaryService: summaryService,
		chatService:    chatService,
	}
}

type (
	GetRecentInput    struct{}
	GetSessionInput   struct{}
	ClearSessionInput struct{}
)

type (
	RecentOutput  = commonModel.Result[string]
	SessionOutput = commonModel.Result[[]copilotService.ChatMessage]
	EmptyOutput   = commonModel.Result[any]
)

func (h *CopilotHandler) GetRecent(ctx context.Context, _ *GetRecentInput) (RecentOutput, error) {
	gen, err := h.summaryService.GetRecent(ctx)
	if err != nil {
		return RecentOutput{}, err
	}
	return commonModel.OK(gen, commonModel.AGENT_GET_RECENT_SUCCESS), nil
}

func (h *CopilotHandler) GetSession(ctx context.Context, _ *GetSessionInput) (SessionOutput, error) {
	session, err := h.chatService.GetSession(ctx)
	if err != nil {
		return SessionOutput{}, err
	}
	return commonModel.OK(session, commonModel.CHAT_SESSION_GET_SUCCESS), nil
}

func (h *CopilotHandler) ClearSession(ctx context.Context, _ *ClearSessionInput) (EmptyOutput, error) {
	if err := h.chatService.ClearSession(ctx); err != nil {
		return EmptyOutput{}, err
	}
	return commonModel.OK[any](nil, commonModel.CHAT_SESSION_CLEAR_SUCCESS), nil
}

type askRequest struct {
	Question string `json:"question"`
}

func (h *CopilotHandler) Ask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req askRequest
		_ = ctx.ShouldBindJSON(&req)
		locale := i18n.LocaleFromGin(ctx)
		timezone := timezoneUtil.NormalizeTimezone(ctx.GetHeader(timezoneUtil.DefaultTimezoneHeader))
		_ = h.chatService.AskStream(ctx.Request.Context(), req.Question, locale, timezone, ctx.Writer)
	}
}
