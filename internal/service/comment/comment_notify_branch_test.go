// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"testing"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mailEnabledSetting() commentModel.SystemSetting {
	s := enabledSetting()
	s.EmailNotify.Enabled = true
	return s
}

func TestUpdateCommentStatus_NotifySkipsInvalidRecipient(t *testing.T) {
	d := newDeps(t)
	expectAdmin(t, d, "admin-1")
	d.repo.EXPECT().
		UpdateCommentStatus(mock.Anything, "c-1", commentModel.StatusRejected).
		Return(nil).
		Once()
	d.repo.EXPECT().
		GetCommentByID(mock.Anything, "c-1").
		Return(commentModel.Comment{ID: "c-1", Status: commentModel.StatusRejected, Email: ""}, nil).
		Once()
	d.expectSetting(t, mailEnabledSetting())

	err := d.service().UpdateCommentStatus(helpers.CtxAsUser("admin-1"), "c-1", commentModel.StatusRejected)
	require.NoError(t, err)
}

func TestCreateComment_NotifyOwnerSkipsWhenOwnerEmailMissing(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	s := mailEnabledSetting()
	s.RequireApproval = false
	d.expectSetting(t, s)
	d.repo.EXPECT().
		CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil)
	d.repo.EXPECT().
		CountByEmailWithin(mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil)
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) { c.ID = "g-1" }).
		Return(nil).
		Once()
	d.common.EXPECT().GetOwner().Return(helpers.NewUser(), nil).Once()

	res, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello",
			Nickname:  "Guest",
			Email:     "guest@example.com",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "g-1", res.ID)
}

func TestCreateComment_ReplyNotifySkipsInvalidParentEmail(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, mailEnabledSetting())

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
			Email:  "",
		}, nil).
		Times(2)
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) { c.ID = "child-1" }).
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
}

func TestCreateComment_ReplyNotifySkipsWhenTargetIsOwner(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, mailEnabledSetting())

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
			Email:  "shared@example.com",
		}, nil).
		Times(2)
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) { c.ID = "child-2" }).
		Return(nil).
		Once()
	ownerWithEmail := helpers.NewUser(helpers.AsOwner)
	ownerWithEmail.Email = "shared@example.com"
	d.common.EXPECT().GetOwner().Return(ownerWithEmail, nil).Once()

	res, err := d.service().CreateComment(helpers.CtxAsUser("owner-1"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "a reply",
			ParentID:  "parent-1",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "child-2", res.ID)
}
