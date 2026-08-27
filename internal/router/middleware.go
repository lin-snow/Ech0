// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/config"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	"github.com/lin-snow/ech0/internal/middleware"
)

func setupMiddleware(r *gin.Engine) {
	if config.Config().Server.Mode == "debug" {
		r.Use(gin.Logger())
	}
	r.Use(gin.Recovery())
	r.Use(middleware.PoweredBy())
	r.Use(middleware.Cors())
	r.Use(i18nUtil.Middleware())
	r.Use(middleware.WriteGuard())
}
