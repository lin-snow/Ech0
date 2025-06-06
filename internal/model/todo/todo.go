package model

import "time"

type Todo struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Username  string    `gorm:"type:varchar(100)" json:"username,omitempty"`
	Status    uint      `gorm:"default:0" json:"status"` // 0:未完成 1:已完成
	CreatedAt time.Time `json:"created_at"`
}
