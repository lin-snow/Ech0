// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"encoding/json"
	"testing"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	"github.com/lin-snow/ech0/internal/test/helpers"
	jwtUtil "github.com/lin-snow/ech0/internal/util/jwt"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetFormMeta(t *testing.T) {
	t.Run("setting read error is propagated", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commentModel.CommentSystemSettingKey).
			Return("", errRepoBoom).
			Once()
		_, err := d.service().GetFormMeta(helpers.CtxAnonymous(), testIP, "https://host")
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("captcha disabled still returns a signed token and flags", func(t *testing.T) {
		helpers.SetJWTSecret(t, testSecret)
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		meta, err := d.service().GetFormMeta(helpers.CtxAnonymous(), testIP, "https://host")
		require.NoError(t, err)
		assert.NotEmpty(t, meta.FormToken)
		assert.Equal(t, int64(2000), meta.MinSubmitMs)
		assert.False(t, meta.CaptchaEnabled)
		assert.True(t, meta.EnableComment)
		assert.Contains(t, meta.CaptchaAPIEndpoint, "/api/cap/")
	})

	t.Run("captcha ready when enabled with base url and secret", func(t *testing.T) {
		helpers.SetJWTSecret(t, testSecret)
		d := newDeps(t)
		s := enabledSetting()
		s.CaptchaEnabled = true
		d.expectSetting(t, s)
		meta, err := d.service().GetFormMeta(helpers.CtxAnonymous(), testIP, "https://host")
		require.NoError(t, err)
		assert.True(t, meta.CaptchaEnabled)
		assert.Contains(t, meta.CaptchaAPIEndpoint, "https://host/api/cap/")
	})
}

func TestListPublicByEchoID(t *testing.T) {
	t.Run("setting read error is propagated", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commentModel.CommentSystemSettingKey).
			Return("", errRepoBoom).
			Once()
		_, err := d.service().ListPublicByEchoID(context.Background(), "echo-1")
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("disabled comments returns empty slice without hitting repo", func(t *testing.T) {
		d := newDeps(t)
		s := enabledSetting()
		s.EnableComment = false
		d.expectSetting(t, s)
		got, err := d.service().ListPublicByEchoID(context.Background(), "echo-1")
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.repo.EXPECT().
			ListPublicByEchoID(mock.Anything, "echo-1").
			Return(nil, errRepoBoom).
			Once()
		_, err := d.service().ListPublicByEchoID(context.Background(), "echo-1")
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("success maps rows and trims echo id", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.repo.EXPECT().
			ListPublicByEchoID(mock.Anything, "echo-1").
			Return([]commentModel.Comment{
				{ID: "c-1", EchoID: "echo-1", Status: commentModel.StatusApproved, Nickname: "A"},
				{ID: "c-2", EchoID: "echo-1", Status: commentModel.StatusApproved, Nickname: "B"},
			}, nil).
			Once()
		got, err := d.service().ListPublicByEchoID(context.Background(), "  echo-1  ")
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "c-1", got[0].ID)
	})
}

func TestListPublicComments(t *testing.T) {
	t.Run("setting read error is propagated", func(t *testing.T) {
		d := newDeps(t)
		d.kv.EXPECT().
			Get(mock.Anything, commentModel.CommentSystemSettingKey).
			Return("", errRepoBoom).
			Once()
		_, err := d.service().ListPublicComments(context.Background(), 10)
		require.ErrorIs(t, err, errRepoBoom)
	})

	t.Run("disabled comments returns empty slice", func(t *testing.T) {
		d := newDeps(t)
		s := enabledSetting()
		s.EnableComment = false
		d.expectSetting(t, s)
		got, err := d.service().ListPublicComments(context.Background(), 10)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	limitCases := []struct {
		name     string
		in       int
		wantPass int
	}{
		{"zero defaults to 30", 0, 30},
		{"negative defaults to 30", -7, 30},
		{"within range passes through", 50, 50},
		{"oversized clamps to 100", 250, 100},
	}
	for _, tc := range limitCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			d.expectSetting(t, enabledSetting())
			var captured int
			d.repo.EXPECT().
				ListPublicComments(mock.Anything, mock.Anything).
				Run(func(_ context.Context, limit int) { captured = limit }).
				Return([]commentModel.Comment{}, nil).
				Once()
			_, err := d.service().ListPublicComments(context.Background(), tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.wantPass, captured)
		})
	}

	t.Run("repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.repo.EXPECT().
			ListPublicComments(mock.Anything, mock.Anything).
			Return(nil, errRepoBoom).
			Once()
		_, err := d.service().ListPublicComments(context.Background(), 10)
		require.ErrorIs(t, err, errRepoBoom)
	})
}

