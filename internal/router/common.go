package router

import "github.com/lin-snow/ech0/internal/di"

func setupCommonRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public
	appRouterGroup.PublicRouterGroup.GET("/status", h.CommonHandler.GetStatus)

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/images/upload", h.CommonHandler.UploadImage)
	appRouterGroup.AuthRouterGroup.POST("/images/delete", h.CommonHandler.DeleteImage)
}
