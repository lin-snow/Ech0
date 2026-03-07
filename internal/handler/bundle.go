package handler

import (
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	inboxHandler "github.com/lin-snow/ech0/internal/handler/inbox"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	todoHandler "github.com/lin-snow/ech0/internal/handler/todo"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
)

// Bundle 聚合各业务 Handler。
type Bundle struct {
	WebHandler       *webHandler.WebHandler
	UserHandler      *userHandler.UserHandler
	EchoHandler      *echoHandler.EchoHandler
	CommonHandler    *commonHandler.CommonHandler
	SettingHandler   *settingHandler.SettingHandler
	InboxHandler     *inboxHandler.InboxHandler
	TodoHandler      *todoHandler.TodoHandler
	ConnectHandler   *connectHandler.ConnectHandler
	BackupHandler    *backupHandler.BackupHandler
	DashboardHandler *dashboardHandler.DashboardHandler
	AgentHandler     *agentHandler.AgentHandler
}

// NewBundle 创建 Handler 聚合实例。
func NewBundle(
	webHandler *webHandler.WebHandler,
	userHandler *userHandler.UserHandler,
	echoHandler *echoHandler.EchoHandler,
	commonHandler *commonHandler.CommonHandler,
	settingHandler *settingHandler.SettingHandler,
	inboxHandler *inboxHandler.InboxHandler,
	todoHandler *todoHandler.TodoHandler,
	connectHandler *connectHandler.ConnectHandler,
	backupHandler *backupHandler.BackupHandler,
	dashboardHandler *dashboardHandler.DashboardHandler,
	agentHandler *agentHandler.AgentHandler,
) *Bundle {
	return &Bundle{
		WebHandler:       webHandler,
		UserHandler:      userHandler,
		EchoHandler:      echoHandler,
		CommonHandler:    commonHandler,
		SettingHandler:   settingHandler,
		InboxHandler:     inboxHandler,
		TodoHandler:      todoHandler,
		ConnectHandler:   connectHandler,
		BackupHandler:    backupHandler,
		DashboardHandler: dashboardHandler,
		AgentHandler:     agentHandler,
	}
}
