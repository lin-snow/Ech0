package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var (
	cfg  *AppConfig
	once sync.Once
)

type AppConfig struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Log       LogConfig
	Auth      AuthConfig
	Upload    UploadConfig
	Storage   StorageConfig
	Event     EventConfig
	Migration MigrationConfig
	Setting   SettingConfig
	Comment   CommentConfig
	Security  SecurityConfig
	Web       WebConfig
}

type StorageConfig struct {
	ObjectEnabled bool   // enable object storage alongside local
	DataRoot      string // local root directory, default "data/files"
	Endpoint      string // S3-compatible endpoint
	AccessKey     string
	SecretKey     string
	BucketName    string
	Region        string
	Provider      string // "aws", "r2", "minio", "other"
	UseSSL        bool
	CDNURL        string
	PathPrefix    string
}

type ServerConfig struct {
	Port string // 服务器端口
	Host string // 服务器主机地址
	Mode string // 运行模式，可能的值为 "debug" 或 "release"
}

type DatabaseConfig struct {
	Type    string // 数据库类型
	Path    string // 数据库文件路径
	LogMode string // 数据库日志模式
}

type LogConfig struct {
	Level           string
	Format          string
	Console         bool
	FileEnable      bool
	FilePath        string
	FileMaxSize     int
	FileMaxBackups  int
	FileMaxAge      int
	FileCompress    bool
	BufferSize      int
	RecentSize      int
	DropPolicy      string
	FlushBatch      int
	FlushIntervalMs int
}

type AuthConfig struct {
	Jwt      JWTConfig
	Redirect RedirectConfig
	WebAuthn WebAuthnConfig
}

type JWTConfig struct {
	Expires        int    // Access Token 过期时间，单位为秒
	RefreshExpires int    // Refresh Token 过期时间，单位为秒
	Issuer         string // JWT的发行者
	Audience       string // JWT的受众
}

type RedirectConfig struct {
	AllowedReturnURLs []string
}

type WebAuthnConfig struct {
	RPID    string
	Origins []string
}

type UploadConfig struct {
	ImageMaxSize int      // 图片文件的最大上传大小，单位为字节
	AudioMaxSize int      // 音频文件的最大上传大小，单位为字节
	AllowedTypes []string // 允许上传的文件类型
	ImagePath    string   // 图片文件存储路径
	AudioPath    string   // 音频文件存储路径
}

type SettingConfig struct {
	SiteTitle     string // 网站标题
	ServerLogo    string // 服务器Logo
	Servername    string // 服务器名称
	Serverurl     string // 服务器 URL
	AllowRegister bool   // 是否允许注册
	Icpnumber     string // ICP 备案号
	FooterContent string // 自定义页脚内容
	FooterLink    string // 自定义页脚链接
	MetingAPI     string // Meting API 地址
	CustomCSS     string // 自定义 CSS 样式
	CustomJS      string // 自定义 JS 脚本
}

type CommentConfig struct {
	EnableComment         bool // 是否启用评论
	CaptchaSiteKey        string
	CaptchaSecret         string
	CaptchaDifficulty     int
	CaptchaChallengeCount int
	CaptchaSaltSize       int
	CaptchaChallengeTTL   int
	CaptchaRedeemTTL      int
	CaptchaGCInterval     int
	CaptchaEnableCORS     bool
	CaptchaIPHeader       string
	CaptchaMaxBodyBytes   int
	CaptchaRateLimitMax   int
	CaptchaRateLimitWin   int
	CaptchaRateLimitScope string
	CaptchaLimitOnRedeem  bool
	CaptchaLimitOnVerify  bool
}

type SecurityConfig struct {
	JWTSecret []byte
}

type WebConfig struct {
	CORS CORSConfig
}

type CORSConfig struct {
	AllowedOrigins []string
}

