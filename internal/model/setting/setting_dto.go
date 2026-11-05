// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type SystemSettingDto struct {
	SiteTitle             string `json:"site_title"`
	ServerLogo            string `json:"server_logo"`
	ServerLogoFileID      string `json:"server_logo_file_id"`
	ServerName            string `json:"server_name"`
	ServerURL             string `json:"server_url"`
	AllowRegister         bool   `json:"allow_register"`
	DefaultLocale         string `json:"default_locale"`
	ICPNumber             string `json:"ICP_number"`
	FooterContent         string `json:"footer_content"`
	FooterLink            string `json:"footer_link"`
	MetingAPI             string `json:"meting_api"`
	CustomCSS             string `json:"custom_css"`
	CustomJS              string `json:"custom_js"`
	DisableAgentDiscovery bool   `json:"disable_agent_discovery"`
}

type S3SettingDto struct {
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

type OAuth2SettingDto struct {
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

type OAuth2Status struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"`
	OAuthReady bool   `json:"oauth_ready"`
}

type PasskeySettingDto struct {
	WebAuthnRPID           string   `json:"webauthn_rp_id"`
	WebAuthnAllowedOrigins []string `json:"webauthn_allowed_origins"`
}

type PasskeyStatus struct {
	PasskeyReady bool `json:"passkey_ready"`
}

type WebhookDto struct {
	Name     string `json:"name" required:"true" doc:"Display name"`
	URL      string `json:"url" required:"true" doc:"HTTP(S) endpoint that receives events"`
	Secret   string `json:"secret,omitempty" doc:"Optional signing secret; the server signs payloads with X-Ech0-Signature"`
	IsActive bool   `json:"is_active" gorm:"default:true" doc:"Enabled; defaults to true on create, while omitting it on update disables the webhook"`
}

type AccessTokenSettingDto struct {
	Name     string   `json:"name"`
	Expiry   string   `json:"expiry"`
	Scopes   []string `json:"scopes"`
	Audience string   `json:"audience"`
}

type SnapshotScheduleDto struct {
	Enable         bool   `json:"enable"`
	CronExpression string `json:"cron_expression"`
}

type AgentSettingDto struct {
	Enable        bool     `json:"enable"`
	Protocol      string   `json:"protocol"`
	Model         string   `json:"model"`
	ApiKey        string   `json:"api_key"`
	Prompt        string   `json:"prompt"`
	BaseURL       string   `json:"base_url"`
	Multimodal    bool     `json:"multimodal"`
	ContextWindow int      `json:"context_window"`
	Temperature   *float64 `json:"temperature,omitempty" nullable:"true" minimum:"0" maximum:"2" doc:"采样温度；留空则不发送，使用模型默认值。是否支持以模型官方文档为准"`
}
