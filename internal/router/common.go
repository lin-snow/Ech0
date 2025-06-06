package router

import "github.com/lin-snow/ech0/internal/di"

func setupCommonRoutes(appRouterGroup *AppRouterGroup, h *di.Handlers) {
	// Public

	// Auth
	appRouterGroup.AuthRouterGroup.POST("/images/upload", h.CommonHandler.UploadImage)
}
