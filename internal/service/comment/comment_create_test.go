// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	commentService "github.com/lin-snow/ech0/internal/service/comment"
	"github.com/lin-snow/ech0/internal/test/helpers"
	commentmock "github.com/lin-snow/ech0/internal/test/mocks/commentmock"
	commonmock "github.com/lin-snow/ech0/internal/test/mocks/commonmock"
	kvmock "github.com/lin-snow/ech0/internal/test/mocks/kvmock"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testSecret = "comment-create-test-secret"
	testIP     = "203.0.113.5"
)

type deps struct {
	repo   *commentmock.MockRepository
	kv     *kvmock.MockStore
	common *commonmock.MockService
	mailer *commentmock.MockMailer
}

func newDeps(t *testing.T) deps {
	t.Helper()
	return deps{
		repo:   commentmock.NewMockRepository(t),
		kv:     kvmock.NewMockStore(t),
		common: commonmock.NewMockService(t),
		mailer: commentmock.NewMockMailer(t),
	}
}

func (d deps) service() *commentService.CommentService {
	return commentService.NewCommentService(
		d.common,
		d.repo,
		d.kv,
		func() *busen.Bus { return busen.New() },
		d.mailer,
	)
}

func (d deps) expectSetting(t *testing.T, s commentModel.SystemSetting) {
	t.Helper()
	buf, err := json.Marshal(s)
	require.NoError(t, err)
	d.kv.EXPECT().
		Get(mock.Anything, commentModel.CommentSystemSettingKey).
		Return(string(buf), nil)
}

func enabledSetting() commentModel.SystemSetting {
	return commentModel.SystemSetting{
		EnableComment:   true,
		RequireApproval: true,
		CaptchaEnabled:  false,
		EmailNotify:     commentModel.EmailNotifySetting{Enabled: false},
	}
}

func signFormToken(ip string, issuedAt int64) string {
	mac := hmac.New(sha256.New, []byte(testSecret))
	_, _ = fmt.Fprintf(mac, "%s:%d", ip, issuedAt)
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("%d.%s", issuedAt, sig)
}

func freshToken() string {
	return signFormToken(testIP, time.Now().UnixMilli()-5000)
}

func assertBiz(t *testing.T, err error, wantCode, wantMsg string) {
	t.Helper()
	require.Error(t, err)
	var be *commonModel.BizError
	require.ErrorAs(t, err, &be)
	assert.Equal(t, wantCode, be.Code)
	if wantMsg != "" {
		assert.Equal(t, wantMsg, be.Msg)
	}
}

func TestCreateComment_GuardRails(t *testing.T) {
	t.Run("honeypot filled is rejected before any IO", func(t *testing.T) {
		d := newDeps(t)
		_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:        "echo-1",
				Content:       "hi",
				HoneypotField: "i-am-a-bot",
				FormToken:     "whatever",
			})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "提交被拒绝")
	})

	t.Run("invalid form token is rejected", func(t *testing.T) {
		d := newDeps(t)
		_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-1",
				Content:   "hi",
				FormToken: "garbage-token",
			})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "提交过快或表单已失效")
	})

	t.Run("comment feature disabled", func(t *testing.T) {
		helpers.SetJWTSecret(t, testSecret)
		d := newDeps(t)
		s := enabledSetting()
		s.EnableComment = false
		d.expectSetting(t, s)
		_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-1",
				Content:   "hi",
				FormToken: freshToken(),
			})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "评论功能未启用")
	})

	t.Run("captcha enabled but token invalid", func(t *testing.T) {
		helpers.SetJWTSecret(t, testSecret)
		d := newDeps(t)
		s := enabledSetting()
		s.CaptchaEnabled = true
		d.expectSetting(t, s)
		_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:       "echo-1",
				Content:      "hi",
				FormToken:    freshToken(),
				CaptchaToken: "",
			})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "验证码验证失败")
	})
}

