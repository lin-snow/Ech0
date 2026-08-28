// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"net/http"

	userModel "github.com/lin-snow/ech0/internal/model/user"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
)

type SummaryService interface {
	GetRecent(ctx context.Context) (string, error)
}

type ChatService interface {
	AskStream(ctx context.Context, question string, locale string, timezone string, w http.ResponseWriter) error
	GetSession(ctx context.Context) ([]ChatMessage, error)
	ClearSession(ctx context.Context) error
	AnswerAsk(ctx context.Context, askID string, answers []AskAnswer) error
}

type (
	EchoService      = echoService.Service
	EmbeddingService = embeddingService.Service
)

type UserReader interface {
	GetUserByID(userID string) (userModel.User, error)
}
