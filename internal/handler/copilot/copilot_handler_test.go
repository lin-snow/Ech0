// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	copilotService "github.com/lin-snow/ech0/internal/service/copilot"
	copilotmock "github.com/lin-snow/ech0/internal/test/mocks/copilotmock"
	timezoneUtil "github.com/lin-snow/ech0/internal/util/timezone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGetRecent(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		summary := copilotmock.NewMockSummaryService(t)
		chat := copilotmock.NewMockChatService(t)
		summary.EXPECT().GetRecent(mock.Anything).Return("近期总结文本", nil).Once()

		h := NewCopilotHandler(summary, chat)
		out, err := h.GetRecent(context.Background(), &GetRecentInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.AGENT_GET_RECENT_SUCCESS, out.Message)
		assert.Equal(t, "近期总结文本", out.Data)
	})

	t.Run("service-error-passthrough", func(t *testing.T) {
		summary := copilotmock.NewMockSummaryService(t)
		chat := copilotmock.NewMockChatService(t)
		sentinel := errors.New("agent down")
		summary.EXPECT().GetRecent(mock.Anything).Return("", sentinel).Once()

		h := NewCopilotHandler(summary, chat)
		out, err := h.GetRecent(context.Background(), &GetRecentInput{})

		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, RecentOutput{}, out)
	})
}

func TestGetSession(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		summary := copilotmock.NewMockSummaryService(t)
		chat := copilotmock.NewMockChatService(t)
		want := []copilotService.ChatMessage{
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hello"},
		}
		chat.EXPECT().GetSession(mock.Anything).Return(want, nil).Once()

		h := NewCopilotHandler(summary, chat)
		out, err := h.GetSession(context.Background(), &GetSessionInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.CHAT_SESSION_GET_SUCCESS, out.Message)
		assert.Equal(t, want, out.Data)
	})

	t.Run("service-error-passthrough", func(t *testing.T) {
		summary := copilotmock.NewMockSummaryService(t)
		chat := copilotmock.NewMockChatService(t)
		sentinel := errors.New("load session failed")
		chat.EXPECT().GetSession(mock.Anything).Return(nil, sentinel).Once()

		h := NewCopilotHandler(summary, chat)
		out, err := h.GetSession(context.Background(), &GetSessionInput{})

		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, SessionOutput{}, out)
	})
}

func TestClearSession(t *testing.T) {
	t.Run("success-returns-nil-data", func(t *testing.T) {
		summary := copilotmock.NewMockSummaryService(t)
		chat := copilotmock.NewMockChatService(t)
		chat.EXPECT().ClearSession(mock.Anything).Return(nil).Once()

		h := NewCopilotHandler(summary, chat)
		out, err := h.ClearSession(context.Background(), &ClearSessionInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, commonModel.CHAT_SESSION_CLEAR_SUCCESS, out.Message)
		assert.Nil(t, out.Data)
	})

	t.Run("service-error-passthrough", func(t *testing.T) {
		summary := copilotmock.NewMockSummaryService(t)
		chat := copilotmock.NewMockChatService(t)
		sentinel := errors.New("clear failed")
		chat.EXPECT().ClearSession(mock.Anything).Return(sentinel).Once()

		h := NewCopilotHandler(summary, chat)
		out, err := h.ClearSession(context.Background(), &ClearSessionInput{})

		require.ErrorIs(t, err, sentinel)
		assert.Equal(t, EmptyOutput{}, out)
	})
}

func TestAsk_TimezoneNormalizationAndStream(t *testing.T) {
	cases := []struct {
		name       string
		headerTZ   string
		wantNormTZ string
	}{
		{name: "valid-iana", headerTZ: "Asia/Shanghai", wantNormTZ: "Asia/Shanghai"},
		{name: "empty-falls-back-utc", headerTZ: "", wantNormTZ: "UTC"},
		{name: "garbage-falls-back-utc", headerTZ: "Not/AZone", wantNormTZ: "UTC"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			summary := copilotmock.NewMockSummaryService(t)
			chat := copilotmock.NewMockChatService(t)
			chat.EXPECT().
				AskStream(mock.Anything, "今天怎么样", "zh-CN", tc.wantNormTZ, mock.Anything).
				Return(nil).Once()

			h := NewCopilotHandler(summary, chat)
			r := gin.New()
			r.POST("/chat/ask", h.Ask())

			req := httptest.NewRequest(http.MethodPost, "/chat/ask",
				strings.NewReader(`{"question":"今天怎么样"}`))
			req.Header.Set("Content-Type", "application/json")
			if tc.headerTZ != "" {
				req.Header.Set(timezoneUtil.DefaultTimezoneHeader, tc.headerTZ)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestAsk_InvalidBodyStillStreamsEmptyQuestion(t *testing.T) {
	summary := copilotmock.NewMockSummaryService(t)
	chat := copilotmock.NewMockChatService(t)
	chat.EXPECT().
		AskStream(mock.Anything, "", "zh-CN", "UTC", mock.Anything).
		Return(nil).Once()

	h := NewCopilotHandler(summary, chat)
	r := gin.New()
	r.POST("/chat/ask", h.Ask())

	req := httptest.NewRequest(http.MethodPost, "/chat/ask", strings.NewReader("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
