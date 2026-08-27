// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/comment"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	commonService "github.com/lin-snow/ech0/internal/service/common"
)

type Service interface {
	GetFormMeta(ctx context.Context, clientIP, apiBaseURL string) (model.FormMeta, error)
	CreateComment(
		ctx context.Context,
		clientIP,
		userAgent string,
		dto *model.CreateCommentDto,
	) (model.CreateCommentResult, error)
	CreateIntegrationComment(
		ctx context.Context,
		clientIP,
		userAgent string,
		dto *model.CreateIntegrationCommentDto,
	) (model.CreateCommentResult, error)
	ListPublicByEchoID(ctx context.Context, echoID string) ([]model.PublicComment, error)
	ListPublicComments(ctx context.Context, limit int) ([]model.PublicComment, error)
	ListPanelComments(ctx context.Context, query model.ListCommentQuery) (model.PageResult[model.Comment], error)
	GetCommentByID(ctx context.Context, id string) (model.Comment, error)
	UpdateCommentStatus(ctx context.Context, id string, status model.Status) error
	UpdateCommentHot(ctx context.Context, id string, hot bool) error
	DeleteComment(ctx context.Context, id string) error
	BatchAction(ctx context.Context, action string, ids []string) error
	GetSystemSetting(ctx context.Context) (model.SystemSetting, error)
	UpdateSystemSetting(ctx context.Context, setting model.SystemSetting) error
	SendTestEmail(ctx context.Context, setting model.SystemSetting) error
}

type Repository interface {
	CreateComment(ctx context.Context, c *model.Comment) error
	ListPublicByEchoID(ctx context.Context, echoID string) ([]model.Comment, error)
	ListPublicComments(ctx context.Context, limit int) ([]model.Comment, error)
	ListComments(ctx context.Context, query model.ListCommentQuery) (model.PageResult[model.Comment], error)
	GetCommentByID(ctx context.Context, id string) (model.Comment, error)
	UpdateCommentStatus(ctx context.Context, id string, status model.Status) error
	UpdateCommentHot(ctx context.Context, id string, hot bool) error
	DeleteComment(ctx context.Context, id string) error
	BatchUpdateStatus(ctx context.Context, ids []string, status model.Status) error
	BatchDelete(ctx context.Context, ids []string) error
	CountByIPWithin(ctx context.Context, ipHash string, seconds int64) (int64, error)
	CountByEmailWithin(ctx context.Context, email string, seconds int64) (int64, error)
	CountByUserWithin(ctx context.Context, userID string, seconds int64) (int64, error)
	ExistsRecentDuplicate(
		ctx context.Context,
		echoID, content, email, ipHash, userID string,
		seconds int64,
	) (bool, error)
}

type CommonService = commonService.Service

type UserContext struct {
	User  userModel.User
	Valid bool
}

type MailMessage struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
}

type MailerConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

type Mailer interface {
	Send(ctx context.Context, cfg MailerConfig, msg MailMessage) error
}
