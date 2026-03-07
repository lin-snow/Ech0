package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
)

// setupTemplateRoutes 设置模板路由
func setupTemplateRoutes(r *gin.Engine, h *handler.Bundle) {
	r.NoRoute(h.WebHandler.Templates())
}
