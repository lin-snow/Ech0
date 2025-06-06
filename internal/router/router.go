package router

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/di"
	"github.com/lin-snow/ech0/internal/middleware"
)

type AppRouterGroup struct {
	PublicRouterGroup *gin.RouterGroup
	AuthRouterGroup   *gin.RouterGroup
}

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
	setupEchoRoutes(appRouterGroup, h)

	// Setup Common Routes
	setupCommonRoutes(appRouterGroup, h)

	// Setup Setting Routes
	setupSettingRoutes(appRouterGroup, h)

	// Setup To Do Routes
	setupTodoRoutes(appRouterGroup, h)

	// Setup Connect Routes
	setupConnectRoutes(appRouterGroup, h)

	// Setup No Routes
	setupNoRoutes(r)
}

func setupRouterGroup(r *gin.Engine) *AppRouterGroup {
	public := r.Group("/api")
	auth := r.Group("/api")
	auth.Use(middleware.JWTAuthMiddleware())
	return &AppRouterGroup{
		PublicRouterGroup: public,
		AuthRouterGroup:   auth,
	}
}

func setupResourceRoutes(r *gin.Engine) {
	r.Use(static.Serve("/", static.LocalFile("./template", true)))
	r.Static("/api/images", "./data/images")
}
