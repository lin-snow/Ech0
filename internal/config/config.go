// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

var (
	cfg  *AppConfig
	once sync.Once
)

type AppConfig struct {
	Server    ServerConfig
	OpenAPI   OpenAPIConfig
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
	Agent     AgentConfig
}

type StorageConfig struct {
	ObjectEnabled bool   `env:"ECH0_OBJECT_ENABLED"`
	DataRoot      string `env:"ECH0_STORAGE_DATA_ROOT"`
	Endpoint      string `env:"ECH0_S3_ENDPOINT"`
	AccessKey     string `env:"ECH0_S3_ACCESS_KEY"`
	SecretKey     string `env:"ECH0_S3_SECRET_KEY"`
	BucketName    string `env:"ECH0_S3_BUCKET"`
	Region        string `env:"ECH0_S3_REGION"`
	Provider      string `env:"ECH0_S3_PROVIDER"`
	UseSSL        bool   `env:"ECH0_S3_USE_SSL"`
	UsePathStyle  bool   `env:"ECH0_S3_USE_PATH_STYLE"`
	CDNURL        string `env:"ECH0_S3_CDN_URL"`
	PathPrefix    string `env:"ECH0_S3_PATH_PREFIX"`
}

type ServerConfig struct {
	Port string `env:"ECH0_SERVER_PORT"`
	Host string `env:"ECH0_SERVER_HOST"`
	Mode string `env:"ECH0_SERVER_MODE"`
}

type OpenAPIConfig struct {
	DocsRenderer string `env:"ECH0_OPENAPI_DOCS_RENDERER"`
}

type DatabaseConfig struct {
	Type    string `env:"ECH0_DB_TYPE"`
	Path    string `env:"ECH0_DB_PATH"`
	LogMode string `env:"ECH0_DB_LOGMODE"`
}

type LogConfig struct {
	Level           string `env:"ECH0_LOG_LEVEL"`
	Format          string `env:"ECH0_LOG_FORMAT"`
	Console         bool   `env:"ECH0_LOG_CONSOLE"`
	FileEnable      bool   `env:"ECH0_LOG_FILE_ENABLE"`
	FilePath        string `env:"ECH0_LOG_FILE_PATH"`
	FileMaxSize     int    `env:"ECH0_LOG_FILE_MAX_SIZE"`
	FileMaxBackups  int    `env:"ECH0_LOG_FILE_MAX_BACKUPS"`
	FileMaxAge      int    `env:"ECH0_LOG_FILE_MAX_AGE"`
	FileCompress    bool   `env:"ECH0_LOG_FILE_COMPRESS"`
	BufferSize      int    `env:"ECH0_LOG_BUFFER_SIZE"`
	RecentSize      int    `env:"ECH0_LOG_RECENT_SIZE"`
	DropPolicy      string `env:"ECH0_LOG_DROP_POLICY"`
	FlushBatch      int    `env:"ECH0_LOG_FLUSH_BATCH"`
	FlushIntervalMs int    `env:"ECH0_LOG_FLUSH_INTERVAL_MS"`
}

type AuthConfig struct {
	Jwt      JWTConfig
	Redirect RedirectConfig
	WebAuthn WebAuthnConfig
}

type JWTConfig struct {
	Expires        int    `env:"ECH0_JWT_EXPIRES"`
	RefreshExpires int    `env:"ECH0_JWT_REFRESH_EXPIRES"`
	Issuer         string `env:"ECH0_JWT_ISSUER"`
	Audience       string `env:"ECH0_JWT_AUDIENCE"`
}

type RedirectConfig struct {
	AllowedReturnURLs []string `env:"ECH0_AUTH_REDIRECT_ALLOWED_RETURN_URLS" envSeparator:","`
}

type WebAuthnConfig struct {
	RPID    string   `env:"ECH0_AUTH_WEBAUTHN_RP_ID"`
	Origins []string `env:"ECH0_AUTH_WEBAUTHN_ORIGINS" envSeparator:","`
}

type UploadConfig struct {
	AllowedTypes []string
	ImageMaxSize int    `env:"ECH0_UPLOAD_IMAGE_MAX_SIZE"`
	AudioMaxSize int    `env:"ECH0_UPLOAD_AUDIO_MAX_SIZE"`
	VideoMaxSize int    `env:"ECH0_UPLOAD_VIDEO_MAX_SIZE"`
	ImagePath    string `env:"ECH0_UPLOAD_IMAGE_PATH"`
	AudioPath    string `env:"ECH0_UPLOAD_AUDIO_PATH"`
	VideoPath    string `env:"ECH0_UPLOAD_VIDEO_PATH"`
}

type SettingConfig struct {
	SiteTitle     string `env:"ECH0_SETTING_SITE_TITLE"`
	ServerLogo    string `env:"ECH0_SETTING_SERVER_LOGO"`
	Servername    string `env:"ECH0_SETTING_SERVER_NAME"`
	Serverurl     string `env:"ECH0_SETTING_SERVER_URL"`
	AllowRegister bool   `env:"ECH0_SETTING_ALLOW_REGISTER"`
	Icpnumber     string `env:"ECH0_SETTING_ICP_NUMBER"`
	FooterContent string `env:"ECH0_SETTING_FOOTER_CONTENT"`
	FooterLink    string `env:"ECH0_SETTING_FOOTER_LINK"`
	MetingAPI     string `env:"ECH0_SETTING_METING_API"`
	CustomCSS     string `env:"ECH0_SETTING_CUSTOM_CSS"`
	CustomJS      string `env:"ECH0_SETTING_CUSTOM_JS"`
}

