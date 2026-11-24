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
	"net/http"
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

	// privateEchos / missingEchos describe the echo table the commentability
	// guard sees. Echo ids absent from both maps are public.
	privateEchos map[string]bool
	missingEchos map[string]error
}

func newDeps(t *testing.T) deps {
	t.Helper()
	repo := commentmock.NewMockRepository(t)
	d := deps{
		repo:         repo,
		kv:           kvmock.NewMockStore(t),
		common:       commonmock.NewMockService(t),
		mailer:       commentmock.NewMockMailer(t),
		privateEchos: make(map[string]bool),
		missingEchos: make(map[string]error),
	}
	repo.EXPECT().
		IsEchoPrivate(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, echoID string) (bool, error) {
			if err, ok := d.missingEchos[echoID]; ok {
				return false, err
			}
			return d.privateEchos[echoID], nil
		}).
		Maybe()
	return d
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

func assertFail(t *testing.T, err error, wantStatus int, wantMsg string) {
	t.Helper()
	require.Error(t, err)
	var se *commonModel.StatusError
	require.ErrorAs(t, err, &se)
	assert.Equal(t, wantStatus, se.Status)
	if wantMsg != "" {
		assert.Equal(t, wantMsg, se.Msg)
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
		assertFail(t, err, http.StatusBadRequest, "submission rejected")
	})

	t.Run("invalid form token is rejected", func(t *testing.T) {
		d := newDeps(t)
		_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-1",
				Content:   "hi",
				FormToken: "garbage-token",
			})
		assertFail(t, err, http.StatusBadRequest, "submitted too fast or the form expired")
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
		assertFail(t, err, http.StatusBadRequest, "comments are disabled")
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
		assertFail(t, err, http.StatusBadRequest, "captcha verification failed")
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
			wantMsg: "comment content must not be empty",
		},
		{
			name:    "content too long",
			dto:     commentModel.CreateCommentDto{EchoID: "echo-1", Content: strings.Repeat("a", 201)},
			wantMsg: "comment content must be at most 200 characters",
		},
		{
			name:    "missing nickname and email",
			dto:     commentModel.CreateCommentDto{EchoID: "echo-1", Content: "hello"},
			wantMsg: "nickname and email are required",
		},
		{
			name: "invalid email",
			dto: commentModel.CreateCommentDto{
				EchoID: "echo-1", Content: "hello", Nickname: "Bob", Email: "not-an-email",
			},
			wantMsg: "invalid email address",
		},
		{
			name: "invalid website",
			dto: commentModel.CreateCommentDto{
				EchoID: "echo-1", Content: "hello", Nickname: "Bob", Email: "bob@example.com", Website: "notaurl",
			},
			wantMsg: "invalid website url",
		},
		{
			// ParseRequestURI accepts any scheme with a host, so a
			// javascript: URL used to pass and then ran on author-link click.
			name: "javascript website scheme",
			dto: commentModel.CreateCommentDto{
				EchoID: "echo-1", Content: "hello", Nickname: "Bob", Email: "bob@example.com",
				Website: "javascript://evil.example/%0aalert(document.domain)",
			},
			wantMsg: "invalid website url",
		},
		{
			name: "data website scheme",
			dto: commentModel.CreateCommentDto{
				EchoID: "echo-1", Content: "hello", Nickname: "Bob", Email: "bob@example.com",
				Website: "data:text/html,<script>alert(1)</script>",
			},
			wantMsg: "invalid website url",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDeps(t)
			d.expectSetting(t, enabledSetting())
			dto := tc.dto
			dto.FormToken = freshToken()
			_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua", &dto)
			assertFail(t, err, http.StatusBadRequest, tc.wantMsg)
		})
	}
}

func TestCreateComment_MemberAutoApprove(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())

	member := helpers.NewUser()
	member.ID = "user-normal"
	member.Username = "Bob"
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "user-normal").
		Return(member, nil).
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

	res, err := d.service().CreateComment(helpers.CtxAsUser("user-normal"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, commentModel.StatusApproved, res.Status)
	assert.Equal(t, commentModel.SourceSystem, captured.Source)
	assert.Equal(t, "Bob", captured.Nickname)
	assert.Empty(t, captured.Email)
	require.NotNil(t, captured.UserID)
	assert.Equal(t, "user-normal", *captured.UserID)
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
	assertFail(t, err, http.StatusBadRequest, "comment rate limit exceeded")
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
	assertFail(t, err, http.StatusBadRequest, "duplicate comment")
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
		assertFail(t, err, http.StatusBadRequest, "the comment being replied to does not exist")
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
		assertFail(t, err, http.StatusBadRequest, "the comment being replied to does not exist")
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
		assertFail(t, err, http.StatusBadRequest, "this comment cannot be replied to")
	})
}

