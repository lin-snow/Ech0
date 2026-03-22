package router

import "github.com/lin-snow/ech0/internal/handler"

func setupFileRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Auth
	appRouterGroup.FullAuthRouterGroup.POST("/files/upload", h.FileHandler.UploadFile())
	appRouterGroup.FullAuthRouterGroup.GET("/files", h.FileHandler.ListFiles())
	appRouterGroup.FullAuthRouterGroup.GET("/file/tree", h.FileHandler.ListFileTree())
	appRouterGroup.FullAuthRouterGroup.GET("/file/stream", h.FileHandler.StreamFileByPath)
	appRouterGroup.FullAuthRouterGroup.GET("/file/:id", h.FileHandler.GetFileByID())
	appRouterGroup.FullAuthRouterGroup.GET("/file/:id/stream", h.FileHandler.StreamFileByID)
	appRouterGroup.FullAuthRouterGroup.PUT("/file/:id/meta", h.FileHandler.UpdateFileMeta())
	appRouterGroup.FullAuthRouterGroup.POST("/files/external", h.FileHandler.CreateExternalFile())
	appRouterGroup.FullAuthRouterGroup.DELETE("/file/:id", h.FileHandler.DeleteFile())
	appRouterGroup.FullAuthRouterGroup.PUT("/files/presign", h.FileHandler.GetFilePresignURL())
}