type EventConfig struct {
	DefaultBuffer      int
	DefaultOverflow    string
	DeadLetterBuffer   int
	SystemBuffer       int
	AgentBuffer        int
	AgentParallelism   int
	WebhookPoolWorkers int
	WebhookPoolQueue   int
}

type MigrationConfig struct {
	WorkerEnabled   bool
	MaxConcurrency  int
	BatchSize       int
	RateLimitPerSec int
}

// Config 返回全局配置中心
func Config() *AppConfig {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "No .env file found, using system environment variables")
		}
		cfg = defaultConfig()
		applyEnvOverrides(cfg)
		cfg.Security.JWTSecret = getJWTSecret()
	})
	return cfg
}

func defaultConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Port: "6277",
			Host: "0.0.0.0",
			Mode: "release",
		},
		Database: DatabaseConfig{
			Type:    "sqlite",
			Path:    "data/ech0.db",
			LogMode: "release",
		},
		Log: LogConfig{
			Level:           "info",
			Format:          "json",
			Console:         false,
			FileEnable:      true,
			FilePath:        "data/app.log",
			FileMaxSize:     100,
			FileMaxBackups:  5,
			FileMaxAge:      30,
			FileCompress:    true,
			BufferSize:      2048,
			RecentSize:      2000,
			DropPolicy:      "drop_oldest",
			FlushBatch:      128,
			FlushIntervalMs: 500,
		},
		Auth: AuthConfig{
			Jwt: JWTConfig{
				Expires:        900,
				RefreshExpires: 604800,
				Issuer:         "ech0",
				Audience:       "ech0",
			},
			Redirect: RedirectConfig{
				AllowedReturnURLs: []string{},
			},
			WebAuthn: WebAuthnConfig{
				RPID:    "",
				Origins: []string{},
			},
		},
		Storage: StorageConfig{
			ObjectEnabled: false,
			DataRoot:      "data/files",
		},
		Upload: UploadConfig{
			ImageMaxSize: 20971520,
			AudioMaxSize: 20971520,
			ImagePath:    "data/files/images/",
			AudioPath:    "data/files/audios/",
			AllowedTypes: []string{
				"image/jpeg",
				"image/png",
				"image/gif",
				"image/webp",
				"image/avif",
				"audio/mpeg",
				"audio/flac",
				"audio/wav",
				"audio/mp4",
			},
		},
		Event: EventConfig{
			DefaultBuffer:      512,
			DefaultOverflow:    "block",
			DeadLetterBuffer:   64,
			SystemBuffer:       64,
			AgentBuffer:        128,
			AgentParallelism:   2,
			WebhookPoolWorkers: 6,
			WebhookPoolQueue:   6,
		},
		Migration: MigrationConfig{
			WorkerEnabled:   false,
			MaxConcurrency:  1,
			BatchSize:       100,
			RateLimitPerSec: 20,
		},
		Setting: SettingConfig{
			SiteTitle:     "Ech0",
			ServerLogo:    "/Ech0.svg",
			Servername:    "Ech0",
			Serverurl:     "https://ech0.example.com",
			AllowRegister: true,
			Icpnumber:     "",
			FooterContent: "",
			FooterLink:    "",
			MetingAPI:     "",
			CustomCSS:     "",
			CustomJS:      "",
		},
		Comment: CommentConfig{
			EnableComment:         false,
			CaptchaSiteKey:        "ech0-comment",
			CaptchaSecret:         "",
			CaptchaDifficulty:     4,
			CaptchaChallengeCount: 80,
			CaptchaSaltSize:       32,
			CaptchaChallengeTTL:   900,
			CaptchaRedeemTTL:      7200,
			CaptchaGCInterval:     2,
			CaptchaEnableCORS:     true,
			CaptchaIPHeader:       "",
			CaptchaMaxBodyBytes:   1048576,
			CaptchaRateLimitMax:   30,
			CaptchaRateLimitWin:   5,
			CaptchaRateLimitScope: "cap",
			CaptchaLimitOnRedeem:  false,
			CaptchaLimitOnVerify:  false,
		},
		Web: WebConfig{
			CORS: CORSConfig{
				AllowedOrigins: []string{},
			},
		},
	}
}

