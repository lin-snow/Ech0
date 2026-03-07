package service

import (
	"github.com/google/wire"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
	backupService "github.com/lin-snow/ech0/internal/service/backup"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	dashboardService "github.com/lin-snow/ech0/internal/service/dashboard"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	inboxService "github.com/lin-snow/ech0/internal/service/inbox"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

var (
	UserSet      = wire.NewSet(userService.NewUserService)
	EchoSet      = wire.NewSet(echoService.NewEchoService)
	CommonSet    = wire.NewSet(commonService.NewCommonService)
	SettingSet   = wire.NewSet(settingService.NewSettingService)
	TodoSet      = wire.NewSet(todoService.NewTodoService)
	ConnectSet   = wire.NewSet(connectService.NewConnectService)
	BackupSet    = wire.NewSet(backupService.NewBackupService)
	DashboardSet = wire.NewSet(dashboardService.NewDashboardService)
	AgentSet     = wire.NewSet(agentService.NewAgentService)
	InboxSet     = wire.NewSet(inboxService.NewInboxService)
)
