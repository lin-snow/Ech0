package model

import (
	"time"

	fileModel "github.com/lin-snow/ech0/internal/model/file"
)

// Echo 定义Echo实体
type Echo struct {
	ID            uint                 `gorm:"primaryKey"                                    json:"id"`
	Content       string               `gorm:"type:text;not null"                            json:"content"`
	Username      string               `gorm:"type:varchar(100)"                             json:"username,omitempty"`
	EchoFiles     []fileModel.EchoFile `gorm:"foreignKey:EchoID;constraint:OnDelete:CASCADE" json:"echo_files,omitempty"`
	Layout        string               `gorm:"type:varchar(50);default:'waterfall'"          json:"layout,omitempty"`
	Private       bool                 `gorm:"default:false"                                 json:"private"`
	UserID        uint                 `gorm:"not null;index"                                json:"user_id"`
	Extension     string               `gorm:"type:text"                                     json:"extension,omitempty"`
	ExtensionType string               `gorm:"type:varchar(100)"                             json:"extension_type,omitempty"`
	Tags          []Tag                `gorm:"many2many:echo_tags;"                          json:"tags,omitempty"`
	FavCount      int                  `gorm:"default:0"                                     json:"fav_count"`
	CreatedAt     time.Time            `                                                     json:"created_at"`
}

// Tag 定义Tag实体
type Tag struct {
	ID         uint      `gorm:"primaryKey"                            json:"id"`
	Name       string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	UsageCount int       `gorm:"default:0"                             json:"usage_count"`
	CreatedAt  time.Time `                                             json:"created_at"`
}

// EchoTag 纯关系表，联合主键
type EchoTag struct {
	EchoID uint `gorm:"primaryKey;autoIncrement:false"`
	TagID  uint `gorm:"primaryKey;autoIncrement:false"`
}

const (
	Extension_MUSIC      = "MUSIC"
	Extension_VIDEO      = "VIDEO"
	Extension_GITHUBPROJ = "GITHUBPROJ"
	Extension_WEBSITE    = "WEBSITE"

	LayoutWaterfall  = "waterfall"
	LayoutGrid       = "grid"
	LayoutHorizontal = "horizontal"
	LayoutCarousel   = "carousel"
)
