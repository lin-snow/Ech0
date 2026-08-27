// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/job"
	"github.com/lin-snow/ech0/internal/kvstore"
	authRepository "github.com/lin-snow/ech0/internal/repository/auth"
	commentRepository "github.com/lin-snow/ech0/internal/repository/comment"
	commonRepository "github.com/lin-snow/ech0/internal/repository/common"
	connectRepository "github.com/lin-snow/ech0/internal/repository/connect"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	embeddingRepository "github.com/lin-snow/ech0/internal/repository/embedding"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	initRepository "github.com/lin-snow/ech0/internal/repository/init"
	jobRepository "github.com/lin-snow/ech0/internal/repository/job"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	settingRepository "github.com/lin-snow/ech0/internal/repository/setting"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	visitorRepository "github.com/lin-snow/ech0/internal/repository/visitor"
	webhookRepository "github.com/lin-snow/ech0/internal/repository/webhook"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	initService "github.com/lin-snow/ech0/internal/service/init"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	userService "github.com/lin-snow/ech0/internal/service/user"
	webhookmodule "github.com/lin-snow/ech0/internal/webhook"
)

var (
	AuthSet = wire.NewSet(
		authRepository.NewAuthRepository,
		wire.Bind(new(authService.AuthRepo), new(*authRepository.AuthRepository)),
		wire.Bind(new(authService.TokenRevoker), new(*authRepository.AuthRepository)),
		wire.Bind(new(authService.Repository), new(*authRepository.AuthRepository)),
		wire.Bind(new(settingService.TokenRevoker), new(*authRepository.AuthRepository)),
	)
	UserSet = wire.NewSet(
		userRepository.NewUserRepository,
		wire.Bind(new(userService.Repository), new(*userRepository.UserRepository)),
	)
	EchoSet = wire.NewSet(
		echoRepository.NewEchoRepository,
		wire.Bind(new(echoService.Repository), new(*echoRepository.EchoRepository)),
		wire.Bind(new(connectService.EchoRepository), new(*echoRepository.EchoRepository)),
		wire.Bind(new(embeddingService.EchoReader), new(*echoRepository.EchoRepository)),
	)
	EmbeddingSet = wire.NewSet(
		embeddingRepository.NewEmbeddingRepository,
		wire.Bind(new(embeddingService.Repository), new(*embeddingRepository.EmbeddingRepository)),
	)
	CommonSet = wire.NewSet(
		commonRepository.NewCommonRepository,
		wire.Bind(new(commonService.CommonRepository), new(*commonRepository.CommonRepository)),
		wire.Bind(new(fileService.CommonRepository), new(*commonRepository.CommonRepository)),
	)
	FileSet = wire.NewSet(
		fileRepository.NewFileRepository,
		wire.Bind(new(fileService.FileRepository), new(*fileRepository.FileRepository)),
	)
	CommentSet = wire.NewSet(
		commentRepository.NewCommentRepository,
		wire.Bind(new(commentService.Repository), new(*commentRepository.CommentRepository)),
	)
	InitSet = wire.NewSet(
		initRepository.NewInitRepository,
		wire.Bind(new(initService.Repository), new(*initRepository.InitRepository)),
	)
	KeyValueSet = wire.NewSet(
		keyvalueRepository.NewKeyValueRepository,
		wire.Bind(new(kvstore.Backend), new(*keyvalueRepository.KeyValueRepository)),
		kvstore.NewPersistent,
		wire.Bind(new(kvstore.Store), new(*kvstore.Persistent)),
	)
	SettingSet = wire.NewSet(
		settingRepository.NewSettingRepository,
		wire.Bind(new(settingService.SettingRepository), new(*settingRepository.SettingRepository)),
	)
	ConnectSet = wire.NewSet(
		connectRepository.NewConnectRepository,
		wire.Bind(new(connectService.Repository), new(*connectRepository.ConnectRepository)),
	)
	WebhookSet = wire.NewSet(
		webhookRepository.NewWebhookRepository,
		wire.Bind(new(settingService.WebhookRepository), new(*webhookRepository.WebhookRepository)),
		wire.Bind(new(webhookmodule.WebhookStore), new(*webhookRepository.WebhookRepository)),
	)
	JobSet = wire.NewSet(
		jobRepository.NewJobRepository,
		wire.Bind(new(job.JobRepository), new(*jobRepository.JobRepository)),
	)
	VisitorSet = wire.NewSet(
		visitorRepository.NewVisitorRepository,
	)
)