type CommentConfig struct {
	EnableComment         bool   `env:"ECH0_COMMENT_ENABLE"`
	CaptchaSiteKey        string `env:"ECH0_COMMENT_CAPTCHA_SITE_KEY"`
	CaptchaSecret         string `env:"ECH0_COMMENT_CAPTCHA_SECRET"`
	CaptchaDifficulty     int    `env:"ECH0_COMMENT_CAPTCHA_DIFFICULTY"`
	CaptchaChallengeCount int    `env:"ECH0_COMMENT_CAPTCHA_CHALLENGE_COUNT"`
	CaptchaSaltSize       int    `env:"ECH0_COMMENT_CAPTCHA_SALT_SIZE"`
	CaptchaChallengeTTL   int    `env:"ECH0_COMMENT_CAPTCHA_CHALLENGE_TTL"`
	CaptchaRedeemTTL      int    `env:"ECH0_COMMENT_CAPTCHA_REDEEM_TTL"`
	CaptchaGCInterval     int    `env:"ECH0_COMMENT_CAPTCHA_GC_INTERVAL"`
	CaptchaEnableCORS     bool   `env:"ECH0_COMMENT_CAPTCHA_ENABLE_CORS"`
	CaptchaIPHeader       string `env:"ECH0_COMMENT_CAPTCHA_IP_HEADER"`
	CaptchaMaxBodyBytes   int    `env:"ECH0_COMMENT_CAPTCHA_MAX_BODY_BYTES"`
	CaptchaRateLimitMax   int    `env:"ECH0_COMMENT_CAPTCHA_RATE_LIMIT_MAX"`
	CaptchaRateLimitWin   int    `env:"ECH0_COMMENT_CAPTCHA_RATE_LIMIT_WINDOW"`
	CaptchaRateLimitScope string `env:"ECH0_COMMENT_CAPTCHA_RATE_LIMIT_SCOPE"`
	CaptchaLimitOnRedeem  bool   `env:"ECH0_COMMENT_CAPTCHA_RATE_LIMIT_ON_REDEEM"`
	CaptchaLimitOnVerify  bool   `env:"ECH0_COMMENT_CAPTCHA_RATE_LIMIT_ON_SITEVERIFY"`
}

type SecurityConfig struct {
	JWTSecret []byte
}

type WebConfig struct {
	CORS CORSConfig
}

type CORSConfig struct {
	AllowedOrigins []string `env:"ECH0_WEB_CORS_ALLOWED_ORIGINS" envSeparator:","`
}

type EventConfig struct {
	DefaultBuffer      int    `env:"ECH0_EVENT_DEFAULT_BUFFER"`
	DefaultOverflow    string `env:"ECH0_EVENT_DEFAULT_OVERFLOW"`
	SystemBuffer       int    `env:"ECH0_EVENT_SYSTEM_BUFFER"`
	AgentBuffer        int    `env:"ECH0_EVENT_AGENT_BUFFER"`
	AgentParallelism   int    `env:"ECH0_EVENT_AGENT_PARALLELISM"`
	WebhookPoolWorkers int    `env:"ECH0_EVENT_WEBHOOK_POOL_WORKERS"`
	WebhookPoolQueue   int    `env:"ECH0_EVENT_WEBHOOK_POOL_QUEUE"`
}

type MigrationConfig struct {
	WorkerEnabled   bool `env:"ECH0_MIGRATION_WORKER_ENABLED"`
	MaxConcurrency  int  `env:"ECH0_MIGRATION_MAX_CONCURRENCY"`
	BatchSize       int  `env:"ECH0_MIGRATION_BATCH_SIZE"`
	RateLimitPerSec int  `env:"ECH0_MIGRATION_RATE_LIMIT_PER_SEC"`
}

type AgentConfig struct {
	TimeoutSeconds int `env:"ECH0_AGENT_TIMEOUT_SECONDS"`
	MaxRounds      int `env:"ECH0_AGENT_MAX_ROUNDS"`
}

func Config() *AppConfig {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "No .env file found, using system environment variables")
		}
		cfg = defaultConfig()
		if err := env.Parse(cfg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to parse env overrides: %v\n", err)
		}
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
		OpenAPI: OpenAPIConfig{
			DocsRenderer: "stoplight",
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
			VideoMaxSize: 67108864,
			ImagePath:    "data/files/images/",
			AudioPath:    "data/files/audios/",
			VideoPath:    "data/files/videos/",
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
				"video/mp4",
			},
		},
		Event: EventConfig{
			DefaultBuffer:      512,
			DefaultOverflow:    "block",
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
		Agent: AgentConfig{
			TimeoutSeconds: 120,
			MaxRounds:      4,
		},
	}
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random JWT secret: %v", err))
		}
		secret = hex.EncodeToString(b)
	}

	return []byte(secret)
}
