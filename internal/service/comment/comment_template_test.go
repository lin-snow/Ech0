// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	model "github.com/lin-snow/ech0/internal/model/comment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldNotify(t *testing.T) {
	enabled := func() model.SystemSetting {
		return model.SystemSetting{EmailNotify: model.EmailNotifySetting{Enabled: true}}
	}

	tests := []struct {
		name            string
		enabled         bool
		requireApproval bool
		kind            string
		status          model.Status
		want            bool
	}{
		{"disabled mail blocks everything", false, true, "created", model.StatusPending, false},
		{"created always notifies", true, false, "created", model.StatusPending, true},
		{"hot always notifies", true, false, "hot", model.StatusApproved, true},
		{"status rejected notifies", true, false, "status", model.StatusRejected, true},
		{"status approved with approval flow notifies", true, true, "status", model.StatusApproved, true},
		{"status approved without approval flow stays quiet", true, false, "status", model.StatusApproved, false},
		{"status pending is not notified", true, true, "status", model.StatusPending, false},
		{"unknown kind is not notified", true, true, "mystery", model.StatusApproved, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := enabled()
			s.EmailNotify.Enabled = tc.enabled
			s.RequireApproval = tc.requireApproval
			assert.Equal(t, tc.want, shouldNotify(s, tc.kind, tc.status))
		})
	}
}

func TestBuildNotifyContent(t *testing.T) {
	const server = "https://ech0.example.com"

	tests := []struct {
		name        string
		kind        string
		comment     model.Comment
		serverURL   string
		wantSubject string
		wantInText  string
		wantLink    bool
	}{
		{
			name:        "created pending with full fields and echo link",
			kind:        "created",
			comment:     model.Comment{EchoID: "e1", Nickname: "Bob", Email: "bob@x.com", Content: "hello", Status: model.StatusPending, CreatedAt: 1700000000},
			serverURL:   server,
			wantSubject: "[Ech0] 新评论待审核",
			wantInText:  "hello",
			wantLink:    true,
		},
		{
			name:        "reply approved",
			kind:        "reply",
			comment:     model.Comment{EchoID: "e1", Nickname: "Bob", Content: "re", Status: model.StatusApproved},
			serverURL:   server,
			wantSubject: "[Ech0] 有人回复了你的评论",
			wantInText:  "re",
			wantLink:    true,
		},
		{
			name:        "status approved",
			kind:        "status",
			comment:     model.Comment{EchoID: "e1", Content: "ok", Status: model.StatusApproved},
			serverURL:   server,
			wantSubject: "[Ech0] 您的评论已通过审核",
			wantInText:  "ok",
			wantLink:    true,
		},
		{
			name:        "status rejected",
			kind:        "status",
			comment:     model.Comment{EchoID: "e1", Content: "no", Status: model.StatusRejected},
			serverURL:   server,
			wantSubject: "[Ech0] 您的评论未通过审核",
			wantInText:  "no",
			wantLink:    true,
		},
		{
			name:        "status pending falls back to generic subject",
			kind:        "status",
			comment:     model.Comment{EchoID: "e1", Content: "pend", Status: model.StatusPending},
			serverURL:   server,
			wantSubject: "[Ech0] 评论状态已更新",
			wantInText:  "pend",
			wantLink:    true,
		},
		{
			name:        "hot",
			kind:        "hot",
			comment:     model.Comment{EchoID: "e1", Content: "wow", Status: model.StatusApproved},
			serverURL:   server,
			wantSubject: "[Ech0] 您的评论被标为精选",
			wantInText:  "wow",
			wantLink:    true,
		},
		{
			name:        "test kind with empty content and missing author falls back, no link",
			kind:        "test",
			comment:     model.Comment{EchoID: "", Nickname: "", Email: "", Content: "", Status: model.StatusPending},
			serverURL:   "",
			wantSubject: "[Ech0] 邮件通知测试",
			wantInText:  "匿名用户",
			wantLink:    false,
		},
		{
			name:        "unknown status uses default style",
			kind:        "created",
			comment:     model.Comment{EchoID: "e1", Content: "x", Status: model.Status("weird")},
			serverURL:   server,
			wantSubject: "[Ech0] 新评论待审核",
			wantInText:  "x",
			wantLink:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildNotifyContent(tc.kind, tc.comment, tc.serverURL)
			assert.Equal(t, tc.wantSubject, got.Subject)
			assert.NotEmpty(t, got.TextBody)
			assert.NotEmpty(t, got.HTMLBody)
			assert.Contains(t, got.TextBody, tc.wantInText)
			if tc.wantLink {
				assert.Contains(t, got.TextBody, "查看 Echo")
				assert.Contains(t, got.HTMLBody, "/echo/")
			} else {
				assert.NotContains(t, got.TextBody, "查看 Echo")
			}
		})
	}

	t.Run("test kind injects placeholder body when content empty", func(t *testing.T) {
		got := buildNotifyContent("test", model.Comment{Status: model.StatusPending}, "")
		assert.Contains(t, got.TextBody, "测试邮件")
	})
}

