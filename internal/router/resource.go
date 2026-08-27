// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/lin-snow/ech0/internal/handler"
)

func setupResourceRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	appRouterGroup.ResourceGroup.GET("/robots.txt", h.CommonHandler.GetRobotsTxt)
	appRouterGroup.ResourceGroup.GET("/sitemap.xml", h.CommonHandler.GetSitemap)
	appRouterGroup.ResourceGroup.GET("/rss", h.CommonHandler.GetRss)
	appRouterGroup.ResourceGroup.GET("/healthz", h.CommonHandler.Healthz())
}
