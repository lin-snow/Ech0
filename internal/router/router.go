package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/di"
)

func SetupRouter(r *gin.Engine, h *di.Handlers) {
	// Setup Middleware
	setupMiddleware(r)

	// Setup Resource Routes
	setupResourceRoutes(r)

	// Setup Router Groups
	appRouterGroup := setupRouterGroup(r)

	// Setup User Routes
	setupUserRoutes(appRouterGroup, h)

	// Setup Echo Routes
}
