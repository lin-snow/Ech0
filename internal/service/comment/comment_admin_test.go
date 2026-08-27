// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"errors"
	"testing"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errRepoBoom = errors.New("repo boom")

func expectAdmin(t *testing.T, d deps, userID string) {
	t.Helper()
	owner := helpers.NewUser(helpers.AsOwner)
	owner.ID = userID
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, userID).
		Return(owner, nil).
		Once()
}

func TestRequireAdminGuard(t *testing.T) {
	t.Run("anonymous is denied before any IO", func(t *testing.T) {
		d := newDeps(t)
		err := d.service().UpdateCommentStatus(helpers.CtxAnonymous(), "c-1", commentModel.StatusApproved)
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("non-admin user is denied", func(t *testing.T) {
		d := newDeps(t)
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, "user-normal").
			Return(helpers.NewUser(), nil).
			Once()
		err := d.service().UpdateCommentStatus(helpers.CtxAsUser("user-normal"), "c-1", commentModel.StatusApproved)
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("user lookup error is propagated verbatim", func(t *testing.T) {
		d := newDeps(t)
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, "admin-1").
			Return(helpers.NewUser(), errRepoBoom).
			Once()
		err := d.service().UpdateCommentStatus(helpers.CtxAsUser("admin-1"), "c-1", commentModel.StatusApproved)
		require.ErrorIs(t, err, errRepoBoom)
	})
}

func TestUpdateCommentStatus(t *testing.T) {
	t.Run("invalid status is rejected after admin check", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		err := d.service().UpdateCommentStatus(helpers.CtxAsUser("admin-1"), "c-1", commentModel.Status("weird"))
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "无效的评论状态")
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			UpdateCommentStatus(mock.Anything, "c-1", commentModel.StatusRejected).
			Return(errRepoBoom).
			Once()
		err := d.service().UpdateCommentStatus(helpers.CtxAsUser("admin-1"), "c-1", commentModel.StatusRejected)
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("success re-reads comment then emits and returns nil", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			UpdateCommentStatus(mock.Anything, "c-1", commentModel.StatusApproved).
			Return(nil).
			Once()
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{ID: "c-1", Status: commentModel.StatusApproved}, nil).
			Once()
		d.expectSetting(t, enabledSetting())

		err := d.service().UpdateCommentStatus(helpers.CtxAsUser("admin-1"), "c-1", commentModel.StatusApproved)
		require.NoError(t, err)
	})

	t.Run("success but re-read returns empty skips emit and notify", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			UpdateCommentStatus(mock.Anything, "c-1", commentModel.StatusApproved).
			Return(nil).
			Once()
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{}, nil).
			Once()

		err := d.service().UpdateCommentStatus(helpers.CtxAsUser("admin-1"), "c-1", commentModel.StatusApproved)
		require.NoError(t, err)
	})
}

func TestUpdateCommentHot(t *testing.T) {
	t.Run("anonymous is denied", func(t *testing.T) {
		d := newDeps(t)
		err := d.service().UpdateCommentHot(helpers.CtxAnonymous(), "c-1", true)
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			UpdateCommentHot(mock.Anything, "c-1", true).
			Return(errRepoBoom).
			Once()
		err := d.service().UpdateCommentHot(helpers.CtxAsUser("admin-1"), "c-1", true)
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("hot=true re-reads comment for notify", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			UpdateCommentHot(mock.Anything, "c-1", true).
			Return(nil).
			Once()
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{ID: "c-1"}, nil).
			Once()
		d.expectSetting(t, enabledSetting())

		err := d.service().UpdateCommentHot(helpers.CtxAsUser("admin-1"), "c-1", true)
		require.NoError(t, err)
	})

	t.Run("hot=false skips re-read and notify", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			UpdateCommentHot(mock.Anything, "c-1", false).
			Return(nil).
			Once()
		err := d.service().UpdateCommentHot(helpers.CtxAsUser("admin-1"), "c-1", false)
		require.NoError(t, err)
	})
}

func TestDeleteComment(t *testing.T) {
	t.Run("anonymous is denied", func(t *testing.T) {
		d := newDeps(t)
		err := d.service().DeleteComment(helpers.CtxAnonymous(), "c-1")
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("repo delete error is propagated", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{ID: "c-1"}, nil).
			Once()
		d.repo.EXPECT().
			DeleteComment(mock.Anything, "c-1").
			Return(errRepoBoom).
			Once()
		err := d.service().DeleteComment(helpers.CtxAsUser("admin-1"), "c-1")
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("success with existing comment emits deleted event", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{ID: "c-1"}, nil).
			Once()
		d.repo.EXPECT().
			DeleteComment(mock.Anything, "c-1").
			Return(nil).
			Once()
		err := d.service().DeleteComment(helpers.CtxAsUser("admin-1"), "c-1")
		require.NoError(t, err)
	})

	t.Run("success when comment missing skips emit", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "missing").
			Return(commentModel.Comment{}, errRepoBoom).
			Once()
		d.repo.EXPECT().
			DeleteComment(mock.Anything, "missing").
			Return(nil).
			Once()
		err := d.service().DeleteComment(helpers.CtxAsUser("admin-1"), "missing")
		require.NoError(t, err)
	})
}

