//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	busen "github.com/lin-snow/Busen"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/database"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	eventpublisher "github.com/lin-snow/ech0/internal/event/publisher"
	eventregistry "github.com/lin-snow/ech0/internal/event/registry"
	eventsubscriber "github.com/lin-snow/ech0/internal/event/subscriber"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/repository"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/service"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/internal/visitor"
	"github.com/lin-snow/ech0/internal/webhook"
	"gorm.io/gorm"
)

var AppSet = app.ProviderSet

var DomainSet = wire.NewSet(
	BuildHandlers,
	BuildTasker,
	BuildMigrator,
	ProvideBackupScheduleApplier,
	BuildEventRegistrar,
)

var InfraSet = wire.NewSet(
	database.ProviderSet,
	eventbus.ProvideProvider,
	cache.ProviderSet,
	transaction.ProviderSet,
)

var RuntimeSet = server.ProviderSet

var EventGraphSet = wire.NewSet(
	repository.EchoSet,

	repository.UserSet,

	repository.KeyValueSet,
	repository.QueueSet,
	repository.WebhookSet,

	wire.Bind(new(eventregistry.WebhookObserver), new(*webhook.Dispatcher)),
	wire.Bind(new(eventsubscriber.DeadLetterProcessor), new(*webhook.Dispatcher)),

	webhook.NewDispatcher,
	eventsubscriber.NewBackupScheduler,
	eventsubscriber.NewDeadLetterResolver,
	eventsubscriber.NewAgentProcessor,
	ProvideSubscriptionProviders,
	eventregistry.NewEventRegistry,
)

var HandlerGraphSet = wire.NewSet(
	eventpublisher.New,
	wire.Bind(new(commentService.EventPublisher), new(*eventpublisher.Publisher)),
	storage.ProviderSet,
	wire.Bind(new(storage.S3SettingStore), new(*keyvalueRepository.KeyValueRepository)),
	repository.FileSet,
	visitor.NewTracker,
	handler.WebSet,

	repository.UserSet,
	repository.AuthSet,
	service.UserSet,
	service.AuthSet,
	handler.UserSet,
	handler.AuthSet,

	repository.EchoSet,
	service.EchoSet,
	handler.EchoSet,
	repository.CommentSet,
	service.CommentSet,
	handler.CommentSet,

	repository.CommonSet,
	service.FileSet,
	handler.FileSet,
	repository.InitSet,
	service.InitSet,
	handler.InitSet,
	service.CommonSet,
	handler.CommonSet,

	repository.WebhookSet,
	repository.KeyValueSet,

	repository.SettingSet,
	service.SettingSet,
	handler.SettingSet,

	repository.ConnectSet,
	service.ConnectSet,
	handler.ConnectSet,

	service.DashboardSet,
	handler.DashboardSet,

	service.AgentSet,
	handler.AgentSet,

	service.BackupSet,
	handler.BackupSet,
	repository.MigrationSet,
	service.MigratorSet,
	handler.MigrationSet,

	handler.MCPSet,

	handler.NewBundle,
)

var TaskerGraphSet = wire.NewSet(
	eventpublisher.New,
	storage.ProviderSet,
	wire.Bind(new(storage.S3SettingStore), new(*keyvalueRepository.KeyValueRepository)),
	repository.FileSet,
	repository.KeyValueSet,
	repository.WebhookSet,

	repository.SettingSet,
	service.SettingSet,

	repository.EchoSet,
	service.EchoSet,

	repository.CommonSet,
	service.FileSet,
	service.CommonSet,

	repository.QueueSet,
	task.ProviderSet,
)

var MigratorGraphSet = wire.NewSet(
	migrator.ProviderSet,
)

// BuildApp 构建 Web 生命周期应用。
func BuildApp() (*app.App, error) {
	wire.Build(
		InfraSet,
		DomainSet,
		RuntimeSet,
		AppSet,
	)
	return &app.App{}, nil
}

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() *busen.Bus,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	backupScheduleApplier eventsubscriber.BackupScheduleApplier,
) (*eventregistry.EventRegistrar, error) {
	wire.Build(EventGraphSet)
	return &eventregistry.EventRegistrar{}, nil
}

// BuildHandlers 使用 wire 生成的代码来构建 Handlers 实例。
func BuildHandlers(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
) (*handler.Bundle, error) {
	wire.Build(HandlerGraphSet)
	return &handler.Bundle{}, nil
}

// BuildServer 构建 HTTP server
func BuildServer() (*server.Server, error) {
	wire.Build(
		InfraSet,
		BuildHandlers,
		server.ProviderSet,
	)
	return &server.Server{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
) (*task.Tasker, error) {
	wire.Build(TaskerGraphSet)
	return &task.Tasker{}, nil
}

func BuildMigrator(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
) (*migrator.Worker, error) {
	wire.Build(MigratorGraphSet)
	return &migrator.Worker{}, nil
}

func ProvideBackupScheduleApplier(t *task.Tasker) eventsubscriber.BackupScheduleApplier {
	return t
}

func ProvideSubscriptionProviders(
	dlr *eventsubscriber.DeadLetterResolver,
	bs *eventsubscriber.BackupScheduler,
	ap *eventsubscriber.AgentProcessor,
) []eventregistry.SubscriptionProvider {
	return []eventregistry.SubscriptionProvider{dlr, bs, ap}
}
