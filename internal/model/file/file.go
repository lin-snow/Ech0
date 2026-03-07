package model

import "time"

// File represents a managed file in the VireFS-backed storage system.
// Key is the VireFS storage key; URL is the pre-resolved access URL
// stored at upload time so that pagination queries return directly
// usable URLs with zero runtime computation.
type File struct {
	ID          uint      `gorm:"primaryKey"                                json:"id"`
	Key         string    `gorm:"type:varchar(500);uniqueIndex;not null"    json:"key"`
	URL         string    `gorm:"type:text"                                 json:"url"`
	Name        string    `gorm:"type:varchar(255)"                         json:"name"`
	ContentType string    `gorm:"type:varchar(100)"                         json:"content_type,omitempty"`
	Size        int64     `gorm:"default:0"                                 json:"size"`
	Category    string    `gorm:"type:varchar(20);index"                    json:"category"`
	Width       int       `gorm:"default:0"                                 json:"width,omitempty"`
	Height      int       `gorm:"default:0"                                 json:"height,omitempty"`
	UserID      uint      `gorm:"index;not null"                            json:"user_id"`
	CreatedAt   time.Time `                                                 json:"created_at"`
}

// EchoFile links a File to an Echo with ordering support.
type EchoFile struct {
	ID        uint `gorm:"primaryKey"                                  json:"id"`
	EchoID    uint `gorm:"uniqueIndex:idx_echo_file;not null"          json:"echo_id"`
	FileID    uint `gorm:"uniqueIndex:idx_echo_file;not null"          json:"file_id"`
	File      File `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	SortOrder int  `gorm:"default:0"                                   json:"sort_order"`
}
