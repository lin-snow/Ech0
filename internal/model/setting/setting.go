package model

import "time"

const (
	EIGHT_HOUR_EXPIRY string = "8_hours"
	ONE_MONTH_EXPIRY  string = "1_month"
	NEVER_EXPIRY      string = "never"
)

// SystemSetting 定义系统设置实体
type SystemSetting struct {
	SiteTitle     string `json:"site_title"`     // 站点标题
	ServerLogo    string `json:"server_logo"`    // 服务器Logo
	ServerName    string `json:"server_name"`    // 服务器名称
	ServerURL     string `json:"server_url"`     // 服务器地址
	AllowRegister bool   `json:"allow_register"` // 是否允许注册'
	ICPNumber     string `json:"ICP_number"`     // 备案号
	MetingAPI     string `json:"meting_api"`     // Meting API 地址
	CustomCSS     string `json:"custom_css"`     // 自定义 CSS
	CustomJS      string `json:"custom_js"`      // 自定义 JS
}

// CommentSetting 定义评论设置实体
type CommentSetting struct {
	EnableComment bool   `json:"enable_comment"` // 是否启用评论
	Provider      string `json:"provider"`       // 评论提供者
	CommentAPI    string `json:"comment_api"`    // 评论 API 地址
}

// S3Setting 定义 S3 存储设置实体
type S3Setting struct {
	Enable     bool   `json:"enable"`      // 是否启用 S3 存储
	Provider   string `json:"provider"`    // S3 服务提供商，例如 "aws", "r2", "minio", "other"
	Endpoint   string `json:"endpoint"`    // S3 端点
	AccessKey  string `json:"access_key"`  // 访问密钥 ID
	SecretKey  string `json:"secret_key"`  // 秘密访问密钥
	BucketName string `json:"bucket_name"` // 存储桶名称
	Region     string `json:"region"`      // 区域
	UseSSL     bool   `json:"use_ssl"`     // 是否使用 SSL
	CDNURL     string `json:"cdn_url"`     // CDN 加速域名（可选，没有就走 Endpoint）
	PathPrefix string `json:"path_prefix"` // 存储路径前缀，例如 "uploads/"，方便隔离目录
	PublicRead bool   `json:"public_read"` // 上传时是否默认设置对象为 public-read
}

// OAuth2Setting 定义 OAuth2 配置结构体
type OAuth2Setting struct {
	Enable       bool     `json:"enable"`        // 是否启用 OAuth2 登录
	Provider     string   `json:"provider"`      // OAuth2 提供商
	ClientID     string   `json:"client_id"`     // OAuth2 Client ID
	ClientSecret string   `json:"client_secret"` // OAuth2 Client Secret
	RedirectURI  string   `json:"redirect_uri"`  // OAuth2 重定向 URI
	Scopes       []string `json:"scopes"`        // OAuth2 请求的权限范围
	AuthURL      string   `json:"auth_url"`      // OAuth2 授权 URL
	TokenURL     string   `json:"token_url"`     // OAuth2 令牌 URL
	UserInfoURL  string   `json:"user_info_url"` // OAuth2 用户信息 URL

	// OIDC 扩展
	IsOIDC  bool   `json:"is_oidc"`  // 是否启用 OIDC
	Issuer  string `json:"issuer"`   // OIDC 颁发者
	JWKSURL string `json:"jwks_url"` // OIDC JWKS URL
}

// AccessTokenSetting 定义访问令牌设置实体
type AccessTokenSetting struct {
	ID        int        `json:"id"`         // 访问令牌 ID
	UserID    uint       `json:"user_id"`    // 创建该访问令牌的用户 ID
	Token     string     `json:"token"`      // 访问令牌
	Name      string     `json:"name"`       // 访问令牌名称
	Expiry    *time.Time `json:"expiry"`     // 指针类型，NULL 表示永不过期
	CreatedAt time.Time  `json:"created_at"` // 访问令牌创建时间，RFC3339 时间字符串
}

// FediverseSetting 定义联邦网络设置实体
type FediverseSetting struct {
	Enable    bool   `json:"enable"`     // 是否启用联邦网络功能
	ServerURL string `json:"server_url"` // 服务器 URL
}

// AgentSetting 定义 LLM Agent 设置实体
type AgentSetting struct {
	Enable   bool   `json:"enable"`   // 是否启用 Agent 功能
	Provider string `json:"provider"` // LLM 提供商 （OpenAI、DeepSeek、Anthropic、Gemini、阿里百炼、Ollama等）
	Model    string `json:"model"`    // LLM 模型名称
	ApiKey   string `json:"api_key"`  // LLM API Key
	Prompt   string `json:"prompt"`   // Agent 额外使用的提示词
	BaseURL  string `json:"base_url"` // 自定义 API URL（可选）
}

type BackupSchedule struct {
	Enable         bool   `json:"enable"`          // 是否启用备份计划
	CronExpression string `json:"cron_expression"` // 备份计划的 Cron 表达式
}
