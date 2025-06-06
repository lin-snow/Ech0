package router

import (
	"github.com/gin-gonic/gin"
)

func setupNoRoutes(r *gin.Engine) {
	r.NoRoute(func(c *gin.Context) {
		c.File("./template/index.html")
	})
}
