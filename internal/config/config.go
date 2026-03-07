package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var (
	cfg  *AppConfig
	once sync.Once
)

type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Upload   UploadConfig
	Storage  StorageConfig
	Setting  SettingConfig
	Comment  CommentConfig
	Security SecurityConfig
}

type StorageConfig struct {
	Mode       string // "local" or "s3", default "local"
	DataRoot   string // local root directory, default "data/files"
	Endpoint   string // S3-compatible endpoint
	AccessKey  string
	SecretKey  string
	BucketName string
	Region     string
	Provider   string // "aws", "r2", "minio", "other"
	UseSSL     bool
	CDNURL     string
	PathPrefix string
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

type AuthConfig struct {
	Jwt JWTConfig
}

type JWTConfig struct {
	Expires  int    // JWT的过期时间，单位为秒
	Issuer   string // JWT的发行者
	Audience string // JWT的受众
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
	MetingAPI     string // Meting API 地址
	CustomCSS     string // 自定义 CSS 样式
	CustomJS      string // 自定义 JS 脚本
}

type CommentConfig struct {
	EnableComment bool   // 是否启用评论
	Provider      string // 评论提供者
	CommentAPI    string // 评论 API 地址
}

type SecurityConfig struct {
	JWTSecret []byte
}

// Config 返回全局配置中心
func Config() *AppConfig {
	once.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system environment variables")
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
		Auth: AuthConfig{
			Jwt: JWTConfig{
				Expires:  2592000,
				Issuer:   "ech0",
				Audience: "ech0",
			},
		},
		Storage: StorageConfig{
			Mode:     "local",
			DataRoot: "data/files",
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
				"image/svg+xml",
				"image/avif",
				"audio/mpeg",
				"audio/flac",
				"audio/wav",
				"audio/mp4",
			},
		},
		Setting: SettingConfig{
			SiteTitle:     "Ech0",
			ServerLogo:    "/Ech0.svg",
			Servername:    "Ech0",
			Serverurl:     "https://ech0.example.com",
			AllowRegister: true,
			Icpnumber:     "",
			MetingAPI:     "",
			CustomCSS:     "",
			CustomJS:      "",
		},
		Comment: CommentConfig{
			EnableComment: false,
			Provider:      "twikoo",
			CommentAPI:    "",
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

	// Auth / JWT
	setIntEnv("ECH0_JWT_EXPIRES", &cfg.Auth.Jwt.Expires)
	setStringEnv("ECH0_JWT_ISSUER", &cfg.Auth.Jwt.Issuer)
	setStringEnv("ECH0_JWT_AUDIENCE", &cfg.Auth.Jwt.Audience)

	// Upload
	setIntEnv("ECH0_UPLOAD_IMAGE_MAX_SIZE", &cfg.Upload.ImageMaxSize)
	setIntEnv("ECH0_UPLOAD_AUDIO_MAX_SIZE", &cfg.Upload.AudioMaxSize)
	setStringEnv("ECH0_UPLOAD_IMAGE_PATH", &cfg.Upload.ImagePath)
	setStringEnv("ECH0_UPLOAD_AUDIO_PATH", &cfg.Upload.AudioPath)

	// Storage (local)
	setStringEnv("ECH0_STORAGE_MODE", &cfg.Storage.Mode)
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

	// Setting
	setStringEnv("ECH0_SETTING_SITE_TITLE", &cfg.Setting.SiteTitle)
	setStringEnv("ECH0_SETTING_SERVER_LOGO", &cfg.Setting.ServerLogo)
	setStringEnv("ECH0_SETTING_SERVER_NAME", &cfg.Setting.Servername)
	setStringEnv("ECH0_SETTING_SERVER_URL", &cfg.Setting.Serverurl)
	setBoolEnv("ECH0_SETTING_ALLOW_REGISTER", &cfg.Setting.AllowRegister)
	setStringEnv("ECH0_SETTING_ICP_NUMBER", &cfg.Setting.Icpnumber)
	setStringEnv("ECH0_SETTING_METING_API", &cfg.Setting.MetingAPI)
	setStringEnv("ECH0_SETTING_CUSTOM_CSS", &cfg.Setting.CustomCSS)
	setStringEnv("ECH0_SETTING_CUSTOM_JS", &cfg.Setting.CustomJS)

	// Comment
	setBoolEnv("ECH0_COMMENT_ENABLE", &cfg.Comment.EnableComment)
	setStringEnv("ECH0_COMMENT_PROVIDER", &cfg.Comment.Provider)
	setStringEnv("ECH0_COMMENT_API", &cfg.Comment.CommentAPI)
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

// getJWTSecret 加载JWT密钥
func getJWTSecret() []byte {
	// 从环境变量中获取JWT密钥
	secret := os.Getenv("JWT_SECRET")
	if secret == "" { // 如果没有设置环境变量，则使用UUID生成默认密钥
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			log.Fatal("failed to generate random JWT secret:", err)
		}
		secret = hex.EncodeToString(b)
	}

	return []byte(secret)
}
