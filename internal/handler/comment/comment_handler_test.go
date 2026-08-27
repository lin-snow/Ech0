// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package handler_test

import (
	"context"
	"errors"
	"testing"

	handler "github.com/lin-snow/ech0/internal/handler/comment"
	model "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/test/mocks/commentmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func bg() context.Context { return context.Background() }

func TestCommentHandler_GetFormMeta(t *testing.T) {
	t.Run("success with empty meta when no middleware ran", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().GetFormMeta(mock.Anything, "", "").
			Return(model.FormMeta{FormToken: "tok", EnableComment: true}, nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.GetFormMeta(bg(), &handler.GetFormMetaInput{})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
		assert.Equal(t, "tok", out.Data.FormToken)
		assert.True(t, out.Data.EnableComment)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().GetFormMeta(mock.Anything, "", "").
			Return(model.FormMeta{}, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.GetFormMeta(bg(), &handler.GetFormMetaInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, handler.FormMetaOutput{}, out)
	})
}

func TestCommentHandler_ListCommentsByEchoID(t *testing.T) {
	t.Run("trims echo id before delegating", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().ListPublicByEchoID(mock.Anything, "echo-1").
			Return([]model.PublicComment{{ID: "c1"}}, nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.ListCommentsByEchoID(bg(), &handler.ListCommentsByEchoInput{EchoID: "  echo-1  "})

		require.NoError(t, err)
		assert.Len(t, out.Data, 1)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().ListPublicByEchoID(mock.Anything, mock.Anything).Return(nil, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.ListCommentsByEchoID(bg(), &handler.ListCommentsByEchoInput{EchoID: "echo-1"})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_ListPublicComments(t *testing.T) {
	t.Run("forwards limit", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().ListPublicComments(mock.Anything, 30).
			Return([]model.PublicComment{{ID: "c1"}, {ID: "c2"}}, nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.ListPublicComments(bg(), &handler.ListPublicCommentsInput{Limit: 30})

		require.NoError(t, err)
		assert.Len(t, out.Data, 2)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().ListPublicComments(mock.Anything, mock.Anything).Return(nil, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.ListPublicComments(bg(), &handler.ListPublicCommentsInput{Limit: 10})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_CreateComment(t *testing.T) {
	t.Run("forwards meta and body, returns result", func(t *testing.T) {
		dto := model.CreateCommentDto{EchoID: "e1", Content: "hi", FormToken: "tok"}
		svc := commentmock.NewMockService(t)
		svc.EXPECT().
			CreateComment(mock.Anything, "", "", mock.MatchedBy(func(d *model.CreateCommentDto) bool {
				return d != nil && d.EchoID == "e1" && d.Content == "hi"
			})).
			Return(model.CreateCommentResult{ID: "c1", Status: model.StatusPending}, nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.CreateComment(bg(), &handler.CreateCommentInput{Body: dto})

		require.NoError(t, err)
		assert.Equal(t, "c1", out.Data.ID)
		assert.Equal(t, model.StatusPending, out.Data.Status)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().CreateComment(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(model.CreateCommentResult{}, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.CreateComment(bg(), &handler.CreateCommentInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, handler.CreateCommentOutput{}, out)
	})
}

func TestCommentHandler_CreateIntegrationComment(t *testing.T) {
	t.Run("forwards body and returns result", func(t *testing.T) {
		dto := model.CreateIntegrationCommentDto{EchoID: "e1", Content: "via token"}
		svc := commentmock.NewMockService(t)
		svc.EXPECT().
			CreateIntegrationComment(mock.Anything, "", "", mock.MatchedBy(func(d *model.CreateIntegrationCommentDto) bool {
				return d != nil && d.EchoID == "e1" && d.Content == "via token"
			})).
			Return(model.CreateCommentResult{ID: "c9", Status: model.StatusApproved}, nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.CreateIntegrationComment(bg(), &handler.CreateIntegrationCommentInput{Body: dto})

		require.NoError(t, err)
		assert.Equal(t, "c9", out.Data.ID)
		assert.Equal(t, model.StatusApproved, out.Data.Status)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().CreateIntegrationComment(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(model.CreateCommentResult{}, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.CreateIntegrationComment(bg(), &handler.CreateIntegrationCommentInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_ListPanelComments(t *testing.T) {
	hotCases := []struct {
		name    string
		raw     string
		wantHot *bool
	}{
		{"empty leaves hot unset", "", nil},
		{"true parses to pointer true", "true", new(true)},
		{"false parses to pointer false", "false", new(false)},
		{"unparseable leaves hot unset", "garbage", nil},
		{"whitespace-only leaves hot unset", "   ", nil},
	}
	for _, tc := range hotCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := commentmock.NewMockService(t)
			svc.EXPECT().
				ListPanelComments(mock.Anything, mock.MatchedBy(func(q model.ListCommentQuery) bool {
					return q.Page == 1 && q.PageSize == 20 && q.Keyword == "k" &&
						q.Status == "approved" && q.EchoID == "e1" && eqBoolPtr(q.Hot, tc.wantHot)
				})).
				Return(model.PageResult[model.Comment]{Total: 3}, nil).Once()

			h := handler.NewCommentHandler(svc)
			out, err := h.ListPanelComments(bg(), &handler.ListPanelCommentsInput{
				Page: 1, PageSize: 20, Keyword: "k", Status: "approved", EchoID: "e1", Hot: tc.raw,
			})

			require.NoError(t, err)
			assert.Equal(t, int64(3), out.Data.Total)
		})
	}

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().ListPanelComments(mock.Anything, mock.Anything).
			Return(model.PageResult[model.Comment]{}, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.ListPanelComments(bg(), &handler.ListPanelCommentsInput{Page: 1})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_GetCommentByID(t *testing.T) {
	t.Run("trims id before delegating", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().GetCommentByID(mock.Anything, "c1").
			Return(model.Comment{ID: "c1", Content: "x"}, nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.GetCommentByID(bg(), &handler.GetCommentByIDInput{ID: "  c1 "})

		require.NoError(t, err)
		assert.Equal(t, "c1", out.Data.ID)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().GetCommentByID(mock.Anything, mock.Anything).
			Return(model.Comment{}, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.GetCommentByID(bg(), &handler.GetCommentByIDInput{ID: "c1"})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, handler.CommentOutput{}, out)
	})
}

func TestCommentHandler_UpdateCommentStatus(t *testing.T) {
	t.Run("trims id and forwards status", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().UpdateCommentStatus(mock.Anything, "c1", model.StatusApproved).Return(nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.UpdateCommentStatus(bg(), &handler.UpdateCommentStatusInput{
			ID:   " c1 ",
			Body: model.UpdateCommentStatusDto{Status: model.StatusApproved},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().UpdateCommentStatus(mock.Anything, mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.UpdateCommentStatus(bg(), &handler.UpdateCommentStatusInput{
			ID:   "c1",
			Body: model.UpdateCommentStatusDto{Status: model.StatusRejected},
		})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_UpdateCommentHot(t *testing.T) {
	t.Run("trims id and forwards hot flag", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().UpdateCommentHot(mock.Anything, "c1", true).Return(nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.UpdateCommentHot(bg(), &handler.UpdateCommentHotInput{
			ID:   " c1 ",
			Body: model.UpdateCommentHotDto{Hot: true},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().UpdateCommentHot(mock.Anything, mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.UpdateCommentHot(bg(), &handler.UpdateCommentHotInput{
			ID:   "c1",
			Body: model.UpdateCommentHotDto{Hot: false},
		})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	t.Run("trims id and returns delete message", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().DeleteComment(mock.Anything, "c1").Return(nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.DeleteComment(bg(), &handler.DeleteCommentInput{ID: "  c1  "})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DELETE_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().DeleteComment(mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.DeleteComment(bg(), &handler.DeleteCommentInput{ID: "c1"})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_BatchAction(t *testing.T) {
	t.Run("forwards action and ids", func(t *testing.T) {
		ids := []string{"c1", "c2"}
		svc := commentmock.NewMockService(t)
		svc.EXPECT().BatchAction(mock.Anything, "approve", ids).Return(nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.BatchAction(bg(), &handler.BatchActionInput{
			Body: model.BatchCommentActionDto{Action: "approve", IDs: ids},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().BatchAction(mock.Anything, mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.BatchAction(bg(), &handler.BatchActionInput{
			Body: model.BatchCommentActionDto{Action: "delete", IDs: []string{"c1"}},
		})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_GetCommentSetting(t *testing.T) {
	t.Run("success returns system setting", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().GetSystemSetting(mock.Anything).
			Return(model.SystemSetting{EnableComment: true, RequireApproval: true}, nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.GetCommentSetting(bg(), &handler.GetCommentSettingInput{})

		require.NoError(t, err)
		assert.True(t, out.Data.EnableComment)
		assert.True(t, out.Data.RequireApproval)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().GetSystemSetting(mock.Anything).Return(model.SystemSetting{}, errBoom).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.GetCommentSetting(bg(), &handler.GetCommentSettingInput{})

		require.ErrorIs(t, err, errBoom)
		assert.Equal(t, handler.CommentSettingOutput{}, out)
	})
}

func TestCommentHandler_UpdateCommentSetting(t *testing.T) {
	t.Run("forwards setting and returns update message", func(t *testing.T) {
		setting := model.SystemSetting{EnableComment: true, CaptchaEnabled: true}
		svc := commentmock.NewMockService(t)
		svc.EXPECT().UpdateSystemSetting(mock.Anything, setting).Return(nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.UpdateCommentSetting(bg(), &handler.UpdateCommentSettingInput{Body: setting})

		require.NoError(t, err)
		assert.Equal(t, commonModel.UPDATE_SETTINGS_SUCCESS, out.Message)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().UpdateSystemSetting(mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.UpdateCommentSetting(bg(), &handler.UpdateCommentSettingInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func TestCommentHandler_TestCommentEmail(t *testing.T) {
	t.Run("forwards the nested setting", func(t *testing.T) {
		setting := model.SystemSetting{EmailNotify: model.EmailNotifySetting{Enabled: true, SMTPHost: "smtp.example.com"}}
		svc := commentmock.NewMockService(t)
		svc.EXPECT().SendTestEmail(mock.Anything, setting).Return(nil).Once()

		h := handler.NewCommentHandler(svc)
		out, err := h.TestCommentEmail(bg(), &handler.TestCommentEmailInput{
			Body: model.TestEmailRequest{Setting: setting},
		})

		require.NoError(t, err)
		assert.Equal(t, commonModel.DEFAULT_SUCCESS_CODE, out.Code)
	})

	t.Run("service error is propagated", func(t *testing.T) {
		svc := commentmock.NewMockService(t)
		svc.EXPECT().SendTestEmail(mock.Anything, mock.Anything).Return(errBoom).Once()

		h := handler.NewCommentHandler(svc)
		_, err := h.TestCommentEmail(bg(), &handler.TestCommentEmailInput{})

		require.ErrorIs(t, err, errBoom)
	})
}

func eqBoolPtr(a, b *bool) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