func TestCreateComment_GuestValidation(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)

	cases := []struct {
		name    string
		dto     commentModel.CreateCommentDto
		wantMsg string
	}{
		{
			name:    "empty content",
			dto:     commentModel.CreateCommentDto{EchoID: "echo-1", Content: "   "},
			wantMsg: "评论内容不能为空",
		},
		{
			name:    "content too long",
			dto:     commentModel.CreateCommentDto{EchoID: "echo-1", Content: strings.Repeat("a", 201)},
			wantMsg: "评论内容不能超过200字",
		},
		{
			name:    "missing nickname and email",
			dto:     commentModel.CreateCommentDto{EchoID: "echo-1", Content: "hello"},
			wantMsg: "昵称和邮箱不能为空",
		},
		{
			name: "invalid email",
			dto: commentModel.CreateCommentDto{
				EchoID: "echo-1", Content: "hello", Nickname: "Bob", Email: "not-an-email",
			},
			wantMsg: "邮箱格式无效",
		},
		{
			name: "invalid website",
			dto: commentModel.CreateCommentDto{
				EchoID: "echo-1", Content: "hello", Nickname: "Bob", Email: "bob@example.com", Website: "notaurl",
			},
			wantMsg: "网址格式无效",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			d.expectSetting(t, enabledSetting())
			dto := tc.dto
			dto.FormToken = freshToken()
			_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua", &dto)
			assertBiz(t, err, commonModel.ErrCodeInvalidRequest, tc.wantMsg)
		})
	}
}

func TestCreateComment_NonAdminUserTreatedAsGuest(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "user-normal").
		Return(helpers.NewUser(), nil).
		Once()

	_, err := d.service().CreateComment(helpers.CtxAsUser("user-normal"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello",
			FormToken: freshToken(),
		})
	assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "昵称和邮箱不能为空")
}

func TestCreateComment_RateLimitExceeded(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())
	d.repo.EXPECT().
		CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
		Return(int64(3), nil)

	_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello",
			Nickname:  "Bob",
			Email:     "bob@example.com",
			FormToken: freshToken(),
		})
	assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "评论过于频繁，请稍后再试")
}

func TestCreateComment_DuplicateRejected(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "admin-1").
		Return(helpers.NewUser(helpers.AsAdmin), nil).
		Once()
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(true, nil).
		Once()

	_, err := d.service().CreateComment(helpers.CtxAsUser("admin-1"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello",
			FormToken: freshToken(),
		})
	assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "请勿重复提交相同评论")
}

func TestCreateComment_ParentResolution(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)

	t.Run("parent not found", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, "admin-1").
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "missing-parent").
			Return(commentModel.Comment{}, nil).
			Once()

		_, err := d.service().CreateComment(helpers.CtxAsUser("admin-1"), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-1",
				Content:   "hello",
				ParentID:  "missing-parent",
				FormToken: freshToken(),
			})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "回复的评论不存在")
	})

	t.Run("parent belongs to a different echo", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, "admin-1").
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "parent-1").
			Return(commentModel.Comment{
				ID:     "parent-1",
				EchoID: "other-echo",
				Status: commentModel.StatusApproved,
			}, nil).
			Once()

		_, err := d.service().CreateComment(helpers.CtxAsUser("admin-1"), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-1",
				Content:   "hello",
				ParentID:  "parent-1",
				FormToken: freshToken(),
			})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "回复的评论不存在")
	})

	t.Run("parent not approved", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.common.EXPECT().
			CommonGetUserByUserId(mock.Anything, "admin-1").
			Return(helpers.NewUser(helpers.AsAdmin), nil).
			Once()
		d.repo.EXPECT().
			GetCommentByID(mock.Anything, "parent-1").
			Return(commentModel.Comment{
				ID:     "parent-1",
				EchoID: "echo-1",
				Status: commentModel.StatusPending,
			}, nil).
			Once()

		_, err := d.service().CreateComment(helpers.CtxAsUser("admin-1"), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-1",
				Content:   "hello",
				ParentID:  "parent-1",
				FormToken: freshToken(),
			})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "该评论暂不可回复")
	})
}

