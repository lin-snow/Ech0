// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package capsule 实现 Capsule 交换格式的核心数据模型与编解码。
//
// 规范见 docs/dev/capsule/spec.md。本包只负责「格式」——磁盘布局、
// YAML/frontmatter 结构、时间表示——不触碰数据库，也不做任何面向消费者的
// 转换（「导出即转储，构建即转换」，spec §11）。
//
// 结构体字段与取值一律对齐数据库列：Site 的 yaml 键名逐字等于
// settingModel.SystemSetting 的 json tag（由 schema_test.go 守卫），
// FileRef 的字段逐字等于 fileModel.File 的列。
package capsule

// SchemaVersion 是本实现产出的胶囊版本；消费者遇到更高版本必须拒绝。
const SchemaVersion = 1

// 顶层布局常量（spec §2）。
const (
	ManifestPath = "ech0.yaml"
	CommentsPath = "comments.yaml"
	EchoesDir    = "echoes"
	FilesDir     = "files"
)

// Manifest 是清单文件 ech0.yaml 的结构（spec §3）。
type Manifest struct {
	SchemaVersion int       `yaml:"schema_version"`
	Generator     string    `yaml:"generator,omitempty"`
	ExportedAt    string    `yaml:"exported_at,omitempty"`
	Site          Site      `yaml:"site"`
	Owner         Owner     `yaml:"owner"`
	Connects      []Connect `yaml:"connects,omitempty"`
}

// Site 是站点设置的公开子集。键名逐字对齐 SystemSetting 的 json tag——
// 唯一被剔除的是行为开关 allow_register（spec §3：渲染所需皆入，运维行为皆弃）。
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

// Owner 是归属兜底：Echo 未标 username 时的默认作者。
type Owner struct {
	Username string `yaml:"username"`
}

// Connect 是互联实例快照的元素。
type Connect struct {
	URL string `yaml:"url"`
}

// EchoDoc 是单个 Echo 内容文件：frontmatter 字段 + 正文（spec §4）。
// Content 不进 frontmatter，它就是 --- 之后的全部字节。
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

// FileRef 是 frontmatter 中的媒体引用。字段名与取值对齐 fileModel.File 的列；
// 位置不入胶囊——由 MediaPath(Key) 纯函数派生（spec §4.2）。
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

// Managed 报告该引用是否为托管文件（字节随胶囊走）。key 与 url 互斥，
// 校验在 check 侧执行，此处只做分流。
func (f FileRef) Managed() bool { return f.Key != "" }

// Extension 对应 echoModel.EchoExtension 的 Type/Payload。
type Extension struct {
	Type    string         `yaml:"type"`
	Payload map[string]any `yaml:"payload"`
}

// CommentsDoc 是 comments.yaml 的结构（spec §5）。
type CommentsDoc struct {
	SchemaVersion int       `yaml:"schema_version"`
	Comments      []Comment `yaml:"comments"`
}

// Comment 对齐 commentModel.PublicComment 投影：email / ip_hash /
// user_agent / user_id 禁止出现。
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

// ForbiddenCommentFields 是 comments.yaml 中出现即为校验错误的键（spec §5）。
var ForbiddenCommentFields = []string{"email", "ip_hash", "user_agent", "user_id"}
