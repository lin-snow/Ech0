package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/middleware"
)

func setupMiddleware(r *gin.Engine) {
	// Cors middleware
	r.Use(middleware.Cors())
}
