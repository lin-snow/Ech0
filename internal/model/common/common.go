package model

// File 相关
type UploadFileType string
type FileStorageType string

const (
	ImageType UploadFileType = "image"
	AudioType UploadFileType = "audio"
)

const (
	LOCAL_FILE FileStorageType = "local"
	S3_FILE    FileStorageType = "s3"
	R2_FILE    FileStorageType = "r2"
)

// key value表
type KeyValue struct {
	Key   string `json:"key" gorm:"primaryKey"`
	Value string `json:"value"`
}

// 键值对相关
const (
	SystemSettingsKey = "system_settings" // 系统设置的键
	ConnectKey        = "connect"         // Connect 信息的键
)
