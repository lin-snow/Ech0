// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

const (
	EIGHT_HOUR_EXPIRY string = "8_hours"
	ONE_MONTH_EXPIRY  string = "1_month"
	NEVER_EXPIRY      string = "never"
)

type SystemSetting struct {
	SiteTitle     string `json:"site_title"`
	ServerLogo    string `json:"server_logo"`
	ServerName    string `json:"server_name"`
	ServerURL     string `json:"server_url"`
	HomeLayout    string `json:"home_layout"`
	AllowRegister bool   `json:"allow_register"`
	DefaultLocale string `json:"default_locale"`
	ICPNumber     string `json:"ICP_number"`
	FooterContent string `json:"footer_content"`
	FooterLink    string `json:"footer_link"`
	MetingAPI     string `json:"meting_api"`
	CustomCSS     string `json:"custom_css"`
	CustomJS      string `json:"custom_js"`
}

type S3Setting struct {
	Enable       bool   `json:"enable"`
	Provider     string `json:"provider"`
	Endpoint     string `json:"endpoint"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	BucketName   string `json:"bucket_name"`
	Region       string `json:"region"`
	UseSSL       bool   `json:"use_ssl"`
	CDNURL       string `json:"cdn_url"`
	PathPrefix   string `json:"path_prefix"`
	PublicRead   bool   `json:"public_read"`
	UsePathStyle bool   `json:"use_path_style"`
}

type OAuth2Setting struct {
	Enable       bool     `json:"enable"`
	Provider     string   `json:"provider"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURI  string   `json:"redirect_uri"`
	Scopes       []string `json:"scopes"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	UserInfoURL  string   `json:"user_info_url"`

	IsOIDC  bool   `json:"is_oidc"`
	Issuer  string `json:"issuer"`
	JWKSURL string `json:"jwks_url"`

	AuthRedirectAllowedReturnURLs []string `json:"auth_redirect_allowed_return_urls"`
	CORSAllowedOrigins            []string `json:"cors_allowed_origins"`
}

type PasskeySetting struct {
	WebAuthnRPID           string   `json:"webauthn_rp_id"`
	WebAuthnAllowedOrigins []string `json:"webauthn_allowed_origins"`
}

type AccessTokenSetting struct {
	ID         string `gorm:"type:char(36);primaryKey" json:"id"`
	UserID     string `gorm:"type:char(36);index" json:"user_id"`
	Token      string `gorm:"type:varchar(255);uniqueIndex" json:"token"`
	Name       string `json:"name"`
	TokenType  string `gorm:"size:32;index" json:"token_type"`
	Scopes     string `gorm:"type:text" json:"scopes"`
	Audience   string `gorm:"size:64;index" json:"audience"`
	JTI        string `gorm:"size:64;uniqueIndex" json:"jti"`
	Expiry     *int64 `json:"expiry"`
	LastUsedAt *int64 `json:"last_used_at,omitempty"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at"`
}

func (a *AccessTokenSetting) BeforeCreate(_ *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuidUtil.NewV7()
	}
	return nil
}

type AgentSetting struct {
	Enable        bool   `json:"enable"`
	Protocol      string `json:"protocol"`
	Model         string `json:"model"`
	ApiKey        string `json:"api_key"`
	Prompt        string `json:"prompt"`
	BaseURL       string `json:"base_url"`
	Multimodal    bool   `json:"multimodal"`
	ContextWindow int    `json:"context_window"`
}

type SnapshotSchedule struct {
	Enable         bool   `json:"enable"`
	CronExpression string `json:"cron_expression"`
}
