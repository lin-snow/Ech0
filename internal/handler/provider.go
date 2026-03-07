package handler

import (
	"github.com/google/wire"
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

var (
	WebSet       = wire.NewSet(webHandler.NewWebHandler)
	UserSet      = wire.NewSet(userHandler.NewUserHandler)
	EchoSet      = wire.NewSet(echoHandler.NewEchoHandler)
	CommonSet    = wire.NewSet(commonHandler.NewCommonHandler)
	SettingSet   = wire.NewSet(settingHandler.NewSettingHandler)
	TodoSet      = wire.NewSet(todoHandler.NewTodoHandler)
	ConnectSet   = wire.NewSet(connectHandler.NewConnectHandler)
	BackupSet    = wire.NewSet(backupHandler.NewBackupHandler)
	DashboardSet = wire.NewSet(dashboardHandler.NewDashboardHandler)
	AgentSet     = wire.NewSet(agentHandler.NewAgentHandler)
	InboxSet     = wire.NewSet(inboxHandler.NewInboxHandler)
)
