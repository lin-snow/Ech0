// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

package di

import (
	"context"

	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	eventsubscriber "github.com/lin-snow/ech0/internal/event/subscriber"
	"github.com/lin-snow/ech0/internal/handler"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
	"github.com/lin-snow/ech0/internal/i18n"
	"github.com/lin-snow/ech0/internal/job"
	jobRunner "github.com/lin-snow/ech0/internal/job/runner"
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/middleware"
	"github.com/lin-snow/ech0/internal/migrator"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	"github.com/lin-snow/ech0/internal/ratelimit"
	"github.com/lin-snow/ech0/internal/repository"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	embeddingRepository "github.com/lin-snow/ech0/internal/repository/embedding"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/service"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/task/scheduled"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/internal/visitor"
	"github.com/lin-snow/ech0/internal/webhook"
	"github.com/lin-snow/ech0/pkg/app"
	"github.com/lin-snow/ech0/pkg/busen"
	"gorm.io/gorm"
)

var AppSet = wire.NewSet(ProvideApp)

// ProvideApp composes Ech0's runtime on the generic app lifecycle: components
// start in order job → task → server; the JWT signing secret is pinned and
// settings seeded before any of them, and event subscriptions registered and
// torn down around them.
func ProvideApp(
	registrar *eventbus.EventRegistrar,
	jobManager *job.Manager,
	taskManager *task.Manager,
	httpServer *server.Server,
	durableKV kvstore.Store,
	dbProvider func() *gorm.DB,
	storages *storage.Registry,
) *app.App {
	_ = storages
	return app.New(
		app.Components(jobManager, taskManager, httpServer),
		app.BeforeStart(func(ctx context.Context) error {
			if err := config.EnsureJWTSecret(ctx, durableKV); err != nil {
				return err
			}
			// The default locale is a setting; i18n asks through this hook so it
			// never reads the settings table itself.
			i18n.SetDefaultLocaleProvider(func() string {
				system, err := setting.System.Get(context.Background(), durableKV)
				if err != nil {
					return ""
				}
				return system.DefaultLocale
			})
			if err := setting.Migrate(ctx, durableKV); err != nil {
				return err
			}
			return registrar.Register()
		}),
		app.AfterStop(func(context.Context) error {
			return registrar.Stop()
		}),
	)
}

var VisitorSet = wire.NewSet(visitor.NewTracker)

func ProvideJobManager(
	repo job.JobRepository,
	reindex *jobRunner.ReindexRunner,
	migration *jobRunner.MigrationRunner,
	export *jobRunner.ExportRunner,
) *job.Manager {
	m := job.NewManager(repo)
	m.Register(jobModel.Reindex, reindex.Run)
	m.Register(jobModel.Migration, migration.Run)
	m.Register(jobModel.Export, export.Run)
	return m
}

func ProvideTaskManager(
	cleanup *scheduled.Cleanup,
	snapshot *scheduled.Snapshot,
	visitorSnapshot *scheduled.VisitorSnapshot,
) (*task.Manager, error) {
	return task.NewManager(cleanup, snapshot, visitorSnapshot)
}

var StorageSet = wire.NewSet(
	keyvalueRepository.NewKeyValueRepository,
	repository.FileSet,
	ProvideStorageKV,
	ProvideStorageRegistry,
)

// ProvideStorageRegistry builds the one registry every writer and reader shares.
// It starts from the storages table (seeding the configured default when the
// table is empty), so a boot-time consumer — the snapshot exporter, say — sees
// the same stores as the first upload.
func ProvideStorageRegistry(repo *fileRepository.FileRepository) (*storage.Registry, error) {
	ctx := context.Background()
	cfgs, err := fileService.LoadStorageConfigs(ctx, repo)
	if err != nil {
		return nil, err
	}
	registry := storage.NewRegistry()
	if err := registry.Reload(ctx, cfgs); err != nil {
		return nil, err
	}
	return registry, nil
}

func ProvideStorageKV(repo *keyvalueRepository.KeyValueRepository) kvstore.Store {
	return kvstore.NewPersistent(repo)
}

func ProvideGormDB(dbProvider func() *gorm.DB) *gorm.DB {
	return dbProvider()
}

var DomainSet = wire.NewSet(
	BuildHandlers,
	BuildMiddlewares,
	BuildTasker,
	BuildJobManager,
	BuildEventRegistrar,
)

var InfraSet = wire.NewSet(
	database.ProviderSet,
	eventbus.ProvideProvider,
	cache.ProviderSet,
	transaction.ProviderSet,
)

var RuntimeSet = server.ProviderSet

var EventSet = wire.NewSet(
	repository.EchoSet,

	repository.UserSet,

	repository.KeyValueSet,
	repository.WebhookSet,
	repository.EmbeddingSet,

	webhook.NewDispatcher,
	eventsubscriber.NewAgentProcessor,
	eventsubscriber.NewEmbeddingProcessor,
	service.EmbeddingSet,
	ProvideSubscriptionProviders,
	eventbus.NewEventRegistry,
)

