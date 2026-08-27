// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/app"
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/database"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	eventsubscriber "github.com/lin-snow/ech0/internal/event/subscriber"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/job"
	jobRunner "github.com/lin-snow/ech0/internal/job/runner"
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/middleware"
	"github.com/lin-snow/ech0/internal/migrator"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	"github.com/lin-snow/ech0/internal/repository"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/service"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	userService "github.com/lin-snow/ech0/internal/service/user"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/task"
	"github.com/lin-snow/ech0/internal/task/scheduled"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/internal/visitor"
	"github.com/lin-snow/ech0/internal/webhook"
	"github.com/lin-snow/ech0/pkg/busen"
	"gorm.io/gorm"
)

var AppSet = app.ProviderSet

var VisitorSet = wire.NewSet(visitor.NewTracker)

func ProvideJobManager(
	repo job.JobRepository,
	reindex *jobRunner.ReindexRunner,
	migration *jobRunner.MigrationRunner,
	export *jobRunner.ExportRunner,
) *job.Manager {
	m := job.NewManager(repo)
	m.Register(jobModel.TypeReindex, job.Adapt(reindex.Run))
	m.Register(jobModel.TypeMigration, job.Adapt(migration.Run))
	m.Register(jobModel.TypeExport, job.Adapt(export.Run))
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
	ProvideStorageKV,
	storage.ProviderSet,
)

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
	repository.FileSet,
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

	handler.NewBundle,
)

var MiddlewareSet = wire.NewSet(
	repository.AuthSet,
	middleware.ProviderSet,
)

var TaskerSet = wire.NewSet(
	repository.FileSet,
	repository.KeyValueSet,
	repository.WebhookSet,

	repository.AuthSet,
	repository.SettingSet,
	service.SettingSet,

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
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
) (*eventbus.EventRegistrar, error) {
	wire.Build(EventSet)
	return &eventbus.EventRegistrar{}, nil
}

func BuildHandlers(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
	tracker *visitor.Tracker,
	jobManager *job.Manager,
	storageManager *storage.Manager,
) (*handler.Bundle, error) {
	wire.Build(HandlerSet)
	return &handler.Bundle{}, nil
}

func BuildJobManager(
	dbProvider func() *gorm.DB,
	appCache cache.ICache[string, any],
	storageManager *storage.Manager,
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
	appCache cache.ICache[string, any],
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
	appCache cache.ICache[string, any],
	tx transaction.Transactor,
	ebProvider func() *busen.Bus,
	tracker *visitor.Tracker,
	storageManager *storage.Manager,
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
