package router

import "github.com/lin-snow/ech0/internal/handler"

// setupCommonRoutes 设置普通路由
func setupCommonRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/status", h.CommonHandler.GetStatus())
	appRouterGroup.PublicRouterGroup.GET("/heatmap", h.CommonHandler.GetHeatMap())
	appRouterGroup.PublicRouterGroup.GET("/getmusic", h.CommonHandler.GetPlayMusic())
	appRouterGroup.PublicRouterGroup.GET("/playmusic", h.CommonHandler.PlayMusic)
	appRouterGroup.PublicRouterGroup.GET("/hello", h.CommonHandler.HelloEch0())
	appRouterGroup.PublicRouterGroup.GET("/backup/export", h.BackupHandler.ExportBackup())
	appRouterGroup.PublicRouterGroup.GET("/website/title", h.CommonHandler.GetWebsiteTitle())

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/files/upload", h.CommonHandler.UploadFile())
	appRouterGroup.AuthRouterGroup.DELETE("/files/delete", h.CommonHandler.DeleteFile())
	appRouterGroup.AuthRouterGroup.PUT("/files/presign", h.CommonHandler.GetFilePresignURL())
	appRouterGroup.AuthRouterGroup.POST("/audios/upload", h.CommonHandler.UploadAudio())
	appRouterGroup.AuthRouterGroup.DELETE("/audios/delete", h.CommonHandler.DeleteAudio())
	appRouterGroup.AuthRouterGroup.GET("/backup", h.BackupHandler.Backup())
	appRouterGroup.AuthRouterGroup.POST("/backup/import", h.BackupHandler.ImportBackup())
}
