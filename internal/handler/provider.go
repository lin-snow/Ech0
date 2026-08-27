// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"github.com/google/wire"
	authHandler "github.com/lin-snow/ech0/internal/handler/auth"
	commentHandler "github.com/lin-snow/ech0/internal/handler/comment"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	copilotHandler "github.com/lin-snow/ech0/internal/handler/copilot"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	embeddingHandler "github.com/lin-snow/ech0/internal/handler/embedding"
	fileHandler "github.com/lin-snow/ech0/internal/handler/file"
	initHandler "github.com/lin-snow/ech0/internal/handler/init"
	migratorHandler "github.com/lin-snow/ech0/internal/handler/migrator"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
	"github.com/lin-snow/ech0/internal/mcp"
)

var (
	WebSet       = wire.NewSet(webHandler.NewWebHandler)
	UserSet      = wire.NewSet(userHandler.NewUserHandler)
	AuthSet      = wire.NewSet(authHandler.NewAuthHandler)
	EchoSet      = wire.NewSet(echoHandler.NewEchoHandler)
	FileSet      = wire.NewSet(fileHandler.NewFileHandler)
	CommentSet   = wire.NewSet(commentHandler.NewCommentHandler)
	InitSet      = wire.NewSet(initHandler.NewInitHandler)
	CommonSet    = wire.NewSet(commonHandler.NewCommonHandler)
	SettingSet   = wire.NewSet(settingHandler.NewSettingHandler)
	ConnectSet   = wire.NewSet(connectHandler.NewConnectHandler)
	DashboardSet = wire.NewSet(dashboardHandler.NewDashboardHandler)
	CopilotSet   = wire.NewSet(copilotHandler.NewCopilotHandler)
	EmbeddingSet = wire.NewSet(embeddingHandler.NewEmbeddingHandler)
	MigrationSet = wire.NewSet(migratorHandler.NewMigrationHandler)
	MCPSet       = wire.NewSet(mcp.NewHandler)
)
