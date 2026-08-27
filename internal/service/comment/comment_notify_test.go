// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"testing"

	model "github.com/lin-snow/ech0/internal/model/comment"
)

func TestUseCommentRecipient(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want bool
	}{
		{
			name: "status uses commenter email",
			kind: "status",
			want: true,
		},
		{
			name: "hot uses commenter email",
			kind: "hot",
			want: true,
		},
		{
			name: "created keeps owner email",
			kind: "created",
			want: false,
		},
		{
			name: "other kinds keep owner email",
			kind: "test",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := useCommentRecipient(tc.kind)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestShouldNotifyOwnerOnCreate(t *testing.T) {
	tests := []struct {
		name   string
		source model.SourceType
		want   bool
	}{
		{
			name:   "guest comment notifies owner",
			source: model.SourceGuest,
			want:   true,
		},
		{
			name:   "integration comment notifies owner",
			source: model.SourceIntegration,
			want:   true,
		},
		{
			name:   "owner's own comment does not notify owner",
			source: model.SourceSystem,
			want:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldNotifyOwnerOnCreate(tc.source); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestParseValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{
			name:  "valid email",
			input: "author@example.com",
			want:  "author@example.com",
			ok:    true,
		},
		{
			name:  "trimmed valid email",
			input: "  author@example.com  ",
			want:  "author@example.com",
			ok:    true,
		},
		{
			name:  "empty email",
			input: "   ",
			want:  "",
			ok:    false,
		},
		{
			name:  "invalid format",
			input: "not-an-email",
			want:  "",
			ok:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseValidEmail(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestValidateEmailNotifySetting(t *testing.T) {
	base := model.EmailNotifySetting{
		Enabled:      true,
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "secret",
	}

	tests := []struct {
		name      string
		cfg       model.EmailNotifySetting
		ownerMail string
		wantErr   bool
	}{
		{
			name:      "standard email username without sender",
			cfg:       base,
			ownerMail: "owner@example.com",
			wantErr:   false,
		},
		{
			name: "non-email username with valid sender (resend style)",
			cfg: model.EmailNotifySetting{
				Enabled:      true,
				SMTPHost:     "smtp.resend.com",
				SMTPPort:     587,
				SMTPUsername: "resend",
				SMTPPassword: "re_xxx",
				SMTPSender:   "hi@example.com",
			},
			ownerMail: "owner@example.com",
			wantErr:   false,
		},
		{
			name: "non-email username without sender fails",
			cfg: model.EmailNotifySetting{
				Enabled:      true,
				SMTPHost:     "smtp.resend.com",
				SMTPPort:     587,
				SMTPUsername: "resend",
				SMTPPassword: "re_xxx",
			},
			ownerMail: "owner@example.com",
			wantErr:   true,
		},
		{
			name: "empty host fails",
			cfg: model.EmailNotifySetting{
				SMTPPort:     587,
				SMTPUsername: "user@example.com",
				SMTPPassword: "secret",
			},
			ownerMail: "owner@example.com",
			wantErr:   true,
		},
		{
			name:    "empty owner email fails",
			cfg:     base,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEmailNotifySetting(tc.cfg, tc.ownerMail)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
