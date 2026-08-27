// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
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

type Bundle struct {
	WebHandler       *webHandler.WebHandler
	UserHandler      *userHandler.UserHandler
	AuthHandler      *authHandler.AuthHandler
	EchoHandler      *echoHandler.EchoHandler
	FileHandler      *fileHandler.FileHandler
	CommentHandler   *commentHandler.CommentHandler
	InitHandler      *initHandler.InitHandler
	CommonHandler    *commonHandler.CommonHandler
	SettingHandler   *settingHandler.SettingHandler
	ConnectHandler   *connectHandler.ConnectHandler
	MigrationHandler *migratorHandler.MigrationHandler
	DashboardHandler *dashboardHandler.DashboardHandler
	CopilotHandler   *copilotHandler.CopilotHandler
	EmbeddingHandler *embeddingHandler.EmbeddingHandler
	MCPHandler       *mcp.Handler
}

func NewBundle(
	webHandler *webHandler.WebHandler,
	userHandler *userHandler.UserHandler,
	authHandler *authHandler.AuthHandler,
	echoHandler *echoHandler.EchoHandler,
	fileHandler *fileHandler.FileHandler,
	commentHandler *commentHandler.CommentHandler,
	initHandler *initHandler.InitHandler,
	commonHandler *commonHandler.CommonHandler,
	settingHandler *settingHandler.SettingHandler,
	connectHandler *connectHandler.ConnectHandler,
	migratorHandler *migratorHandler.MigrationHandler,
	dashboardHandler *dashboardHandler.DashboardHandler,
	copilotHandler *copilotHandler.CopilotHandler,
	embeddingHandler *embeddingHandler.EmbeddingHandler,
	mcpHandler *mcp.Handler,
) *Bundle {
	return &Bundle{
		WebHandler:       webHandler,
		UserHandler:      userHandler,
		AuthHandler:      authHandler,
		EchoHandler:      echoHandler,
		FileHandler:      fileHandler,
		CommentHandler:   commentHandler,
		InitHandler:      initHandler,
		CommonHandler:    commonHandler,
		SettingHandler:   settingHandler,
		ConnectHandler:   connectHandler,
		MigrationHandler: migratorHandler,
		DashboardHandler: dashboardHandler,
		CopilotHandler:   copilotHandler,
		EmbeddingHandler: embeddingHandler,
		MCPHandler:       mcpHandler,
	}
}
