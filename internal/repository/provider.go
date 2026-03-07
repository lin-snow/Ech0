package repository

import (
	"github.com/google/wire"
	commonRepository "github.com/lin-snow/ech0/internal/repository/common"
	connectRepository "github.com/lin-snow/ech0/internal/repository/connect"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	settingRepository "github.com/lin-snow/ech0/internal/repository/setting"
	todoRepository "github.com/lin-snow/ech0/internal/repository/todo"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	webhookRepository "github.com/lin-snow/ech0/internal/repository/webhook"
)

var (
	UserSet     = wire.NewSet(userRepository.NewUserRepository)
	EchoSet     = wire.NewSet(echoRepository.NewEchoRepository)
	CommonSet   = wire.NewSet(commonRepository.NewCommonRepository)
	FileSet     = wire.NewSet(fileRepository.NewFileRepository)
	KeyValueSet = wire.NewSet(keyvalueRepository.NewKeyValueRepository)
	SettingSet  = wire.NewSet(settingRepository.NewSettingRepository)
	TodoSet     = wire.NewSet(todoRepository.NewTodoRepository)
	ConnectSet  = wire.NewSet(connectRepository.NewConnectRepository)
	WebhookSet  = wire.NewSet(webhookRepository.NewWebhookRepository)
	InboxSet    = wire.NewSet(inboxRepository.NewInboxRepository)
	QueueSet    = wire.NewSet(queueRepository.NewQueueRepository)
)
