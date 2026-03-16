package model

import (
	"time"

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
	SourceGuest  SourceType = "guest"
	SourceSystem SourceType = "system"
)

const (
	CommentSystemSettingKey = "comment_system_setting"
)

type Comment struct {
	ID        string     `gorm:"type:char(36);primaryKey" json:"id"`
	EchoID    string     `gorm:"type:char(36);not null;index" json:"echo_id"`
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
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (c *Comment) BeforeCreate(_ *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuidUtil.MustNewV7()
	}
	return nil
}

type CreateCommentDto struct {
	EchoID        string `json:"echo_id" binding:"required"`
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
	EnableComment           bool   `json:"enable_comment"`
	RequireApproval         bool   `json:"require_approval"`
	CaptchaEnabled          bool   `json:"captcha_enabled"`
	CaptchaVerify           string `json:"captcha_verify_url"`
	CaptchaSecret           string `json:"captcha_secret"`
	EnableEmailNotification bool   `json:"enable_email_notification"`
	SMTPHost                string `json:"smtp_host"`
	SMTPPort                int    `json:"smtp_port"`
	SMTPUser                string `json:"smtp_user"`
	SMTPPassword            string `json:"smtp_password"`
}
