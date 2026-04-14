package handler

import (
	"github.com/google/wire"
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	authHandler "github.com/lin-snow/ech0/internal/handler/auth"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commentHandler "github.com/lin-snow/ech0/internal/handler/comment"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	fileHandler "github.com/lin-snow/ech0/internal/handler/file"
	initHandler "github.com/lin-snow/ech0/internal/handler/init"
	migrationHandler "github.com/lin-snow/ech0/internal/handler/migration"
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
	BackupSet    = wire.NewSet(backupHandler.NewBackupHandler)
	DashboardSet = wire.NewSet(dashboardHandler.NewDashboardHandler)
	AgentSet     = wire.NewSet(agentHandler.NewAgentHandler)
	MigrationSet = wire.NewSet(migrationHandler.NewMigrationHandler)
	MCPSet       = wire.NewSet(mcp.NewHandler)
)