func TestCreateComment_AdminAutoApprove(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())

	admin := helpers.NewUser(helpers.AsAdmin)
	admin.ID = "admin-1"
	admin.Username = "AdminBob"
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "admin-1").
		Return(admin, nil).
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

	res, err := d.service().CreateComment(helpers.CtxAsUser("admin-1"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-1",
			Content:   "hello from admin",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "new-comment-1", res.ID)
	assert.Equal(t, commentModel.StatusApproved, res.Status)
	assert.Equal(t, commentModel.SourceSystem, captured.Source)
	assert.Equal(t, commentModel.StatusApproved, captured.Status)
	assert.Equal(t, "AdminBob", captured.Nickname)
	assert.Empty(t, captured.Email)
	require.NotNil(t, captured.UserID)
	assert.Equal(t, "admin-1", *captured.UserID)
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
		assertFail(t, err, http.StatusBadRequest, "comments are disabled")
	})

	t.Run("empty content", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
			&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: "   "})
		assertFail(t, err, http.StatusBadRequest, "comment content must not be empty")
	})

	t.Run("content too long", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
			&commentModel.CreateIntegrationCommentDto{EchoID: "echo-1", Content: strings.Repeat("z", 201)})
		assertFail(t, err, http.StatusBadRequest, "comment content must be at most 200 characters")
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
		assertFail(t, err, http.StatusBadRequest, "integration comment rate limit exceeded")
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
		assertFail(t, err, http.StatusBadRequest, "integration comment rate limit exceeded")
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
	assertFail(t, err, http.StatusBadRequest, "duplicate comment")
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

func TestCreateComment_RejectsInvisibleEcho(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)

	t.Run("private echo", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.repo.EXPECT().
			CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
			Return(int64(0), nil)
		d.repo.EXPECT().
			CountByEmailWithin(mock.Anything, mock.Anything, mock.Anything).
			Return(int64(0), nil)
		d.privateEchos["echo-private"] = true

		_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-private",
				Content:   "hi",
				Nickname:  "Bob",
				Email:     "bob@example.com",
				FormToken: freshToken(),
			})
		assertFail(t, err, http.StatusBadRequest, "comment target is not visible")
	})

	t.Run("missing echo", func(t *testing.T) {
		d := newDeps(t)
		d.expectSetting(t, enabledSetting())
		d.repo.EXPECT().
			CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
			Return(int64(0), nil)
		d.repo.EXPECT().
			CountByEmailWithin(mock.Anything, mock.Anything, mock.Anything).
			Return(int64(0), nil)
		d.missingEchos["echo-missing"] = assert.AnError

		_, err := d.service().CreateComment(helpers.CtxAnonymous(), testIP, "ua",
			&commentModel.CreateCommentDto{
				EchoID:    "echo-missing",
				Content:   "hi",
				Nickname:  "Bob",
				Email:     "bob@example.com",
				FormToken: freshToken(),
			})
		assertFail(t, err, http.StatusBadRequest, "comment target is not visible")
	})
}

func TestCreateComment_MemberCanCommentOnPrivateEcho(t *testing.T) {
	helpers.SetJWTSecret(t, testSecret)
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())
	d.privateEchos["echo-private"] = true

	member := helpers.NewUser()
	member.ID = "user-normal"
	member.Username = "Bob"
	d.common.EXPECT().
		CommonGetUserByUserId(mock.Anything, "user-normal").
		Return(member, nil).
		Once()
	d.repo.EXPECT().
		ExistsRecentDuplicate(
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything,
		).
		Return(false, nil).
		Once()
	d.repo.EXPECT().
		CreateComment(mock.Anything, mock.Anything).
		Run(func(_ context.Context, c *commentModel.Comment) { c.ID = "private-cmt" }).
		Return(nil).
		Once()

	res, err := d.service().CreateComment(helpers.CtxAsUser("user-normal"), testIP, "ua",
		&commentModel.CreateCommentDto{
			EchoID:    "echo-private",
			Content:   "hello member",
			FormToken: freshToken(),
		})
	require.NoError(t, err)
	assert.Equal(t, "private-cmt", res.ID)
}

func TestCreateIntegrationComment_RejectsPrivateEcho(t *testing.T) {
	d := newDeps(t)
	d.expectSetting(t, enabledSetting())
	d.repo.EXPECT().
		CountByIPWithin(mock.Anything, mock.Anything, mock.Anything).
		Return(int64(0), nil)
	d.repo.EXPECT().
		CountByUserWithin(mock.Anything, "user-9", mock.Anything).
		Return(int64(0), nil).
		Once()
	d.privateEchos["echo-private"] = true

	_, err := d.service().CreateIntegrationComment(integrationCtx(), testIP, "ua",
		&commentModel.CreateIntegrationCommentDto{
			EchoID:  "echo-private",
			Content: "hi from bot",
		})
	assertFail(t, err, http.StatusBadRequest, "comment target is not visible")
}
