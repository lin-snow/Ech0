// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	dashboardService "github.com/lin-snow/ech0/internal/service/dashboard"
	dashboardmock "github.com/lin-snow/ech0/internal/test/mocks/dashboardmock"
	"github.com/lin-snow/ech0/internal/visitor"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetSystemLogs_TailParsing(t *testing.T) {
	cases := []struct {
		name     string
		input    dashboardHandler.GetSystemLogsInput
		wantTail int
	}{
		{
			name:     "empty-tail-defaults-200",
			input:    dashboardHandler.GetSystemLogsInput{},
			wantTail: 200,
		},
		{
			name:     "explicit-tail",
			input:    dashboardHandler.GetSystemLogsInput{Tail: "50", Level: "error", Keyword: "boom"},
			wantTail: 50,
		},
		{
			name:     "whitespace-tail-defaults-200",
			input:    dashboardHandler.GetSystemLogsInput{Tail: "   "},
			wantTail: 200,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := dashboardmock.NewMockService(t)
			want := []logUtil.LogEntry{{Time: "t", Level: "info", Msg: "hi"}}
			svc.EXPECT().
				GetSystemLogs(dashboardService.SystemLogQuery{
					Tail:    tc.wantTail,
					Level:   tc.input.Level,
					Keyword: tc.input.Keyword,
				}).
				Return(want, nil).Once()

			h := dashboardHandler.NewDashboardHandler(svc)
			out, err := h.GetSystemLogs(context.Background(), &tc.input)

			require.NoError(t, err)
			assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
			assert.Equal(t, commonModel.MsgKeyDashboardLogsOk, out.MessageKey)
			assert.Equal(t, want, out.Data)
		})
	}
}

func TestGetSystemLogs_InvalidTail(t *testing.T) {
	cases := []struct {
		name string
		tail string
	}{
		{name: "non-numeric", tail: "abc"},
		{name: "zero", tail: "0"},
		{name: "negative", tail: "-5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := dashboardmock.NewMockService(t)
			h := dashboardHandler.NewDashboardHandler(svc)

			out, err := h.GetSystemLogs(context.Background(), &dashboardHandler.GetSystemLogsInput{Tail: tc.tail})

			require.Error(t, err)
			var biz *commonModel.BizError
			require.ErrorAs(t, err, &biz)
			assert.Equal(t, commonModel.ErrCodeInvalidQuery, biz.Code)
			assert.Equal(t, commonModel.MsgKeyDashboardTailBad, biz.MessageKey)
			assert.Equal(t, dashboardHandler.LogsOutput{}, out)
		})
	}
}

func TestGetSystemLogs_ServiceError(t *testing.T) {
	svc := dashboardmock.NewMockService(t)
	sentinel := errors.New("read log file failed")
	svc.EXPECT().GetSystemLogs(dashboardService.SystemLogQuery{Tail: 200}).Return(nil, sentinel).Once()

	h := dashboardHandler.NewDashboardHandler(svc)
	out, err := h.GetSystemLogs(context.Background(), &dashboardHandler.GetSystemLogsInput{})

	require.ErrorIs(t, err, sentinel)
	assert.Equal(t, dashboardHandler.LogsOutput{}, out)
}

func TestGetVisitorStats(t *testing.T) {
	svc := dashboardmock.NewMockService(t)
	want := []visitor.DayStat{
		{Date: "2026-06-29", PV: 10, UV: 4},
		{Date: "2026-06-30", PV: 20, UV: 7},
	}
	svc.EXPECT().GetVisitorStats().Return(want).Once()

	h := dashboardHandler.NewDashboardHandler(svc)
	out, err := h.GetVisitorStats(context.Background(), &dashboardHandler.GetVisitorStatsInput{})

	require.NoError(t, err)
	assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	assert.Equal(t, want, out.Data)
}

func TestStreamSubscribe_AuthGuard(t *testing.T) {
	routes := map[string]func(*dashboardHandler.DashboardHandler) gin.HandlerFunc{
		"ws":  (*dashboardHandler.DashboardHandler).WSSubscribeSystemLogs,
		"sse": (*dashboardHandler.DashboardHandler).SSESubscribeSystemLogs,
	}
	tokenCases := []struct {
		name  string
		query string
	}{
		{name: "missing-token", query: "/stream"},
		{name: "invalid-token", query: "/stream?token=not-a-jwt"},
	}
	for routeName, build := range routes {
		for _, tc := range tokenCases {
			t.Run(routeName+"/"+tc.name, func(t *testing.T) {
				svc := dashboardmock.NewMockService(t)
				h := dashboardHandler.NewDashboardHandler(svc)
				r := gin.New()
				r.GET("/stream", build(h))

				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.query, nil))

				assert.Equal(t, http.StatusUnauthorized, rec.Code)
			})
		}
	}
}
