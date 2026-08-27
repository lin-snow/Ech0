// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func (a *Adapter) registerDashboardResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://stats/visitors",
		Name:        "visitor_stats",
		Title:       "Visitor Stats",
		Description: "Daily visitor statistics (page views and unique visitors) for the past 7 days (UTC day boundaries). Returns an array of {date, pv, uv} objects. Requires admin scope.",
		MimeType:    "application/json",
	}, a.resourceVisitorStats, authModel.ScopeAdminSettings)
}

func (a *Adapter) resourceVisitorStats(_ context.Context, _ string) (*ResourceReadResult, error) {
	stats := a.dashboardSvc.GetVisitorStats()
	data, _ := json.Marshal(stats)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://stats/visitors", MimeType: "application/json", Text: string(data)}},
	}, nil
}
