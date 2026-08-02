// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

// dataset.json 的形状是**契约**：字段名逐字等于活实例对应 HTTP 端点的 JSON
// 契约，好让前端静态 adapter 用同一套解析代码消费。因此这里不复用领域模型
// （领域模型带 omitempty / gorm tag，序列化出来的字段集合会随数据漂移），
// 而是原地声明一套「只为序列化而生」的扁平结构：所有字段恒定出现，
// 时间一律 Unix 秒 int64。

// datasetSchemaVersion 是烘焙产物的版本；前端遇到更高版本应拒绝。
const datasetSchemaVersion = 1

// dataset 是 <dist>/dataset.json 的根对象。
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

// initStatus 对应 GET /init/status。静态站永远是「已初始化」。
type initStatus struct {
	Initialized bool `json:"initialized"`
	OwnerExists bool `json:"owner_exists"`
}

// settings 对应 GET /settings，字段逐字等于 settingModel.SystemSetting 的 json tag。
type settings struct {
	SiteTitle     string `json:"site_title"`
	ServerLogo    string `json:"server_logo"`
	ServerName    string `json:"server_name"`
	ServerURL     string `json:"server_url"`
	AllowRegister bool   `json:"allow_register"`
	DefaultLocale string `json:"default_locale"`
	ICPNumber     string `json:"ICP_number"`
	FooterContent string `json:"footer_content"`
	FooterLink    string `json:"footer_link"`
	MetingAPI     string `json:"meting_api"`
	CustomCSS     string `json:"custom_css"`
	CustomJS      string `json:"custom_js"`
}

// hello 对应 GET /hello（handler 侧把 version.Info 扁平化到顶层，这里照做）。
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

// agent 对应 GET /agent/info。静态站没有 LLM 后端，恒关且不携带任何凭据。
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

// echo 是 echoModel.Echo 的 JSON 形状。dataset 内已按 created_at 降序排好，
// 前端查询引擎默认序即数组序，无需再排。
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

	// tagNames 是烘焙期的临时载体：标签的 usage_count 要等全量 Echo 都过一遍
	// 才算得出来，先记名字，最后统一回填 Tags。未导出，不参与序列化。
	tagNames []string
}

// echoFile 对应 fileModel.EchoFile。sort_order 即数组下标（胶囊用顺序表达它）。
type echoFile struct {
	ID        string `json:"id"`
	EchoID    string `json:"echo_id"`
	FileID    string `json:"file_id"`
	SortOrder int    `json:"sort_order"`
	File      file   `json:"file"`
}

// file 对应 fileModel.File 的公开投影。
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

// tag 对应 echoModel.Tag。usage_count 是全站引用次数（不是本 Echo 内的）。
type tag struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	UsageCount int    `json:"usage_count"`
	CreatedAt  int64  `json:"created_at"`
}

// extension 对应 echoModel.EchoExtension。
type extension struct {
	ID        string         `json:"id"`
	EchoID    string         `json:"echo_id"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload"`
	CreatedAt int64          `json:"created_at"`
	UpdatedAt int64          `json:"updated_at"`
}

// heatmapEntry 对应 GET /heatmap 的元素；date 为 UTC 的 YYYY-MM-DD。
type heatmapEntry struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// comment 对应 commentModel.PublicComment。email 恒为空串：胶囊不携带它，
// 保留字段只为形状对齐（前端按同一套结构解析）。
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

// commentForm 对应 GET /comments/form。静态站是冻结展示，表单整体关闭。
type commentForm struct {
	FormToken          string `json:"form_token"`
	MinSubmitMs        int    `json:"min_submit_ms"`
	CaptchaEnabled     bool   `json:"captcha_enabled"`
	CaptchaAPIEndpoint string `json:"captcha_api_endpoint"`
	EnableComment      bool   `json:"enable_comment"`
}

// connectItem 对应 GET /connect/list 的元素。
type connectItem struct {
	ID         string `json:"id"`
	ConnectURL string `json:"connect_url"`
}

// connectInfo 对应 GET /connect 的载荷，同时是 <dist>/api/connect 文件里的 data。
// 统计值是构建时的冻结快照（spec §10）。
type connectInfo struct {
	ServerName  string `json:"server_name"`
	ServerURL   string `json:"server_url"`
	Logo        string `json:"logo"`
	TotalEchos  int    `json:"total_echos"`
	TodayEchos  int    `json:"today_echos"`
	SysUsername string `json:"sys_username"`
	Version     string `json:"version"`
}

// resultEnvelope 是活实例 HTTP 响应的信封形状（commonModel.Result）。
// api/connect 必须与之同形，远端实例的既有探测路径才能零改动消费。
type resultEnvelope struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data connectInfo `json:"data"`
}
