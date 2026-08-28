// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

const datasetSchemaVersion = 1

type dataset struct {
	SchemaVersion int    `json:"schema_version"`
	GeneratedAt   int64  `json:"generated_at"`
	BaseURL       string `json:"base_url"`

	InitStatus  initStatus     `json:"init_status"`
	Settings    settings       `json:"settings"`
	Hello       hello          `json:"hello"`
	Agent       agent          `json:"agent"`
	Echos       []echo         `json:"echos"`
	Tags        []tag          `json:"tags"`
	Heatmap     []heatmapEntry `json:"heatmap"`
	Comments    []comment      `json:"comments"`
	CommentForm commentForm    `json:"comment_form"`
	Connects    []connectItem  `json:"connects"`
	Connect     connectInfo    `json:"connect"`
}

type initStatus struct {
	Initialized bool `json:"initialized"`
	OwnerExists bool `json:"owner_exists"`
}

type settings struct {
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

type hello struct {
	Hello     string `json:"hello"`
	Copyright string `json:"copyright"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	License   string `json:"license"`
	Author    string `json:"author"`
	RepoURL   string `json:"repo_url"`
}

type agent struct {
	Enable        bool   `json:"enable"`
	Protocol      string `json:"protocol"`
	Model         string `json:"model"`
	APIKey        string `json:"api_key"`
	Prompt        string `json:"prompt"`
	BaseURL       string `json:"base_url"`
	Multimodal    bool   `json:"multimodal"`
	ContextWindow int    `json:"context_window"`
}

type echo struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Username  string     `json:"username"`
	Layout    string     `json:"layout"`
	Private   bool       `json:"private"`
	UserID    string     `json:"user_id"`
	FavCount  int        `json:"fav_count"`
	CreatedAt int64      `json:"created_at"`
	EchoFiles []echoFile `json:"echo_files"`
	Tags      []tag      `json:"tags"`
	Extension *extension `json:"extension"`

	tagNames []string
}

type echoFile struct {
	ID        string `json:"id"`
	EchoID    string `json:"echo_id"`
	FileID    string `json:"file_id"`
	SortOrder int    `json:"sort_order"`
	File      file   `json:"file"`
}

type file struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	StorageType string `json:"storage_type"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Category    string `json:"category"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	UserID      string `json:"user_id"`
	CreatedAt   int64  `json:"created_at"`
}

type tag struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsageCount int    `json:"usage_count"`
	CreatedAt  int64  `json:"created_at"`
}

type extension struct {
	ID        string         `json:"id"`
	EchoID    string         `json:"echo_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	CreatedAt int64          `json:"created_at"`
	UpdatedAt int64          `json:"updated_at"`
}

type heatmapEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type comment struct {
	ID        string  `json:"id"`
	EchoID    string  `json:"echo_id"`
	ParentID  *string `json:"parent_id"`
	UserID    *string `json:"user_id"`
	Nickname  string  `json:"nickname"`
	Email     string  `json:"email"`
	Website   string  `json:"website"`
	Content   string  `json:"content"`
	Status    string  `json:"status"`
	Hot       bool    `json:"hot"`
	Source    string  `json:"source"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

type commentForm struct {
	FormToken          string `json:"form_token"`
	MinSubmitMs        int    `json:"min_submit_ms"`
	CaptchaEnabled     bool   `json:"captcha_enabled"`
	CaptchaAPIEndpoint string `json:"captcha_api_endpoint"`
	EnableComment      bool   `json:"enable_comment"`
}

type connectItem struct {
	ID         string `json:"id"`
	ConnectURL string `json:"connect_url"`
}

type connectInfo struct {
	ServerName  string `json:"server_name"`
	ServerURL   string `json:"server_url"`
	Logo        string `json:"logo"`
	TotalEchos  int    `json:"total_echos"`
	TodayEchos  int    `json:"today_echos"`
	SysUsername string `json:"sys_username"`
	Version     string `json:"version"`
}

type resultEnvelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data connectInfo `json:"data"`
}