func TestCreateComment_AdminAutoApprove(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())

	owner := helpers.NewUser(helpers.AsOwner)
	owner.ID = "owner-1"
	owner.Username = "OwnerBob"
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "owner-1").
		Return(owner, nil).
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
			c.ID = "new-comment-1"
			captured = *c
		}).
		Return(nil).
		Once()

	res, err := d.service().CreateComment(helpers.CtxAsUser("owner-1"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello from owner",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "new-comment-1", res.ID)
	assert.Equal(t, commentModel.StatusApproved, res.Status)
	assert.Equal(t, commentModel.SourceSystem, captured.Source)
	assert.Equal(t, commentModel.StatusApproved, captured.Status)
	assert.Equal(t, "OwnerBob", captured.Nickname)
	assert.Empty(t, captured.Email)
	require.NotNil(t, captured.UserID)
	assert.Equal(t, "owner-1", *captured.UserID)
}

func TestCreateComment_GuestHappy(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)

	cases := []struct {
		name            string
		requireApproval bool
		wantStatus      commentModel.Status
	}{
		{"require approval -> pending", true, commentModel.StatusPending},
		{"auto approve -> approved", false, commentModel.StatusApproved},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			s := enabledSetting()
			s.RequireApproval = tc.requireApproval
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

			var captured commentModel.Comment
			d.repo.EXPECT().
				CreateComment(mock.Anything, mock.Anything).
				Run(func(_ context.Context, c *commentModel.Comment) {
					c.ID = "guest-cmt"
					captured = *c
				}).
				Return(nil).
				Once()

			res, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
				&commentModel.CreateCommentDto{
					EchoID:    "echo-1",
					Content:   "hello",
					Nickname:  "Guest",
					Email:     "guest@example.com",
					FormToken: freshToken(),
				})
			require.NoError(t, err)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, commentModel.SourceGuest, captured.Source)
			assert.Equal(t, tc.wantStatus, captured.Status)
			assert.Equal(t, "Guest", captured.Nickname)
			assert.Equal(t, "guest@example.com", captured.Email)
			assert.Nil(t, captured.UserID)
		})
	}
}

func integrationCtx() context.Context {
	return helpers.CtxAsToken(
		"user-9", "access",
		[]string{"comment:write"}, []string{"integration"}, "jti-9",
	)
}

func TestCreateIntegrationComment_GuardRails(t *testing.T) {
	t.Run("comment disabled", func(t *testing.T) {
		d := newDeps(t)
		s := enabledSetting()
		s.EnableComment = false
		d.expectSetting(t, s)
		_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
			&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: "hi"})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "评论功能未启用")
	})

	t.Run("empty content", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
			&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: "   "})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "评论内容不能为空")
	})

	t.Run("content too long", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
			&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: strings.Repeat("z", 201)})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "评论内容不能超过200字")
	})
}

func TestCreateIntegrationComment_RateLimit(t *testing.T) {
	t.Run("ip short window exceeded", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.repo.EXPECT().
			CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
			Return(int64(5), nil)
		_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
			&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: "hi"})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "集成评论过于频繁，请稍后再试")
	})

	t.Run("user short window exceeded", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.repo.EXPECT().
			CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
			Return(int64(0), nil)
		d.repo.EXPECT().
			CountByUserWithin(mock.Anything, "user-9", mock.Anything).
			Return(int64(5), nil).
			Once()
		_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
			&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: "hi"})
		assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "集成评论过于频繁，请稍后再试")
	})
}

func TestCreateIntegrationComment_DuplicateRejected(t *testing.T) {
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())
	d.repo.EXPECT().
		CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil)
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(true, nil).
		Once()

	_, err := d.service().CreateIntegrationComment(helpers.CtxAnonymous(), testIP, "ua",
		&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: "hi"})
	assertBiz(t, err, commonModel.ErrCodeInvalidRequest, "请勿重复提交相同评论")
}

func TestCreateIntegrationComment_Happy(t *testing.T) {
	cases := []struct {
		name            string
		requireApproval bool
		nickname        string
		wantNickname    string
		wantStatus      commentModel.Status
	}{
		{"require approval, default nickname", true, "", "Integration", commentModel.StatusPending},
		{"auto approve, custom nickname", false, "MyBot", "MyBot", commentModel.StatusApproved},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			s := enabledSetting()
			s.RequireApproval = tc.requireApproval
			d.expectSetting(t, s)
			d.repo.EXPECT().
				CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
				Return(int64(0), nil)
			d.repo.EXPECT().
				CountByUserWithin(mock.Anything, "user-9", mock.Anything).
				Return(int64(0), nil).
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
					c.ID = "int-cmt"
					captured = *c
				}).
				Return(nil).
				Once()

			res, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
				&commentModel.CreateIntegrationCommentDto{
					EchoID:   "echo-1",
					Content:  "hi from bot",
					Nickname: tc.nickname,
					Metadata: "src=test",
				})
			require.NoError(t, err)
			assert.Equal(t, "int-cmt", res.ID)
			assert.Equal(t, tc.wantStatus, res.Status)
			assert.Equal(t, commentModel.SourceIntegration, captured.Source)
			assert.Equal(t, tc.wantStatus, captured.Status)
			assert.Equal(t, tc.wantNickname, captured.Nickname)
			require.NotNil(t, captured.UserID)
			assert.Equal(t, "user-9", *captured.UserID)
		})
	}
}
