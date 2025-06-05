package model

import "time"

type Echo struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	Username      string    `gorm:"type:varchar(100)" json:"username,omitempty"`
	Images        []Image   `gorm:"foreignKey:MessageID" json:"images,omitempty"`
	Private       bool      `gorm:"default:false" json:"private"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Extension     string    `gorm:"type:text" json:"extension,omitempty"`
	ExtensionType string    `gorm:"type:varchar(100)" json:"extension_type,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

type Image struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	MessageID   uint   `gorm:"index;not null" json:"message_id"`
	ImageURL    string `gorm:"type:text" json:"image_url"`
	ImageSource string `gorm:"type:varchar(20)" json:"image_source"`
}