func TestBatchAction(t *testing.T) {
	t.Run("anonymous is denied", func(t *testing.T) {
		d := newDeps(t)
		err := d.service().BatchAction(helpers.CtxAnonymous(), "approve", []string{"c-1"})
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("empty ids is a no-op after admin check", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		err := d.service().BatchAction(helpers.CtxAsUser("admin-1"), "approve", nil)
		require.NoError(t, err)
	})

	t.Run("unknown action is rejected", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		err := d.service().BatchAction(helpers.CtxAsUser("admin-1"), "explode", []string{"c-1"})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "无效的批量动作")
	})

	statusCases := []struct {
		name       string
		action     string
		wantStatus commentModel.Status
	}{
		{"approve maps to approved status", "approve", commentModel.StatusApproved},
		{"reject maps to rejected status", "reject", commentModel.StatusRejected},
	}
	for _, tc := range statusCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			ids := []string{"c-1", "c-2"}
			expectAdmin(t, d, "admin-1")
			d.repo.EXPECT().
				BatchUpdateStatus(mock.Anything, ids, tc.wantStatus).
				Return(nil).
				Once()
			for _, id := range ids {
				d.repo.EXPECT().
					GetCommentByID(mock.Anything, id).
					Return(commentModel.Comment{ID: id, Status: tc.wantStatus}, nil).
					Once()
			}
			d.expectSetting(t, enabledSetting())

			err := d.service().BatchAction(helpers.CtxAsUser("admin-1"), tc.action, ids)
			require.NoError(t, err)
		})
	}

	t.Run("batch status repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			BatchUpdateStatus(mock.Anything, []string{"c-1"}, commentModel.StatusApproved).
			Return(errRepoBoom).
			Once()
		err := d.service().BatchAction(helpers.CtxAsUser("admin-1"), "approve", []string{"c-1"})
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("delete collects payloads then batch deletes", func(t *testing.T) {
		d := newDeps(t)
		ids := []string{"c-1", "c-2"}
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{ID: "c-1"}, nil).
			Once()
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-2").
			Return(commentModel.Comment{}, nil).
			Once()
		d.repo.EXPECT().
			BatchDelete(mock.Anything, ids).
			Return(nil).
			Once()
		err := d.service().BatchAction(helpers.CtxAsUser("admin-1"), "delete", ids)
		require.NoError(t, err)
	})

	t.Run("batch delete repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		ids := []string{"c-1"}
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{ID: "c-1"}, nil).
			Once()
		d.repo.EXPECT().
			BatchDelete(mock.Anything, ids).
			Return(errRepoBoom).
			Once()
		err := d.service().BatchAction(helpers.CtxAsUser("admin-1"), "delete", ids)
		require.ErrorIs(t, err, errRepoBoom)
	})
}

func TestListPanelComments(t *testing.T) {
	t.Run("anonymous is denied", func(t *testing.T) {
		d := newDeps(t)
		_, err := d.service().ListPanelComments(helpers.CtxAnonymous(), commentModel.ListCommentQuery{})
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	normalizeCases := []struct {
		name         string
		inPage       int
		inSize       int
		wantPage     int
		wantPageSize int
	}{
		{"zero page and size default to 1/20", 0, 0, 1, 20},
		{"negative page and size default to 1/20", -3, -1, 1, 20},
		{"oversized page size clamps to 20", 2, 500, 2, 20},
		{"valid query passes through unchanged", 3, 50, 3, 50},
	}
	for _, tc := range normalizeCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			expectAdmin(t, d, "admin-1")
			var captured commentModel.ListCommentQuery
			want := commentModel.PageResult[commentModel.Comment]{
				Items: []commentModel.Comment{{ID: "c-1"}},
				Total: 1,
			}
			d.repo.EXPECT().
				ListComments(mock.Anything, mock.Anything).
				Run(func(_ context.Context, q commentModel.ListCommentQuery) { captured = q }).
				Return(want, nil).
				Once()

			got, err := d.service().ListPanelComments(
				helpers.CtxAsUser("admin-1"),
				commentModel.ListCommentQuery{Page: tc.inPage, PageSize: tc.inSize, Keyword: "kw"},
			)
			require.NoError(t, err)
			assert.Equal(t, tc.wantPage, captured.Page)
			assert.Equal(t, tc.wantPageSize, captured.PageSize)
			assert.Equal(t, "kw", captured.Keyword, "non-pagination fields must be preserved")
			assert.Equal(t, want, got, "repo result is returned verbatim")
		})
	}

	t.Run("repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			ListComments(mock.Anything, mock.Anything).
			Return(commentModel.PageResult[commentModel.Comment]{}, errRepoBoom).
			Once()
		_, err := d.service().ListPanelComments(helpers.CtxAsUser("admin-1"), commentModel.ListCommentQuery{Page: 1, PageSize: 10})
		require.ErrorIs(t, err, errRepoBoom)
	})
}

func TestResolveParentID_ValidParentSetsParentID(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())

	owner := helpers.NewUser(helpers.AsOwner)
	owner.ID = "owner-1"
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "owner-1").
		Return(owner, nil).
		Once()
	d.repo.EXPECT().
		GetCommentByID(mock.Anything, "parent-1").
		Return(commentModel.Comment{
			ID:     "parent-1",
			EchoID: "echo-1",
			Status: commentModel.StatusApproved,
		}, nil).
		Once()
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()

	var captured commentModel.Comment
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) {
			c.ID = "child-1"
			captured = *c
		}).
		Return(nil).
		Once()

	res, err := d.service().CreateComment(helpers.CtxAsUser("owner-1"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "a reply",
			ParentID:  "parent-1",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "child-1", res.ID)
	require.NotNil(t, captured.ParentID)
	assert.Equal(t, "parent-1", *captured.ParentID)
}
