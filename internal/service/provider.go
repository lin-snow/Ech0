package service

import (
	"github.com/google/wire"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	backupService "github.com/lin-snow/ech0/internal/service/backup"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	dashboardService "github.com/lin-snow/ech0/internal/service/dashboard"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
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
	BackupSet = wire.NewSet(
		backupService.NewBackupService,
		wire.Bind(new(backupService.Service), new(*backupService.BackupService)),
	)
	DashboardSet = wire.NewSet(
		dashboardService.NewDashboardService,
		wire.Bind(new(dashboardService.Service), new(*dashboardService.DashboardService)),
	)
	AgentSet = wire.NewSet(
		agentService.NewAgentService,
		wire.Bind(new(agentService.Service), new(*agentService.AgentService)),
	)
	MigratorSet = wire.NewSet(
		migratorService.NewMigratorService,
		wire.Bind(new(migratorService.Service), new(*migratorService.MigratorService)),
	)
)
