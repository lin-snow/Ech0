// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/handler"
)

func setupTemplateRoutes(r *gin.Engine, h *handler.Bundle) {
	r.NoRoute(h.WebHandler.Templates())
}
