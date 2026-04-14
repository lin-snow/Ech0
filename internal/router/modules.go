package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

// RouterContext 聚合路由注册需要的上下文。
type RouterContext struct {
	Engine   *gin.Engine
	Handlers *handler.Bundle
	Groups   *AppRouterGroup
}

// RouteModule 定义统一路由模块接口。
type RouteModule interface {
	Name() string
	Register(ctx *RouterContext)
}

type routeModule struct {
	name     string
	register func(ctx *RouterContext)
}

func (m routeModule) Name() string {
	return m.name
}

func (m routeModule) Register(ctx *RouterContext) {
	m.register(ctx)
}

func coreRouteModules() []RouteModule {
	return []RouteModule{
		routeModule{
			name: "template",
			register: func(ctx *RouterContext) {
				setupTemplateRoutes(ctx.Engine, ctx.Handlers)
			},
		},
		routeModule{
			name: "static-files",
			register: func(ctx *RouterContext) {
				cfg := config.Config().Storage
				root := cfg.DataRoot
				if root == "" {
					root = "data/files"
				}
				filesGroup := ctx.Engine.Group("api/files", middleware.StaticFileSecurity())
				filesGroup.StaticFS("/", http.Dir(root))
			},
		},
		routeModule{
			name: "middleware",
			register: func(ctx *RouterContext) {
				setupMiddleware(ctx.Engine)
			},
		},
		routeModule{
			name: "router-groups",
			register: func(ctx *RouterContext) {
				ctx.Groups = setupRouterGroup(ctx.Engine)
			},
		},
	}
}

func featureRouteModules() []RouteModule {
	return []RouteModule{
		routeModule{
			name: "resource",
			register: func(ctx *RouterContext) {
				setupResourceRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "user",
			register: func(ctx *RouterContext) {
				setupUserRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "auth",
			register: func(ctx *RouterContext) {
				setupAuthRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "init",
			register: func(ctx *RouterContext) {
				setupInitRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "echo",
			register: func(ctx *RouterContext) {
				setupEchoRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "comment",
			register: func(ctx *RouterContext) {
				setupCommentRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "common",
			register: func(ctx *RouterContext) {
				setupCommonRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "file",
			register: func(ctx *RouterContext) {
				setupFileRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "setting",
			register: func(ctx *RouterContext) {
				setupSettingRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "connect",
			register: func(ctx *RouterContext) {
				setupConnectRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "dashboard",
			register: func(ctx *RouterContext) {
				setupDashboardRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "agent",
			register: func(ctx *RouterContext) {
				setupAgentRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "migration",
			register: func(ctx *RouterContext) {
				setupMigrationRoutes(ctx.Groups, ctx.Handlers)
			},
		},
		routeModule{
			name: "mcp",
			register: func(ctx *RouterContext) {
				setupMCPRoutes(ctx.Groups, ctx.Handlers)
			},
		},
	}
}