var HandlerSet = wire.NewSet(
	ratelimit.ProviderSet,
	repository.FileSet,
	handler.WebSet,

	repository.UserSet,
	repository.AuthSet,
	service.UserSet,
	wire.Bind(new(userService.EchoCacheInvalidator), new(*echoRepository.EchoRepository)),
	wire.Bind(new(userService.EmbeddingPurger), new(*embeddingRepository.EmbeddingRepository)),
	service.AuthSet,
	service.AccessSet,
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
	webhook.NewSender,
	repository.KeyValueSet,

	repository.SettingSet,
	service.SettingSet,
	handler.SettingSet,

	repository.ConnectSet,
	service.ConnectSet,
	handler.ConnectSet,

	service.DashboardSet,
	handler.DashboardSet,

	repository.EmbeddingSet,
	service.EmbeddingSet,
	handler.EmbeddingSet,

	service.CopilotSet,
	wire.Bind(new(copilotService.UserReader), new(*userService.UserService)),
	handler.CopilotSet,

	service.MigratorSet,
	handler.MigrationSet,

	handler.MCPSet,

	repository.OAuthServerSet,
	service.OAuthServerSet,
	handler.OAuthServerSet,

	// The shell renderer reads site settings and shareable echos directly.
	wire.Bind(new(webHandler.SettingsReader), new(*settingService.SettingService)),
	wire.Bind(new(webHandler.EchoReader), new(*echoService.EchoService)),

	handler.NewBundle,
)

var MiddlewareSet = wire.NewSet(
	ratelimit.ProviderSet,
	repository.AuthSet,
	middleware.ProviderSet,
)

var TaskerSet = wire.NewSet(
	repository.FileSet,
	repository.KeyValueSet,
	repository.WebhookSet,

	repository.AuthSet,
	repository.SettingSet,
	repository.UserSet,
	service.SettingSet,
	service.AccessSet,

	repository.EchoSet,
	service.EchoSet,

	repository.CommonSet,
	service.FileSet,
	service.CommonSet,

	repository.VisitorSet,
	migrator.NewExportEngine,
	scheduled.ProviderSet,
	ProvideTaskManager,
)

func BuildApp() (*app.App, error) {
	wire.Build(
		InfraSet,
		VisitorSet,
		StorageSet,
		DomainSet,
		RuntimeSet,
		AppSet,
	)
	return &app.App{}, nil
}

func BuildEventRegistrar(
	dbProvider func() *gorm.DB,
	ebProvider func() *busen.Bus,
	appCache *cache.Cache,
	tx transaction.Transactor,
) (*eventbus.EventRegistrar, error) {
	wire.Build(EventSet)
	return &eventbus.EventRegistrar{}, nil
}

func BuildHandlers(
	dbProvider func() *gorm.DB,
	appCache *cache.Cache,
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
	tracker *visitor.Tracker,
	jobManager *job.Manager,
	storages *storage.Registry,
) (*handler.Bundle, error) {
	wire.Build(HandlerSet)
	return &handler.Bundle{}, nil
}

func BuildJobManager(
	dbProvider func() *gorm.DB,
	appCache *cache.Cache,
	storages *storage.Registry,
	ebProvider func() *busen.Bus,
	tx transaction.Transactor,
) (*job.Manager, error) {
	wire.Build(
		repository.JobSet,
		repository.EmbeddingSet,
		repository.EchoSet,
		repository.KeyValueSet,
		service.EmbeddingSet,
		migrator.NewImportEngine,
		migrator.NewExportEngine,
		ProvideGormDB,
		migrator.NewCapsuleEngine,
		jobRunner.ProviderSet,
		ProvideJobManager,
	)
	return nil, nil
}

func BuildMiddlewares(
	dbProvider func() *gorm.DB,
	appCache *cache.Cache,
) (*middleware.Deps, error) {
	wire.Build(MiddlewareSet)
	return &middleware.Deps{}, nil
}

func BuildServer() (*server.Server, error) {
	wire.Build(
		InfraSet,
		VisitorSet,
		StorageSet,
		BuildJobManager,
		BuildHandlers,
		BuildMiddlewares,
		server.ProviderSet,
	)
	return &server.Server{}, nil
}

func BuildTasker(
	dbProvider func() *gorm.DB,
	appCache *cache.Cache,
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
	tracker *visitor.Tracker,
	storages *storage.Registry,
) (*task.Manager, error) {
	wire.Build(TaskerSet)
	return &task.Manager{}, nil
}

func ProvideSubscriptionProviders(
	ap *eventsubscriber.AgentProcessor,
	ep *eventsubscriber.EmbeddingProcessor,
	disp *webhook.Dispatcher,
) []eventbus.Subscriber {
	return []eventbus.Subscriber{ap, ep, disp}
}
