// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package config

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/lin-snow/ech0/internal/kvstore"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

var (
	cfg  *AppConfig
	once sync.Once
)

type AppConfig struct {
	Server   ServerConfig
	OpenAPI  OpenAPIConfig
	Database DatabaseConfig
	Log      LogConfig
	Auth     AuthConfig
	Upload   UploadConfig
	Storage  StorageConfig
	Event    EventConfig
	Setting  SettingConfig
	Comment  CommentConfig
	Security SecurityConfig
	Web      WebConfig
	Agent    AgentConfig
	Egress   EgressConfig
}

// EgressConfig controls the outbound HTTP SSRF guard. AllowPrivate exists for
// local development (self-connect, LAN/Tailscale peers) and must stay false in
// production: it disables the private/reserved-address and localhost blocks for
// every egress caller (connect, webhook, OAuth CIMD, site scraping).
type EgressConfig struct {
	AllowPrivate bool `env:"ECH0_EGRESS_ALLOW_PRIVATE"`
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
	ImageMaxSize int `env:"ECH0_UPLOAD_IMAGE_MAX_SIZE"`
	AudioMaxSize int `env:"ECH0_UPLOAD_AUDIO_MAX_SIZE"`
	VideoMaxSize int `env:"ECH0_UPLOAD_VIDEO_MAX_SIZE"`
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

type AgentConfig struct {
	TimeoutSeconds    int `env:"ECH0_AGENT_TIMEOUT_SECONDS"`
	MaxRounds         int `env:"ECH0_AGENT_MAX_ROUNDS"`
	AskTimeoutSeconds int `env:"ECH0_AGENT_ASK_TIMEOUT_SECONDS"`
}

func Config() *AppConfig {
	once.Do(func() {
		loadDotEnv()
		cfg = defaultConfig()
		if err := env.Parse(cfg); err != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Invalid environment overrides (affected values keep their defaults): %v\n", err,
			)
		}
	})
	return cfg
}

// loadDotEnv reads ./.env relative to the working directory. A missing file is
// normal (process env only); an unreadable or malformed file is reported but
// never fatal, so a broken .env cannot prevent startup.
func loadDotEnv() {
	err := godotenv.Load()
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "Failed to load .env: %v\n", err)
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
			Console:         true,
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
			TimeoutSeconds:    300,
			MaxRounds:         20,
			AskTimeoutSeconds: 300,
		},
	}
}

// jwtSecretKey is the settings key-value entry that holds the signing secret.
// It deliberately has no setting spec, so the value can never be serialized
// back to a client.
const jwtSecretKey = "jwt_secret"

// EnsureJWTSecret pins the instance's JWT signing secret before the server
// starts serving. The secret persisted in the settings key-value store is the
// single source of truth: it is read when present and generated once (then
// written back) when it is missing. Persisting it keeps sessions, access
// tokens and refresh tokens valid across restarts and upgrades, and having no
// second input means a removed environment variable can never silently revert
// the instance to an older key.
//
// Call it once, after the database is ready and before any request can be
// authenticated. A failing store is returned as an error rather than silently
// falling back to an ephemeral secret, which would invalidate every issued
// token on the next restart.
func EnsureJWTSecret(ctx context.Context, kv kvstore.Store) error {
	cfg := Config()

	raw, err := kv.Get(ctx, jwtSecretKey)
	if err != nil && !errors.Is(err, kvstore.ErrNotFound) {
		return fmt.Errorf("read jwt secret: %w", err)
	}
	if secret := strings.TrimSpace(raw); secret != "" {
		cfg.Security.JWTSecret = []byte(secret)
		logUtil.Info(
			"Loaded JWT signing secret from the database",
			slog.String("module", "config"),
			slog.String("fingerprint", jwtSecretFingerprint(secret)),
		)
		warnLegacyJWTSecretEnv()
		return nil
	}

	secret := randomJWTSecret()
	if err := kv.Set(ctx, jwtSecretKey, secret); err != nil {
		return fmt.Errorf("persist jwt secret: %w", err)
	}
	// Re-read: a second process starting against a fresh database may have won
	// the write, and both must end up signing with the same key.
	if raw, err := kv.Get(ctx, jwtSecretKey); err == nil {
		if stored := strings.TrimSpace(raw); stored != "" {
			secret = stored
		}
	}
	cfg.Security.JWTSecret = []byte(secret)
	logUtil.Info(
		"Generated JWT signing secret and stored it in the database",
		slog.String("module", "config"),
		slog.String("fingerprint", jwtSecretFingerprint(secret)),
	)
	warnLegacyJWTSecretEnv()
	return nil
}

// warnLegacyJWTSecretEnv reports a JWT_SECRET / ECH0_JWT_SECRET left over from
// earlier versions. The variables are ignored on purpose (the database is the
// only source), but silently ignoring them would leave operators convinced
// their instance is pinned to a key they no longer control.
func warnLegacyJWTSecretEnv() {
	for _, name := range []string{"JWT_SECRET", "ECH0_JWT_SECRET"} {
		if os.Getenv(name) == "" {
			continue
		}
		logUtil.Warn(
			"JWT_SECRET/ECH0_JWT_SECRET is set but no longer used; the signing secret lives in the database. Remove the variable.",
			slog.String("module", "config"),
			slog.String("variable", name),
		)
	}
}

// jwtSecretFingerprint identifies a secret in logs without revealing it, so an
// operator can tell whether two boots (or a restore) used the same key.
func jwtSecretFingerprint(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:4])
}

func randomJWTSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate random JWT secret: %v", err))
	}
	return hex.EncodeToString(b)
}
