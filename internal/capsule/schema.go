// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

const SchemaVersion = 1

const (
	ManifestPath = "ech0.yaml"
	CommentsPath = "comments.yaml"
	EchoesDir    = "echoes"
	FilesDir     = "files"
)

type Manifest struct {
	SchemaVersion int       `yaml:"schema_version"`
	Generator     string    `yaml:"generator,omitempty"`
	ExportedAt    string    `yaml:"exported_at,omitempty"`
	Site          Site      `yaml:"site"`
	Owner         Owner     `yaml:"owner"`
	Connects      []Connect `yaml:"connects,omitempty"`

	Files []FileRef `yaml:"files,omitempty"`
}

type Site struct {
	SiteTitle     string `yaml:"site_title,omitempty"`
	ServerLogo    string `yaml:"server_logo,omitempty"`
	ServerName    string `yaml:"server_name,omitempty"`
	ServerURL     string `yaml:"server_url,omitempty"`
	DefaultLocale string `yaml:"default_locale,omitempty"`
	ICPNumber     string `yaml:"ICP_number,omitempty"`
	FooterContent string `yaml:"footer_content,omitempty"`
	FooterLink    string `yaml:"footer_link,omitempty"`
	MetingAPI     string `yaml:"meting_api,omitempty"`
	CustomCSS     string `yaml:"custom_css,omitempty"`
	CustomJS      string `yaml:"custom_js,omitempty"`
}

type Owner struct {
	Username string `yaml:"username"`
}

type Connect struct {
	URL string `yaml:"url"`
}

type EchoDoc struct {
	ID        string     `yaml:"id"`
	CreatedAt string     `yaml:"created_at"`
	Username  string     `yaml:"username,omitempty"`
	Tags      []string   `yaml:"tags,omitempty"`
	Layout    string     `yaml:"layout,omitempty"`
	Private   bool       `yaml:"private,omitempty"`
	FavCount  int        `yaml:"fav_count,omitempty"`
	Files     []FileRef  `yaml:"files,omitempty"`
	Extension *Extension `yaml:"extension,omitempty"`

	Content string `yaml:"-"`
}

type FileRef struct {
	ID          string `yaml:"id,omitempty"`
	Key         string `yaml:"key,omitempty"`
	URL         string `yaml:"url,omitempty"`
	Category    string `yaml:"category,omitempty"`
	Name        string `yaml:"name,omitempty"`
	ContentType string `yaml:"content_type,omitempty"`
	Size        int64  `yaml:"size,omitempty"`
	Width       int    `yaml:"width,omitempty"`
	Height      int    `yaml:"height,omitempty"`
}

func (f FileRef) Managed() bool { return f.Key != "" }

type Extension struct {
	Type    string         `yaml:"type"`
	Payload map[string]any `yaml:"payload"`
}

type CommentsDoc struct {
	SchemaVersion int       `yaml:"schema_version"`
	Comments      []Comment `yaml:"comments"`
}

type Comment struct {
	ID        string  `yaml:"id"`
	EchoID    string  `yaml:"echo_id"`
	ParentID  *string `yaml:"parent_id,omitempty"`
	Nickname  string  `yaml:"nickname"`
	Website   string  `yaml:"website,omitempty"`
	Content   string  `yaml:"content"`
	Status    string  `yaml:"status,omitempty"`
	Source    string  `yaml:"source,omitempty"`
	CreatedAt string  `yaml:"created_at"`
}

var ForbiddenCommentFields = []string{"email", "ip_hash", "user_agent", "user_id"}
