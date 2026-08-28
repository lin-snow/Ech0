// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"context"
	"encoding/json"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
)

func (a *Adapter) registerCommonResources(reg *Registry) {
	reg.RegisterResource(ResourceDefinition{
		URI:         "ech0://stats/heatmap",
		Name:        "heatmap",
		Title:       "Post Heatmap",
		Description: "Daily post counts for the past 30 calendar days (UTC day boundaries). Returns an array of {date, count} objects, suitable for calendar heatmap rendering.",
		MimeType:    "application/json",
		Cache:       privateCache(liveTTLMs),
	}, a.resourceHeatmap, authModel.ScopeEchoRead)
}

func (a *Adapter) resourceHeatmap(_ context.Context, _ string) (*ResourceReadResult, error) {
	heatmap, err := a.commonSvc.GetHeatMap("")
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(heatmap)
	return &ResourceReadResult{
		Contents: []ResourceContent{{URI: "ech0://stats/heatmap", MimeType: "application/json", Text: string(data)}},
	}, nil
}
