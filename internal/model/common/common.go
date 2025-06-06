package model

type UserStatus struct {
	UserID   uint   `json:"user_id"`  // 用户ID
	UserName string `json:"username"` // 用户名
	IsAdmin  bool   `json:"is_admin"` // 是否是管理员
}
type Status struct {
	SysAdminID    uint         `json:"sys_admin_id"` // 系统管理员ID
	Username      string       `json:"username"`     // 系统管理员用户名
	Users         []UserStatus `json:"users"`        // 所有用户
	Logo          string       `json:"logo"`         // 站点logo
	TotalMessages int          `json:"total_messages"`
}

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

type PageQueryResult[T any] struct {
	Total int64 `json:"total"`
	Items T     `json:"items"`
}

// 其他

const (
	InitInstallCode = 666
)

const (
	Version = "2.0.0" // 当前版本号
)

const (
	GreetingBanner = `
███████╗     ██████╗    ██╗  ██╗     ██████╗ 
██╔════╝    ██╔════╝    ██║  ██║    ██╔═████╗
█████╗      ██║         ███████║    ██║██╔██║
██╔══╝      ██║         ██╔══██║    ████╔╝██║
███████╗    ╚██████╗    ██║  ██║    ╚██████╔╝
╚══════╝     ╚═════╝    ╚═╝  ╚═╝     ╚═════╝ 
                                             
`
)
