// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package mcp

import (
	"encoding/json"
	"fmt"

	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	dashboardService "github.com/lin-snow/ech0/internal/service/dashboard"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

type Adapter struct {
	echoSvc      echoService.Service
	userSvc      userService.Service
	commentSvc   commentService.Service
	fileSvc      fileService.Service
	commonSvc    commonService.Service
	connectSvc   connectService.Service
	agentSvc     copilotService.SummaryService
	settingSvc   settingService.Service
	dashboardSvc dashboardService.Service
}

func NewAdapter(
	echoSvc echoService.Service,
	userSvc userService.Service,
	commentSvc commentService.Service,
	fileSvc fileService.Service,
	commonSvc commonService.Service,
	connectSvc connectService.Service,
	agentSvc copilotService.SummaryService,
	settingSvc settingService.Service,
	dashboardSvc dashboardService.Service,
) *Adapter {
	return &Adapter{
		echoSvc:      echoSvc,
		userSvc:      userSvc,
		commentSvc:   commentSvc,
		fileSvc:      fileSvc,
		commonSvc:    commonSvc,
		connectSvc:   connectSvc,
		agentSvc:     agentSvc,
		settingSvc:   settingSvc,
		dashboardSvc: dashboardSvc,
	}
}

func (a *Adapter) RegisterAll(reg *Registry) {
	a.registerEchoTools(reg)
	a.registerEchoResources(reg)
	a.registerUserResources(reg)
	a.registerCommentTools(reg)
	a.registerCommentResources(reg)
	a.registerFileTools(reg)
	a.registerFileResources(reg)
	a.registerCommonResources(reg)
	a.registerConnectTools(reg)
	a.registerConnectResources(reg)
	a.registerAgentTools(reg)
	a.registerWebhookTools(reg)
	a.registerDashboardResources(reg)
}

func stringArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intArg(args map[string]any, key string, fallback int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case json.Number:
			if i, err := n.Int64(); err == nil {
				return int(i)
			}
		}
	}
	return fallback
}

func boolArg(args map[string]any, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func buildTags(args map[string]any) []echoModel.Tag {
	raw, ok := args["tags"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	var tags []echoModel.Tag
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			tags = append(tags, echoModel.Tag{Name: s})
		}
	}
	return tags
}

func buildEchoFiles(args map[string]any) []echoModel.EchoFile {
	raw, ok := args["echo_files"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	var files []echoModel.EchoFile
	for i, v := range arr {
		obj, ok := v.(map[string]any)
		if !ok {
			continue
		}
		fileID := stringArg(obj, "file_id")
		if fileID == "" {
			continue
		}
		sortOrder := intArg(obj, "sort_order", i)
		files = append(files, echoModel.EchoFile{
			FileID:    fileID,
			SortOrder: sortOrder,
		})
	}
	return files
}

func buildExtension(args map[string]any) *echoModel.EchoExtension {
	raw, ok := args["extension"]
	if !ok {
		return nil
	}
	obj, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	extType := stringArg(obj, "type")
	if extType == "" {
		return nil
	}
	payload, _ := obj["payload"].(map[string]any)
	return &echoModel.EchoExtension{
		Type:    extType,
		Payload: payload,
	}
}

func jsonResult(v any) (*ToolCallResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return &ToolCallResult{
		Content: []ContentItem{{Type: "text", Text: string(data)}},
	}, nil
}

func textResult(msg string) *ToolCallResult {
	return &ToolCallResult{
		Content: []ContentItem{{Type: "text", Text: msg}},
	}
}

func textError(msg string) *ToolCallResult {
	return &ToolCallResult{
		Content: []ContentItem{{Type: "text", Text: msg}},
		IsError: true,
	}
}
