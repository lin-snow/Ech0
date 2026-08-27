// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package server

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/handler"
	"github.com/lin-snow/ech0/internal/middleware"
	"github.com/lin-snow/ech0/internal/router"
)

func ProvideGinEngine() *gin.Engine {
	if config.Config().Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		gin.DebugPrintRouteFunc = func(_, _, _ string, _ int) {}
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	return gin.New()
}

func ProvideHTTPServer(engine *gin.Engine, handlers *handler.Bundle, mwDeps *middleware.Deps) *Server {
	router.SetupRouter(engine, handlers, mwDeps)
	return New(engine)
}

var ProviderSet = wire.NewSet(ProvideGinEngine, ProvideHTTPServer)
