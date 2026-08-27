// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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
	CheckUpdateOutput  = commonModel.Result[CheckUpdateResponse]
	LogsOutput         = commonModel.Result[[]logUtil.LogEntry]
	VisitorStatsOutput = commonModel.Result[[]visitor.DayStat]
)

func (dashboardHandler *DashboardHandler) CheckUpdate(ctx context.Context, _ *CheckUpdateInput) (CheckUpdateOutput, error) {
	latestVersion, err := githubUtil.GetLatestVersion()
	if err != nil {
		return CheckUpdateOutput{}, &commonModel.BizError{
			Code:       commonModel.ErrCodeInternal,
			Msg:        commonModel.CHECK_UPDATE_FAILED,
			MessageKey: commonModel.MsgKeyDashboardCheckUpdateFailed,
			Err:        err,
		}
	}

	cur := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(versionPkg.Version), "v"))
	lat := semver.Canonical("v" + strings.TrimPrefix(strings.TrimSpace(latestVersion), "v"))
	hasUpdate := cur != "" && lat != "" && semver.Compare(lat, cur) > 0

	return commonModel.OK(CheckUpdateResponse{
		CurrentVersion: versionPkg.Version,
		LatestVersion:  latestVersion,
		HasUpdate:      hasUpdate,
	}), nil
}

func (dashboardHandler *DashboardHandler) GetSystemLogs(ctx context.Context, in *GetSystemLogsInput) (LogsOutput, error) {
	tail := 200
	if rawTail := strings.TrimSpace(in.Tail); rawTail != "" {
		n, err := strconv.Atoi(rawTail)
		if err != nil || n <= 0 {
			return LogsOutput{}, commonModel.NewBizErrorWithMessageKey(
				commonModel.ErrCodeInvalidQuery, commonModel.INVALID_QUERY_PARAMS, commonModel.MsgKeyDashboardTailBad, nil,
			)
		}
		tail = n
	}

	logs, err := dashboardHandler.dashboardService.GetSystemLogs(service.SystemLogQuery{
		Tail:    tail,
		Level:   in.Level,
		Keyword: in.Keyword,
	})
	if err != nil {
		return LogsOutput{}, err
	}
	result := commonModel.OK(logs, "获取系统日志成功")
	result.MessageKey = commonModel.MsgKeyDashboardLogsOk
	return result, nil
}

func (dashboardHandler *DashboardHandler) GetVisitorStats(ctx context.Context, _ *GetVisitorStatsInput) (VisitorStatsOutput, error) {
	return commonModel.OK(dashboardHandler.dashboardService.GetVisitorStats()), nil
}

func (dashboardHandler *DashboardHandler) WSSubscribeSystemLogs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "missing token"})
			return
		}

		token = strings.Trim(token, `"`)
		if _, err := jwtUtil.ParseToken(token); err != nil {
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
		if _, err := jwtUtil.ParseToken(token); err != nil {
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
