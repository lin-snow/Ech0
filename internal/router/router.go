package router

import (
	"github.com/gin-gonic/gin"

	"github.com/lin-snow/ech0/internal/di"
	"github.com/lin-snow/ech0/internal/middleware"
)

type AppRouterGroup struct {
	ResourceGroup     *gin.RouterGroup
	PublicRouterGroup *gin.RouterGroup
	AuthRouterGroup   *gin.RouterGroup
	WSRouterGroup     *gin.RouterGroup
}

// SetupRouter 配置路由
func SetupRouter(r *gin.Engine, h *di.Handlers) {
	// === 使用本地目录提供前端 ===)
	// // Setup Frontend
	// r.Use(static.Serve("/", static.LocalFile("./template", false)))
	// // 由于Vue3 和SPA模式，所以处理匹配不到的路由(重定向到index.html)
	// r.NoRoute(func(c *gin.Context) {
	// 	c.File("./template/index.html")
	// })

	// === 使用 embed 提供前端 ===)
	setupTemplateRoutes(r, h)

	// ===        中间件        ===
	// Setup Middleware
	setupMiddleware(r)

	// ===     静态资源映射     ===
	// 注意：静态资源映射需要在中间件之后，以便CORS中间件生效
	r.Static("api/images", "./data/images")
	r.Static("api/models", "./data/models")

	// ===  路由组与各模块路由  ===
	// Setup Router Groups
	appRouterGroup := setupRouterGroup(r)

	// Setup Resource Routes
	setupResourceRoutes(appRouterGroup, h)

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

	// Setup Fediverse Routes
	setupFediverseRoutes(appRouterGroup, h)

	// Setup Dashboard Routes
	setupDashboardRoutes(appRouterGroup, h)

	// Setup Agent Routes
	setupAgentRoutes(appRouterGroup, h)
}

// setupRouterGroup 初始化路由组
func setupRouterGroup(r *gin.Engine) *AppRouterGroup {
	resource := r.Group("/")
	public := r.Group("/api")
	auth := r.Group("/api")
	auth.Use(middleware.NoCache(), middleware.JWTAuthMiddleware())
	ws := r.Group("/ws")
	return &AppRouterGroup{
		ResourceGroup:     resource,
		PublicRouterGroup: public,
		AuthRouterGroup:   auth,
		WSRouterGroup:     ws,
	}
}
