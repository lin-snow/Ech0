package repository

import (
	"github.com/google/wire"
	eventsubscriber "github.com/lin-snow/ech0/internal/event/subscriber"
	commentRepository "github.com/lin-snow/ech0/internal/repository/comment"
	commonRepository "github.com/lin-snow/ech0/internal/repository/common"
	connectRepository "github.com/lin-snow/ech0/internal/repository/connect"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	initRepository "github.com/lin-snow/ech0/internal/repository/init"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	migrationRepository "github.com/lin-snow/ech0/internal/repository/migration"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	settingRepository "github.com/lin-snow/ech0/internal/repository/setting"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	webhookRepository "github.com/lin-snow/ech0/internal/repository/webhook"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
	authService "github.com/lin-snow/ech0/internal/service/auth"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	initService "github.com/lin-snow/ech0/internal/service/init"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	userService "github.com/lin-snow/ech0/internal/service/user"
	webhookmodule "github.com/lin-snow/ech0/internal/webhook"
)

var (
	AuthSet = wire.NewSet(
		NewAuthRepository,
		wire.Bind(new(authService.AuthRepo), new(*AuthRepository)),
		wire.Bind(new(authService.TokenRevoker), new(*AuthRepository)),
		wire.Bind(new(authService.Repository), new(*AuthRepository)),
	)
	UserSet = wire.NewSet(
		userRepository.NewUserRepository,
		wire.Bind(new(userService.Repository), new(*userRepository.UserRepository)),
	)
	EchoSet = wire.NewSet(
		echoRepository.NewEchoRepository,
		wire.Bind(new(echoService.Repository), new(*echoRepository.EchoRepository)),
		wire.Bind(new(connectService.EchoRepository), new(*echoRepository.EchoRepository)),
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
		wire.Bind(new(fileService.KeyValueRepository), new(*keyvalueRepository.KeyValueRepository)),
		wire.Bind(new(settingService.KeyValueRepository), new(*keyvalueRepository.KeyValueRepository)),
		wire.Bind(new(agentService.KeyValueRepository), new(*keyvalueRepository.KeyValueRepository)),
		wire.Bind(new(commentService.KeyValueRepository), new(*keyvalueRepository.KeyValueRepository)),
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
	QueueSet = wire.NewSet(
		queueRepository.NewQueueRepository,
		wire.Bind(new(webhookmodule.DeadLetterStore), new(*queueRepository.QueueRepository)),
		wire.Bind(new(eventsubscriber.DeadLetterRepo), new(*queueRepository.QueueRepository)),
	)
	MigrationSet = wire.NewSet(
		migrationRepository.NewMigrationRepository,
	)
)
