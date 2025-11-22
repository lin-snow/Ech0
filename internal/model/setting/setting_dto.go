package model

// SystemSettingDto 定义系统设置数据传输对象
type SystemSettingDto struct {
	SiteTitle     string `json:"site_title"`     // 站点标题
	ServerName    string `json:"server_name"`    // 服务器名称
	ServerURL     string `json:"server_url"`     // 服务器地址
	AllowRegister bool   `json:"allow_register"` // 是否允许注册
	ICPNumber     string `json:"ICP_number"`     // 备案号
	MetingAPI     string `json:"meting_api"`     // Meting API 地址
	CommentAPI    string `json:"comment_api"`    // 评论 API 地址
	CustomCSS     string `json:"custom_css"`     // 自定义 CSS
	CustomJS      string `json:"custom_js"`      // 自定义 JS
	Logo          string `json:"logo"`           // 站点Logo路径
}

type CommentSettingDto struct {
	EnableComment bool   `json:"enable_comment"` // 是否启用评论
	Provider      string `json:"provider"`       // 评论提供者
	CommentAPI    string `json:"comment_api"`    // 评论 API 地址
}

type S3SettingDto struct {
	Enable     bool   `json:"enable"`      // 是否启用 S3 存储
	Provider   string `json:"provider"`    // S3 服务提供商，例如 "aws", "aliyun", "minio", "other"
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
}

type OAuth2Status struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
}

type WebhookDto struct {
	Name     string `json:"name"`                                 // Webhook 名称
	URL      string `json:"url"`                                  // Webhook URL
	Secret   string `json:"secret,omitempty"`                     // 签名密钥，用于请求验证（HMAC等）
	IsActive bool   `json:"is_active"        gorm:"default:true"` // 启用/禁用状态
}

type AccessTokenSettingDto struct {
	Name   string `json:"name"`   // 访问令牌名称
	Expiry string `json:"expiry"` // 访问令牌过期时间，Unix 时间戳格式
}

type FediverseSettingDto struct {
	Enable    bool   `json:"enable"`     // 是否启用联邦网络功能
	ServerURL string `json:"server_url"` // 服务器 URL
}

type BackupScheduleDto struct {
	Enable         bool   `json:"enable"`          // 是否启用备份计划
	CronExpression string `json:"cron_expression"` // 备份计划的 Cron 表达式
}