func TestBuildEchoLink(t *testing.T) {
	tests := []struct {
		name      string
		serverURL string
		echoID    string
		want      string
	}{
		{"empty server url yields empty", "", "e1", ""},
		{"empty echo id yields empty", "https://x.com", "", ""},
		{"trims trailing slash and escapes id", "https://x.com/", "a b", "https://x.com/echo/a%20b"},
		{"plain", "https://x.com", "e1", "https://x.com/echo/e1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, buildEchoLink(tc.serverURL, tc.echoID))
		})
	}
}

func TestFallbackText(t *testing.T) {
	assert.Equal(t, "fallback", fallbackText("   ", "fallback"))
	assert.Equal(t, "value", fallbackText("  value  ", "fallback"))
}

func TestNotifyTime(t *testing.T) {
	assert.NotEmpty(t, notifyTime(0))
	got := notifyTime(1700000000)
	assert.Len(t, got, len("2006-01-02 15:04:05"))
}

func TestNotifyStatusStyle(t *testing.T) {
	tests := []struct {
		name      string
		kind      string
		status    model.Status
		wantLabel string
	}{
		{"hot", "hot", model.StatusApproved, "HOT"},
		{"reply", "reply", model.StatusApproved, "回复"},
		{"status approved", "status", model.StatusApproved, "已通过"},
		{"status rejected", "status", model.StatusRejected, "已拒绝"},
		{"default pending", "created", model.StatusPending, "待审核"},
		{"default approved", "created", model.StatusApproved, "已通过"},
		{"default rejected", "created", model.StatusRejected, "已拒绝"},
		{"default unknown", "created", model.Status("weird"), "通知"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			label, color, bg := notifyStatusStyle(tc.kind, tc.status)
			assert.Equal(t, tc.wantLabel, label)
			assert.NotEmpty(t, color)
			assert.NotEmpty(t, bg)
		})
	}
}

func TestApplySettingDefaults(t *testing.T) {
	t.Run("nil is a no-op", func(t *testing.T) {
		applySettingDefaults(nil)
	})
	t.Run("zero port defaults to 587", func(t *testing.T) {
		s := &model.SystemSetting{}
		applySettingDefaults(s)
		assert.Equal(t, 587, s.EmailNotify.SMTPPort)
	})
	t.Run("explicit port is preserved", func(t *testing.T) {
		s := &model.SystemSetting{EmailNotify: model.EmailNotifySetting{SMTPPort: 25}}
		applySettingDefaults(s)
		assert.Equal(t, 25, s.EmailNotify.SMTPPort)
	})
}

type stubMailer struct {
	called bool
	err    error
}

func (m *stubMailer) Send(_ context.Context, _ MailerConfig, _ MailMessage) error {
	m.called = true
	return m.err
}

func TestSendOwnerMail(t *testing.T) {
	validCfg := model.EmailNotifySetting{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "secret",
		SMTPSender:   "noreply@example.com",
	}
	validMsg := MailMessage{To: "owner@example.com", Subject: "s", TextBody: "t", HTMLBody: "<p>t</p>"}

	t.Run("nil mailer is unavailable", func(t *testing.T) {
		s := &CommentService{}
		err := s.sendOwnerMail(context.Background(), validCfg, validMsg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mailer unavailable")
	})

	t.Run("invalid config fails before calling mailer", func(t *testing.T) {
		stub := &stubMailer{}
		s := &CommentService{mailer: stub}
		badCfg := validCfg
		badCfg.SMTPHost = ""
		err := s.sendOwnerMail(context.Background(), badCfg, validMsg)
		require.Error(t, err)
		assert.False(t, stub.called, "mailer must not be invoked when config is invalid")
	})

	t.Run("valid config delegates to mailer and trims fields", func(t *testing.T) {
		stub := &stubMailer{}
		s := &CommentService{mailer: stub}
		cfg := validCfg
		cfg.SMTPHost = "  smtp.example.com  "
		err := s.sendOwnerMail(context.Background(), cfg, validMsg)
		require.NoError(t, err)
		assert.True(t, stub.called)
	})

	t.Run("mailer error is propagated", func(t *testing.T) {
		stub := &stubMailer{err: errors.New("smtp down")}
		s := &CommentService{mailer: stub}
		err := s.sendOwnerMail(context.Background(), validCfg, validMsg)
		require.Error(t, err)
		assert.True(t, stub.called)
		assert.Contains(t, strings.ToLower(err.Error()), "smtp")
	})
}
