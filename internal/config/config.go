package config

import (
	"crypto/rand"
	"encoding/hex"
	model "github.com/lin-snow/ech0/internal/model/common"
	"github.com/spf13/viper"
	"log"
	"os"
)

var Config AppConfig
var JWT_SECRET []byte

type AppConfig struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`
	Database struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"database"`
	Auth struct {
		Jwt struct {
			Expires  int    `yaml:"expires"`
			Issuer   string `yaml:"issuer"`
			Audience string `yaml:"audience"`
		} `yaml:"jwt"`
	} `yaml:"auth"`
	Upload struct {
		ImageMaxSize int      `yaml:"imagemaxsize"`
		AudioMaxSize int      `yaml:"audiomaxsize"`
		AllowedTypes []string `yaml:"allowedtypes"`
		ImagePath    string   `yaml:"imagepath"`
		AudioPath    string   `yaml:"audiopath"`
	} `yaml:"upload"`
	Setting struct {
		SiteTitle     string `yaml:"sitetitle"`
		Servername    string `yaml:"servername"`
		Serverurl     string `yaml:"serverurl"`
		AllowRegister bool   `yaml:"allowregister"`
		Icpnumber     string `yaml:"icpnumber"`
		MetingAPI     string `yaml:"metingapi"`
		CustomCSS     string `yaml:"customcss"`
		CustomJS      string `yaml:"customjs"`
	}
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	DBType string `yaml:"type"`
	DBPath string `yaml:"path"`
}

func LoadAppConfig() {
	viper.SetConfigFile("config/config.yaml")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(model.READ_CONFIG_PANIC + ":" + err.Error())
	}

	// 将配置文件内容反序列化到结构体 Config 中
	err = viper.Unmarshal(&Config)
	if err != nil {
		panic(model.READ_CONFIG_PANIC + ":" + err.Error())
	}

	JWT_SECRET = GetJWTSecret()
}

// GetJWTSecret 加载JWT密钥
func GetJWTSecret() []byte {
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
