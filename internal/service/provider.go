// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"github.com/google/wire"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	dashboardService "github.com/lin-snow/ech0/internal/service/dashboard"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	initService "github.com/lin-snow/ech0/internal/service/init"
	migratorService "github.com/lin-snow/ech0/internal/service/migrator"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

var (
	AuthSet = authService.ProviderSet
	UserSet = wire.NewSet(
		userService.NewUserService,
		wire.Bind(new(userService.Service), new(*userService.UserService)),
	)
	EchoSet = wire.NewSet(
		echoService.NewEchoService,
		wire.Bind(new(echoService.Service), new(*echoService.EchoService)),
	)
	FileSet = wire.NewSet(
		fileService.NewFileService,
		wire.Bind(new(fileService.Service), new(*fileService.FileService)),
	)
	CommentSet = wire.NewSet(
		commentService.NewGoMailSender,
		wire.Bind(new(commentService.Mailer), new(*commentService.GoMailSender)),
		commentService.NewCommentService,
		wire.Bind(new(commentService.Service), new(*commentService.CommentService)),
	)
	InitSet = wire.NewSet(
		initService.NewInitService,
		wire.Bind(new(initService.Service), new(*initService.InitService)),
	)
	CommonSet = wire.NewSet(
		commonService.NewCommonService,
		wire.Bind(new(commonService.Service), new(*commonService.CommonService)),
	)
	SettingSet = wire.NewSet(
		settingService.NewSettingService,
		wire.Bind(new(settingService.Service), new(*settingService.SettingService)),
	)
	ConnectSet = wire.NewSet(
		connectService.NewConnectService,
		wire.Bind(new(connectService.Service), new(*connectService.ConnectService)),
	)
	DashboardSet = wire.NewSet(
		dashboardService.NewDashboardService,
		wire.Bind(new(dashboardService.Service), new(*dashboardService.DashboardService)),
	)
	EmbeddingSet = wire.NewSet(
		embeddingService.NewEmbeddingService,
		wire.Bind(new(embeddingService.Service), new(*embeddingService.EmbeddingService)),
		wire.Bind(new(embeddingService.Indexer), new(*embeddingService.EmbeddingService)),
	)
	CopilotSet = wire.NewSet(
		copilotService.NewCopilotService,
		wire.Bind(new(copilotService.SummaryService), new(*copilotService.CopilotService)),
		wire.Bind(new(copilotService.ChatService), new(*copilotService.CopilotService)),
	)
	MigratorSet = wire.NewSet(
		migratorService.NewMigratorService,
		wire.Bind(new(migratorService.Service), new(*migratorService.MigratorService)),
	)
)
