// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type SourceType string

const (
	SourceGuest       SourceType = "guest"
	SourceSystem      SourceType = "system"
	SourceIntegration SourceType = "integration"
)

const (
	CommentSystemSettingKey = "comment_system_setting"
)

type Comment struct {
	ID        string     `gorm:"type:char(36);primaryKey" json:"id"`
	EchoID    string     `gorm:"type:char(36);not null;index" json:"echo_id"`
	ParentID  *string    `gorm:"type:char(36);index" json:"parent_id,omitempty"`
	UserID    *string    `gorm:"type:char(36);index" json:"user_id,omitempty"`
	Nickname  string     `gorm:"size:100;not null;index" json:"nickname"`
	Email     string     `gorm:"size:255;not null;index" json:"email"`
	Website   string     `gorm:"size:255" json:"website,omitempty"`
	Content   string     `gorm:"type:text;not null" json:"content"`
	Status    Status     `gorm:"type:varchar(20);not null;index" json:"status"`
	Hot       bool       `gorm:"not null;default:false;index" json:"hot"`
	IPHash    string     `gorm:"size:128;index" json:"-"`
	UserAgent string     `gorm:"size:512" json:"-"`
	Source    SourceType `gorm:"type:varchar(20);not null;index" json:"source"`
	CreatedAt int64      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64      `gorm:"autoUpdateTime" json:"updated_at"`
}

type PublicComment struct {
	ID        string     `json:"id"`
	EchoID    string     `json:"echo_id"`
	ParentID  *string    `json:"parent_id,omitempty"`
	Nickname  string     `json:"nickname"`
	Website   string     `json:"website,omitempty"`
	Content   string     `json:"content"`
	Status    Status     `json:"status"`
	Hot       bool       `json:"hot"`
	Source    SourceType `json:"source"`
	CreatedAt int64      `json:"created_at"`
	UpdatedAt int64      `json:"updated_at"`
}

func ToPublicComment(c Comment) PublicComment {
	return PublicComment{
		ID:        c.ID,
		EchoID:    c.EchoID,
		ParentID:  c.ParentID,
		Nickname:  c.Nickname,
		Website:   c.Website,
		Content:   c.Content,
		Status:    c.Status,
		Hot:       c.Hot,
		Source:    c.Source,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ToPublicComments(in []Comment) []PublicComment {
	out := make([]PublicComment, len(in))
	for i := range in {
		out[i] = ToPublicComment(in[i])
	}
	return out
}

func (c *Comment) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuidUtil.NewV7()
	}
	return nil
}

type CreateCommentDto struct {
	EchoID        string `json:"echo_id" binding:"required"`
	ParentID      string `json:"parent_id"`
	Nickname      string `json:"nickname"`
	Email         string `json:"email"`
	Website       string `json:"website"`
	Content       string `json:"content" binding:"required"`
	HoneypotField string `json:"hp_field"`
	FormToken     string `json:"form_token" binding:"required"`
	CaptchaToken  string `json:"captcha_token"`
}

type CreateCommentResult struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}

type CreateIntegrationCommentDto struct {
	EchoID   string `json:"echo_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Nickname string `json:"nickname"`
	Metadata string `json:"metadata"`
}

type UpdateCommentStatusDto struct {
	Status Status `json:"status" binding:"required"`
}

type UpdateCommentHotDto struct {
	Hot bool `json:"hot"`
}

type BatchCommentActionDto struct {
	Action string   `json:"action" binding:"required"`
	IDs    []string `json:"ids" binding:"required"`
}

type ListCommentQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Status   string
	EchoID   string
	Hot      *bool
}

type PageResult[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

type FormMeta struct {
	FormToken          string `json:"form_token"`
	MinSubmitMs        int64  `json:"min_submit_ms"`
	CaptchaEnabled     bool   `json:"captcha_enabled"`
	CaptchaAPIEndpoint string `json:"captcha_api_endpoint"`
	EnableComment      bool   `json:"enable_comment"`
}

type SystemSetting struct {
	EnableComment   bool               `json:"enable_comment"`
	RequireApproval bool               `json:"require_approval"`
	CaptchaEnabled  bool               `json:"captcha_enabled"`
	EmailNotify     EmailNotifySetting `json:"email_notify"`
}

type EmailNotifySetting struct {
	Enabled         bool   `json:"enabled"`
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int    `json:"smtp_port"`
	SMTPUsername    string `json:"smtp_username"`
	SMTPPassword    string `json:"smtp_password,omitempty"`
	SMTPPasswordSet bool   `json:"smtp_password_set,omitempty"`
	SMTPSender      string `json:"smtp_sender"`
}

type TestEmailRequest struct {
	Setting SystemSetting `json:"setting" binding:"required"`
}
