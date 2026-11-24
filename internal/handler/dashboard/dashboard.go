// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lin-snow/ech0/internal/apperr"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	service "github.com/lin-snow/ech0/internal/service/dashboard"
	githubUtil "github.com/lin-snow/ech0/internal/util/github"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	versionPkg "github.com/lin-snow/ech0/internal/version"
	"github.com/lin-snow/ech0/internal/visitor"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"golang.org/x/mod/semver"
)

type DashboardHandler struct {
	dashboardService service.Service
}

func NewDashboardHandler(dashboardService service.Service) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

type (
	CheckUpdateInput   struct{}
	GetSystemLogsInput struct {
		Tail    string `query:"tail" doc:"返回最近多少条（默认 200）"`
		Level   string `query:"level" doc:"日志级别过滤"`
		Keyword string `query:"keyword" doc:"关键词过滤"`
	}
	GetVisitorStatsInput struct{}

	CheckUpdateResponse struct {
		CurrentVersion string `json:"current_version"`
		LatestVersion  string `json:"latest_version"`
		HasUpdate      bool   `json:"has_update"`
	}
)

type (
	CheckUpdateOutput struct{ Body CheckUpdateResponse }
	LogsOutput        struct {
		Body commonModel.List[logUtil.LogEntry]
	}
	VisitorStatsOutput struct {
		Body commonModel.List[visitor.DayStat]
	}
)

func (dashboardHandler *DashboardHandler) CheckUpdate(_ context.Context, _ *CheckUpdateInput) (*CheckUpdateOutput, error) {
	latestVersion, err := githubUtil.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("check for updates: %w", err)
	}

	cur := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(versionPkg.Version), "v"))
	lat := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(latestVersion), "v"))
	hasUpdate := cur != "" && lat != "" && semver.Compare(lat, cur) > 0

	return &CheckUpdateOutput{Body: CheckUpdateResponse{
		CurrentVersion: versionPkg.Version,
		LatestVersion:  latestVersion,
		HasUpdate:      hasUpdate,
	}}, nil
}

func (dashboardHandler *DashboardHandler) GetSystemLogs(_ context.Context, in *GetSystemLogsInput) (*LogsOutput, error) {
	tail := 200
	if rawTail := strings.TrimSpace(in.Tail); rawTail != "" {
		n, err := strconv.Atoi(rawTail)
		if err != nil || n <= 0 {
			return nil, apperr.ErrInvalidQuery
		}
		tail = n
	}

	logs, err := dashboardHandler.dashboardService.GetSystemLogs(service.SystemLogQuery{
		Tail:    tail,
		Level:   in.Level,
		Keyword: in.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &LogsOutput{Body: commonModel.List[logUtil.LogEntry]{Items: logs}}, nil
}

func (dashboardHandler *DashboardHandler) GetVisitorStats(_ context.Context, _ *GetVisitorStatsInput) (*VisitorStatsOutput, error) {
	return &VisitorStatsOutput{Body: commonModel.List[visitor.DayStat]{Items: dashboardHandler.dashboardService.GetVisitorStats()}}, nil
}

func (dashboardHandler *DashboardHandler) WSSubscribeSystemLogs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "missing token"})
			return
		}

		token = strings.Trim(token, `"`)
		if _, err := jwtUtil.ParseSessionToken(token); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "invalid token"})
			return
		}

		err := dashboardHandler.dashboardService.WSSubscribeSystemLogs(
			ctx.Writer,
			ctx.Request,
			service.SystemLogStreamFilter{
				Level:   ctx.Query("level"),
				Keyword: ctx.Query("keyword"),
			},
		)
		if err != nil {
			logUtil.GetLogger().Error("WebSocket Subscribe System Logs Failed", logUtil.Err(err))
		}
	}
}

func (dashboardHandler *DashboardHandler) SSESubscribeSystemLogs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "missing token"})
			return
		}

		token = strings.Trim(token, `"`)
		if _, err := jwtUtil.ParseSessionToken(token); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "invalid token"})
			return
		}

		err := dashboardHandler.dashboardService.SSESubscribeSystemLogs(
			ctx.Writer,
			ctx.Request,
			service.SystemLogStreamFilter{
				Level:   ctx.Query("level"),
				Keyword: ctx.Query("keyword"),
			},
		)
		if err != nil {
			logUtil.GetLogger().Error("SSE Subscribe System Logs Failed", logUtil.Err(err))
		}
	}
}
