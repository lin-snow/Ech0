package router

import "github.com/lin-snow/ech0/internal/di"

// setupCommonRoutes 设置普通路由
func setupCommonRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/status", h.CommonHandler.GetStatus())
	appRouterGroup.PublicRouterGroup.GET("/heatmap", h.CommonHandler.GetHeatMap())
	appRouterGroup.PublicRouterGroup.GET("/getmusic", h.CommonHandler.GetPlayMusic())
	appRouterGroup.PublicRouterGroup.GET("/playmusic", h.CommonHandler.PlayMusic)
	appRouterGroup.PublicRouterGroup.GET("/hello", h.CommonHandler.HelloEch0())
	appRouterGroup.PublicRouterGroup.GET("/backup/export", h.BackupHandler.ExportBackup())
	appRouterGroup.PublicRouterGroup.GET("/website/title", h.CommonHandler.GetWebsiteTitle())

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/images/upload", h.CommonHandler.UploadImage())
	appRouterGroup.AuthRouterGroup.DELETE("/images/delete", h.CommonHandler.DeleteImage())
	appRouterGroup.AuthRouterGroup.POST("/audios/upload", h.CommonHandler.UploadAudio())
	appRouterGroup.AuthRouterGroup.DELETE("/audios/delete", h.CommonHandler.DeleteAudio())
	appRouterGroup.AuthRouterGroup.POST("/models/upload", h.CommonHandler.UploadModel())
	appRouterGroup.AuthRouterGroup.DELETE("/models/delete", h.CommonHandler.DeleteModel())
	appRouterGroup.AuthRouterGroup.GET("/backup", h.BackupHandler.Backup())
	appRouterGroup.AuthRouterGroup.POST("/backup/import", h.BackupHandler.ImportBackup())
	appRouterGroup.AuthRouterGroup.PUT("/s3/presign", h.CommonHandler.GetS3PresignURL())
}
