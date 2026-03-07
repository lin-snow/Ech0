package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/router"
)

func ProvideGinEngine() *gin.Engine {
	if config.Config().Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}

func ProvideHTTPServer(engine *gin.Engine, handlers *handler.Bundle) *Server {
	router.SetupRouter(engine, handlers)
	return New(engine)
}

var ProviderSet = wire.NewSet(ProvideGinEngine, ProvideHTTPServer)
