package model

import "time"

// Echo 定义Echo实体
type Echo struct {
	ID            uint      `gorm:"primaryKey"                                       json:"id"`
	Content       string    `gorm:"type:text;not null"                               json:"content"`
	Username      string    `gorm:"type:varchar(100)"                                json:"username,omitempty"`
	Images        []Image   `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE" json:"images,omitempty"`
	Layout        string    `gorm:"type:varchar(50);default:'waterfall'"             json:"layout,omitempty"`
	Private       bool      `gorm:"default:false"                                    json:"private"`
	UserID        uint      `gorm:"not null;index"                                   json:"user_id"`
	Extension     string    `gorm:"type:text"                                        json:"extension,omitempty"`
	ExtensionType string    `gorm:"type:varchar(100)"                                json:"extension_type,omitempty"`
	Tags          []Tag     `gorm:"many2many:echo_tags;"                             json:"tags,omitempty"`
	FavCount      int       `gorm:"default:0"                                        json:"fav_count"`
	CreatedAt     time.Time `                                                        json:"created_at"`
}

// Message 定义Message实体 (注意⚠️: 该模型为旧版Echo模型,新版已经弃用)
// type Message struct {
// 	ID            uint      `gorm:"primaryKey"           json:"id"`
// 	Content       string    `gorm:"type:text;not null"   json:"content"`
// 	Username      string    `gorm:"type:varchar(100)"    json:"username,omitempty"`
// 	ImageURL      string    `gorm:"type:text"            json:"image_url,omitempty"`
// 	ImageSource   string    `gorm:"type:varchar(20)"     json:"image_source,omitempty"`
// 	Images        []Image   `gorm:"foreignKey:MessageID" json:"images,omitempty"`
// 	Private       bool      `gorm:"default:false"        json:"private"`
// 	UserID        uint      `gorm:"not null;index"       json:"user_id"`
// 	Extension     string    `gorm:"type:text"            json:"extension,omitempty"`
// 	ExtensionType string    `gorm:"type:varchar(100)"    json:"extension_type,omitempty"`
// 	CreatedAt     time.Time `                            json:"created_at"`
// }

// Image 定义Image实体
type Image struct {
	ID          uint   `gorm:"primaryKey"       json:"id"`
	MessageID   uint   `gorm:"index;not null"   json:"message_id"`           // 关联的Echo ID(注意⚠️: 该字段名为MessageID, 但实际关联的是Echo表,因为为了兼容旧版Echo用户)
	ImageURL    string `gorm:"type:text"        json:"image_url"`            // 图片URL
	ImageSource string `gorm:"type:varchar(20)" json:"image_source"`         // 图片来源: local/url/s3
	ObjectKey   string `gorm:"type:text"        json:"object_key,omitempty"` // 对象存储的Key (如果是本地存储则为空)
	Width       int    `gorm:"default:0"        json:"width,omitempty"`      // 图片宽度
	Height      int    `gorm:"default:0"        json:"height,omitempty"`     // 图片高度
}

// Tag 定义Tag实体
type Tag struct {
	ID         uint      `gorm:"primaryKey"                            json:"id"`
	Name       string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`        // 标签名称
	UsageCount int       `gorm:"default:0"                             json:"usage_count"` // 使用计数
	CreatedAt  time.Time `                                             json:"created_at"`  // 创建时间
}

// EchoTag 纯关系表，联合主键
type EchoTag struct {
	EchoID uint `gorm:"primaryKey;autoIncrement:false"` // Echo ID
	TagID  uint `gorm:"primaryKey;autoIncrement:false"` // Tag ID
}

const (
	Extension_MUSIC      = "MUSIC"      // 扩展附加内容--音乐
	Extension_VIDEO      = "VIDEO"      // 扩展附加内容--视频
	Extension_GITHUBPROJ = "GITHUBPROJ" // 扩展附加内容--GitHub项目
	Extension_WEBSITE    = "WEBSITE"    // 扩展附加内容--网站
	Extension_MODEL3D    = "MODEL3D"    // 扩展附加内容--3D模型

	ImageSourceLocal = "local" // 本地图片
	ImageSourceURL   = "url"   // 直链图片
	ImageSourceS3    = "s3"    // S3 图片

	LayoutWaterfall  = "waterfall"  // 瀑布流布局
	LayoutGrid       = "grid"       // 九宫格布局
	LayoutHorizontal = "horizontal" // 横向布局
	LayoutCarousel   = "carousel"   // 单图轮播布局

)
