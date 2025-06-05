package router

import (
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/middleware"
)

type AppRouterGroup struct {
	PublicRouterGroup *gin.RouterGroup
	AuthRouterGroup   *gin.RouterGroup
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
