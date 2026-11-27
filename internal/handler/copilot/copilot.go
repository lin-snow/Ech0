// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	res "github.com/lin-snow/ech0/internal/handler/response"
	i18n "github.com/lin-snow/ech0/internal/i18n"
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

// AnswerAskInput carries a person's reply to the round a run is parked on.
//
// AskID is required and is checked against the outstanding round rather than
// trusted: a stale tab holds a picker for a question that has already been
// answered, and its click has to be refused rather than allowed to overwrite
// the answer that was given.
type AnswerAskInput struct {
	Body struct {
		AskID   string                     `json:"ask_id" minLength:"1" doc:"要回答的提问轮次 ID，来自 SSE 的 ask 事件"`
		Answers []copilotService.AskAnswer `json:"answers" minItems:"1" doc:"每个问题的回答，按问题 ID 对应"`
	}
}

type (
	RecentOutput  = struct{ Body string }
	SessionOutput = struct{ Body []copilotService.ChatMessage }
	EmptyOutput   = struct{}
)

func (h *CopilotHandler) GetRecent(ctx context.Context, _ *GetRecentInput) (*RecentOutput, error) {
	gen, err := h.summaryService.GetRecent(ctx)
	if err != nil {
		return nil, err
	}
	return &RecentOutput{Body: gen}, nil
}

func (h *CopilotHandler) GetSession(ctx context.Context, _ *GetSessionInput) (*SessionOutput, error) {
	session, err := h.chatService.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	return &SessionOutput{Body: session}, nil
}

func (h *CopilotHandler) ClearSession(ctx context.Context, _ *ClearSessionInput) (*EmptyOutput, error) {
	if err := h.chatService.ClearSession(ctx); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

func (h *CopilotHandler) AnswerAsk(ctx context.Context, input *AnswerAskInput) (*EmptyOutput, error) {
	if err := h.chatService.AnswerAsk(ctx, input.Body.AskID, input.Body.Answers); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
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
		err := h.chatService.AskStream(ctx.Request.Context(), req.Question, locale, timezone, ctx.Writer)
		if err != nil && !ctx.Writer.Written() {
			res.Execute(func(*gin.Context) res.Response { return res.Failed(err) })(ctx)
		}
	}
}
