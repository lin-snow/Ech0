package router

import "github.com/lin-snow/ech0/internal/handler"

func setupMigrationRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.FullAuthRouterGroup.POST("/migration/upload", h.MigrationHandler.UploadSourceZip())
	appRouterGroup.FullAuthRouterGroup.POST("/migration/start", h.MigrationHandler.StartMigration())
	appRouterGroup.FullAuthRouterGroup.GET("/migration/status", h.MigrationHandler.GetMigrationStatus())
	appRouterGroup.FullAuthRouterGroup.POST("/migration/cancel", h.MigrationHandler.CancelMigration())
	appRouterGroup.FullAuthRouterGroup.POST("/migration/cleanup", h.MigrationHandler.CleanupMigration())
}