func TestGetCommentByID(t *testing.T) {
	t.Run("anonymous is denied before any IO", func(t *testing.T) {
		d := newDeps(t)
		_, err := d.service().GetCommentByID(helpers.CtxAnonymous(), "c-1")
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("admin gets the comment passed through", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		want := commentModel.Comment{ID: "c-1", Content: "hi"}
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(want, nil).
			Once()
		got, err := d.service().GetCommentByID(helpers.CtxAsUser("admin-1"), "c-1")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("repo error is propagated", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "c-1").
			Return(commentModel.Comment{}, errRepoBoom).
			Once()
		_, err := d.service().GetCommentByID(helpers.CtxAsUser("admin-1"), "c-1")
		require.ErrorIs(t, err, errRepoBoom)
	})
}

func TestUpdateSystemSetting(t *testing.T) {
	t.Run("anonymous is denied", func(t *testing.T) {
		d := newDeps(t)
		err := d.service().UpdateSystemSetting(helpers.CtxAnonymous(), commentModel.SystemSetting{})
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("applies port default and preserves blank password before persist", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		current := enabledSetting()
		current.EmailNotify.SMTPPassword = "old-secret"
		d.expectSetting(t, current)

		var persisted string
		d.kv.EXPECT().
			Set(mock.Anything, commentModel.CommentSystemSettingKey, mock.Anything).
			Run(func(_ context.Context, _ string, v string) { persisted = v }).
			Return(nil).
			Once()

		in := commentModel.SystemSetting{
			EnableComment: true,
			EmailNotify: commentModel.EmailNotifySetting{
				SMTPPort:     0,
				SMTPPassword: "",
			},
		}
		err := d.service().UpdateSystemSetting(helpers.CtxAsUser("admin-1"), in)
		require.NoError(t, err)

		var saved commentModel.SystemSetting
		require.NoError(t, json.Unmarshal([]byte(persisted), &saved))
		assert.Equal(t, 587, saved.EmailNotify.SMTPPort)
		assert.Equal(t, "old-secret", saved.EmailNotify.SMTPPassword)
	})

	t.Run("kv set error is propagated", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.expectSetting(t, enabledSetting())
		d.kv.EXPECT().
			Set(mock.Anything, commentModel.CommentSystemSettingKey, mock.Anything).
			Return(errRepoBoom).
			Once()
		in := commentModel.SystemSetting{
			EmailNotify: commentModel.EmailNotifySetting{SMTPPort: 25, SMTPPassword: "keep"},
		}
		err := d.service().UpdateSystemSetting(helpers.CtxAsUser("admin-1"), in)
		require.ErrorIs(t, err, errRepoBoom)
	})
}

func validEmailNotify() commentModel.EmailNotifySetting {
	return commentModel.EmailNotifySetting{
		Enabled:      true,
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "secret",
		SMTPSender:   "noreply@example.com",
	}
}

func TestSendTestEmail(t *testing.T) {
	t.Run("anonymous is denied", func(t *testing.T) {
		d := newDeps(t)
		err := d.service().SendTestEmail(helpers.CtxAnonymous(), commentModel.SystemSetting{})
		assertBiz(t, err, commonModel.ErrCodePermissionDenied, commonModel.NO_PERMISSION_DENIED)
	})

	t.Run("missing owner email fails before sending", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.expectSetting(t, enabledSetting())
		d.common.EXPECT().
			GetOwner().
			Return(helpers.NewUser(), nil).
			Once()
		err := d.service().SendTestEmail(helpers.CtxAsUser("admin-1"), commentModel.SystemSetting{
			EmailNotify: validEmailNotify(),
		})
		require.Error(t, err)
	})

	t.Run("invalid smtp config returns biz error", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.expectSetting(t, enabledSetting())
		owner := helpers.NewUser()
		owner.Email = "owner@example.com"
		d.common.EXPECT().GetOwner().Return(owner, nil).Once()
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.ServerURLKey).
			Return("https://example.com", nil)
		bad := validEmailNotify()
		bad.SMTPHost = ""
		err := d.service().SendTestEmail(helpers.CtxAsUser("admin-1"), commentModel.SystemSetting{
			EmailNotify: bad,
		})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "")
	})

	t.Run("happy path delivers via mailer", func(t *testing.T) {
		d := newDeps(t)
		expectAdmin(t, d, "admin-1")
		d.expectSetting(t, enabledSetting())
		owner := helpers.NewUser()
		owner.Email = "owner@example.com"
		d.common.EXPECT().GetOwner().Return(owner, nil).Once()
		d.kv.EXPECT().
			Get(mock.Anything, commonModel.ServerURLKey).
			Return("https://example.com", nil)
		d.mailer.EXPECT().
			Send(mock.Anything, mock.Anything, mock.Anything).
			Return(nil).
			Once()
		err := d.service().SendTestEmail(helpers.CtxAsUser("admin-1"), commentModel.SystemSetting{
			EmailNotify: validEmailNotify(),
		})
		require.NoError(t, err)
	})
}

func TestParseOptionalUserIDFromAuthHeader(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	user := helpers.NewUser()
	user.ID = "u-42"
	token, err := jwtUtil.GenerateToken(jwtUtil.CreateClaims(user))
	require.NoError(t, err)

	cases := []struct {
		name   string
		header string
		want   string
	}{
		{"valid bearer token yields user id", "Bearer " + token, "u-42"},
		{"bearer scheme is case-insensitive", "bearer " + token, "u-42"},
		{"empty header yields empty", "", ""},
		{"non-bearer scheme yields empty", "Basic abc123", ""},
		{"bearer without token yields empty", "Bearer", ""},
		{"malformed token yields empty", "Bearer not.a.jwt", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := commentService.ParseOptionalUserIDFromAuthHeader(tc.header)
			assert.Equal(t, tc.want, got)
		})
	}
}
