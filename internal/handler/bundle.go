package handler

import (
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

// Bundle 聚合各业务 Handler。
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
	BackupHandler    *backupHandler.BackupHandler
	MigrationHandler *migrationHandler.MigrationHandler
	DashboardHandler *dashboardHandler.DashboardHandler
	AgentHandler     *agentHandler.AgentHandler
	MCPHandler       *mcp.Handler
}

// NewBundle 创建 Handler 聚合实例。
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
	backupHandler *backupHandler.BackupHandler,
	migrationHandler *migrationHandler.MigrationHandler,
	dashboardHandler *dashboardHandler.DashboardHandler,
	agentHandler *agentHandler.AgentHandler,
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
		BackupHandler:    backupHandler,
		MigrationHandler: migrationHandler,
		DashboardHandler: dashboardHandler,
		AgentHandler:     agentHandler,
		MCPHandler:       mcpHandler,
	}
}