func applyEnvOverrides(cfg *AppConfig) {
	// Server
	setStringEnv("ECH0_SERVER_PORT", &cfg.Server.Port)
	setStringEnv("ECH0_SERVER_HOST", &cfg.Server.Host)
	setStringEnv("ECH0_SERVER_MODE", &cfg.Server.Mode)

	// Database
	setStringEnv("ECH0_DB_TYPE", &cfg.Database.Type)
	setStringEnv("ECH0_DB_PATH", &cfg.Database.Path)
	setStringEnv("ECH0_DB_LOGMODE", &cfg.Database.LogMode)

	// Log
	setStringEnv("ECH0_LOG_LEVEL", &cfg.Log.Level)
	setStringEnv("ECH0_LOG_FORMAT", &cfg.Log.Format)
	setBoolEnv("ECH0_LOG_CONSOLE", &cfg.Log.Console)
	setBoolEnv("ECH0_LOG_FILE_ENABLE", &cfg.Log.FileEnable)
	setStringEnv("ECH0_LOG_FILE_PATH", &cfg.Log.FilePath)
	setIntEnv("ECH0_LOG_FILE_MAX_SIZE", &cfg.Log.FileMaxSize)
	setIntEnv("ECH0_LOG_FILE_MAX_BACKUPS", &cfg.Log.FileMaxBackups)
	setIntEnv("ECH0_LOG_FILE_MAX_AGE", &cfg.Log.FileMaxAge)
	setBoolEnv("ECH0_LOG_FILE_COMPRESS", &cfg.Log.FileCompress)
	setIntEnv("ECH0_LOG_BUFFER_SIZE", &cfg.Log.BufferSize)
	setIntEnv("ECH0_LOG_RECENT_SIZE", &cfg.Log.RecentSize)
	setStringEnv("ECH0_LOG_DROP_POLICY", &cfg.Log.DropPolicy)
	setIntEnv("ECH0_LOG_FLUSH_BATCH", &cfg.Log.FlushBatch)
	setIntEnv("ECH0_LOG_FLUSH_INTERVAL_MS", &cfg.Log.FlushIntervalMs)

	// Auth / JWT
	setIntEnv("ECH0_JWT_EXPIRES", &cfg.Auth.Jwt.Expires)
	setIntEnv("ECH0_JWT_REFRESH_EXPIRES", &cfg.Auth.Jwt.RefreshExpires)
	setStringEnv("ECH0_JWT_ISSUER", &cfg.Auth.Jwt.Issuer)
	setStringEnv("ECH0_JWT_AUDIENCE", &cfg.Auth.Jwt.Audience)
	setStringSliceEnv("ECH0_AUTH_REDIRECT_ALLOWED_RETURN_URLS", &cfg.Auth.Redirect.AllowedReturnURLs)
	setStringEnv("ECH0_AUTH_WEBAUTHN_RP_ID", &cfg.Auth.WebAuthn.RPID)
	setStringSliceEnv("ECH0_AUTH_WEBAUTHN_ORIGINS", &cfg.Auth.WebAuthn.Origins)

	// Upload
	setIntEnv("ECH0_UPLOAD_IMAGE_MAX_SIZE", &cfg.Upload.ImageMaxSize)
	setIntEnv("ECH0_UPLOAD_AUDIO_MAX_SIZE", &cfg.Upload.AudioMaxSize)
	setStringEnv("ECH0_UPLOAD_IMAGE_PATH", &cfg.Upload.ImagePath)
	setStringEnv("ECH0_UPLOAD_AUDIO_PATH", &cfg.Upload.AudioPath)

	// Storage (local)
	setBoolEnv("ECH0_OBJECT_ENABLED", &cfg.Storage.ObjectEnabled)
	setStringEnv("ECH0_STORAGE_DATA_ROOT", &cfg.Storage.DataRoot)

	// Storage (S3-compatible)
	setStringEnv("ECH0_S3_ENDPOINT", &cfg.Storage.Endpoint)
	setStringEnv("ECH0_S3_ACCESS_KEY", &cfg.Storage.AccessKey)
	setStringEnv("ECH0_S3_SECRET_KEY", &cfg.Storage.SecretKey)
	setStringEnv("ECH0_S3_BUCKET", &cfg.Storage.BucketName)
	setStringEnv("ECH0_S3_REGION", &cfg.Storage.Region)
	setStringEnv("ECH0_S3_PROVIDER", &cfg.Storage.Provider)
	setBoolEnv("ECH0_S3_USE_SSL", &cfg.Storage.UseSSL)
	setStringEnv("ECH0_S3_CDN_URL", &cfg.Storage.CDNURL)
	setStringEnv("ECH0_S3_PATH_PREFIX", &cfg.Storage.PathPrefix)

	// Event
	setIntEnv("ECH0_EVENT_DEFAULT_BUFFER", &cfg.Event.DefaultBuffer)
	setStringEnv("ECH0_EVENT_DEFAULT_OVERFLOW", &cfg.Event.DefaultOverflow)
	setIntEnv("ECH0_EVENT_DEADLETTER_BUFFER", &cfg.Event.DeadLetterBuffer)
	setIntEnv("ECH0_EVENT_SYSTEM_BUFFER", &cfg.Event.SystemBuffer)
	setIntEnv("ECH0_EVENT_AGENT_BUFFER", &cfg.Event.AgentBuffer)
	setIntEnv("ECH0_EVENT_AGENT_PARALLELISM", &cfg.Event.AgentParallelism)
	setIntEnv("ECH0_EVENT_WEBHOOK_POOL_WORKERS", &cfg.Event.WebhookPoolWorkers)
	setIntEnv("ECH0_EVENT_WEBHOOK_POOL_QUEUE", &cfg.Event.WebhookPoolQueue)

	// Migration
	setBoolEnv("ECH0_MIGRATION_WORKER_ENABLED", &cfg.Migration.WorkerEnabled)
	setIntEnv("ECH0_MIGRATION_MAX_CONCURRENCY", &cfg.Migration.MaxConcurrency)
	setIntEnv("ECH0_MIGRATION_BATCH_SIZE", &cfg.Migration.BatchSize)
	setIntEnv("ECH0_MIGRATION_RATE_LIMIT_PER_SEC", &cfg.Migration.RateLimitPerSec)

	// Setting
	setStringEnv("ECH0_SETTING_SITE_TITLE", &cfg.Setting.SiteTitle)
	setStringEnv("ECH0_SETTING_SERVER_LOGO", &cfg.Setting.ServerLogo)
	setStringEnv("ECH0_SETTING_SERVER_NAME", &cfg.Setting.Servername)
	setStringEnv("ECH0_SETTING_SERVER_URL", &cfg.Setting.Serverurl)
	setBoolEnv("ECH0_SETTING_ALLOW_REGISTER", &cfg.Setting.AllowRegister)
	setStringEnv("ECH0_SETTING_ICP_NUMBER", &cfg.Setting.Icpnumber)
	setStringEnv("ECH0_SETTING_FOOTER_CONTENT", &cfg.Setting.FooterContent)
	setStringEnv("ECH0_SETTING_FOOTER_LINK", &cfg.Setting.FooterLink)
	setStringEnv("ECH0_SETTING_METING_API", &cfg.Setting.MetingAPI)
	setStringEnv("ECH0_SETTING_CUSTOM_CSS", &cfg.Setting.CustomCSS)
	setStringEnv("ECH0_SETTING_CUSTOM_JS", &cfg.Setting.CustomJS)

	// Comment
	setBoolEnv("ECH0_COMMENT_ENABLE", &cfg.Comment.EnableComment)
	setStringEnv("ECH0_COMMENT_CAPTCHA_SITE_KEY", &cfg.Comment.CaptchaSiteKey)
	setStringEnv("ECH0_COMMENT_CAPTCHA_SECRET", &cfg.Comment.CaptchaSecret)
	setIntEnv("ECH0_COMMENT_CAPTCHA_DIFFICULTY", &cfg.Comment.CaptchaDifficulty)
	setIntEnv("ECH0_COMMENT_CAPTCHA_CHALLENGE_COUNT", &cfg.Comment.CaptchaChallengeCount)
	setIntEnv("ECH0_COMMENT_CAPTCHA_SALT_SIZE", &cfg.Comment.CaptchaSaltSize)
	setIntEnv("ECH0_COMMENT_CAPTCHA_CHALLENGE_TTL", &cfg.Comment.CaptchaChallengeTTL)
	setIntEnv("ECH0_COMMENT_CAPTCHA_REDEEM_TTL", &cfg.Comment.CaptchaRedeemTTL)
	setIntEnv("ECH0_COMMENT_CAPTCHA_GC_INTERVAL", &cfg.Comment.CaptchaGCInterval)
	setBoolEnv("ECH0_COMMENT_CAPTCHA_ENABLE_CORS", &cfg.Comment.CaptchaEnableCORS)
	setStringEnv("ECH0_COMMENT_CAPTCHA_IP_HEADER", &cfg.Comment.CaptchaIPHeader)
	setIntEnv("ECH0_COMMENT_CAPTCHA_MAX_BODY_BYTES", &cfg.Comment.CaptchaMaxBodyBytes)
	setIntEnv("ECH0_COMMENT_CAPTCHA_RATE_LIMIT_MAX", &cfg.Comment.CaptchaRateLimitMax)
	setIntEnv("ECH0_COMMENT_CAPTCHA_RATE_LIMIT_WINDOW", &cfg.Comment.CaptchaRateLimitWin)
	setStringEnv("ECH0_COMMENT_CAPTCHA_RATE_LIMIT_SCOPE", &cfg.Comment.CaptchaRateLimitScope)
	setBoolEnv("ECH0_COMMENT_CAPTCHA_RATE_LIMIT_ON_REDEEM", &cfg.Comment.CaptchaLimitOnRedeem)
	setBoolEnv("ECH0_COMMENT_CAPTCHA_RATE_LIMIT_ON_SITEVERIFY", &cfg.Comment.CaptchaLimitOnVerify)

	// Web/CORS
	setStringSliceEnv("ECH0_WEB_CORS_ALLOWED_ORIGINS", &cfg.Web.CORS.AllowedOrigins)
}

func setStringEnv(key string, target *string) {
	if value := os.Getenv(key); value != "" {
		*target = value
	}
}

func setBoolEnv(key string, target *bool) {
	value := os.Getenv(key)
	if value == "" {
		return
	}
	parsed, err := strconv.ParseBool(value)
	if err == nil {
		*target = parsed
	}
}

func setIntEnv(key string, target *int) {
	value := os.Getenv(key)
	if value == "" {
		return
	}
	parsed, err := strconv.Atoi(value)
	if err == nil {
		*target = parsed
	}
}

func setStringSliceEnv(key string, target *[]string) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		item := strings.TrimSpace(p)
		if item != "" {
			result = append(result, item)
		}
	}
	if len(result) > 0 {
		*target = result
	}
}

// getJWTSecret 加载JWT密钥
func getJWTSecret() []byte {
	// 从环境变量中获取JWT密钥
	secret := os.Getenv("JWT_SECRET")
	if secret == "" { // 如果没有设置环境变量，则使用UUID生成默认密钥
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random JWT secret: %v", err))
		}
		secret = hex.EncodeToString(b)
	}

	return []byte(secret)
}
