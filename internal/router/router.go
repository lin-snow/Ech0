package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
)

type AppRouterGroup struct {
	ResourceGroup     *gin.RouterGroup
	PublicRouterGroup *gin.RouterGroup
	AuthRouterGroup   *gin.RouterGroup
	WSRouterGroup     *gin.RouterGroup
	MCPRouterGroup    *gin.RouterGroup
}

// SetupRouter 配置路由
func SetupRouter(r *gin.Engine, h *handler.Bundle, mwDeps *middleware.Deps) {
	ctx := &RouterContext{
		Engine:   r,
		Handlers: h,
		MWDeps:   mwDeps,
	}

	for _, module := range coreRouteModules() {
		module.Register(ctx)
	}
	for _, module := range featureRouteModules() {
		module.Register(ctx)
	}
}

// setupRouterGroup 初始化路由组
func setupRouterGroup(r *gin.Engine, mwDeps *middleware.Deps) *AppRouterGroup {
	resource := r.Group("/")
	public := r.Group("/api")
	auth := r.Group("/api")
	auth.Use(middleware.NoCache(), middleware.JWTAuthMiddleware(mwDeps.TokenRevoker))
	ws := r.Group("/ws")
	mcpGroup := r.Group("/mcp")
	mcpGroup.Use(middleware.NoCache(), middleware.JWTAuthMiddleware(mwDeps.TokenRevoker))
	return &AppRouterGroup{
		ResourceGroup:     resource,
		PublicRouterGroup: public,
		AuthRouterGroup:   auth,
		WSRouterGroup:     ws,
		MCPRouterGroup:    mcpGroup,
	}
}
