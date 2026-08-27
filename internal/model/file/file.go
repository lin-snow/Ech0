// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

const storageTypeExternal = "external"

var resolveURL func(storageType, key string) string

func RegisterURLResolver(fn func(storageType, key string) string) {
	resolveURL = fn
}

type File struct {
	ID string `gorm:"type:char(36);primaryKey" json:"id"`

	Key string `gorm:"type:varchar(500);not null;uniqueIndex:idx_file_route,priority:4" json:"key"`

	StorageType string `gorm:"type:varchar(20);not null;uniqueIndex:idx_file_route,priority:1" json:"storage_type"`
	Provider    string `gorm:"type:varchar(50);uniqueIndex:idx_file_route,priority:2" json:"provider,omitempty"`
	Bucket      string `gorm:"type:varchar(120);uniqueIndex:idx_file_route,priority:3" json:"bucket,omitempty"`

	URL         string `gorm:"type:text" json:"url"`
	Name        string `gorm:"type:varchar(255)" json:"name"`
	ContentType string `gorm:"type:varchar(100)" json:"content_type,omitempty"`
	Size        int64  `gorm:"default:0" json:"size"`
	Width       int    `gorm:"default:0" json:"width,omitempty"`
	Height      int    `gorm:"default:0" json:"height,omitempty"`

	Category  string `gorm:"type:varchar(20);index" json:"category"`
	UserID    string `gorm:"type:char(36);index;not null" json:"user_id"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
}

type EchoFile struct {
	ID        string `gorm:"type:char(36);primaryKey"                        json:"id"`
	EchoID    string `gorm:"type:char(36);uniqueIndex:idx_echo_file;not null" json:"echo_id"`
	FileID    string `gorm:"type:char(36);uniqueIndex:idx_echo_file;not null" json:"file_id"`
	File      File   `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	SortOrder int    `gorm:"default:0"                                   json:"sort_order"`
}

type TempFile struct {
	ID         string `gorm:"type:char(36);primaryKey"               json:"id"`
	FileID     string `gorm:"type:char(36);not null;uniqueIndex"     json:"file_id"`
	File       File   `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	UploaderID string `gorm:"type:char(36);index;not null"            json:"uploader_id"`
	ExpireAt   int64  `gorm:"index;not null"                          json:"expire_at"`
	CreatedAt  int64  `gorm:"autoCreateTime;index"                    json:"created_at"`
}

func (f *File) BeforeCreate(_ *gorm.DB) error {
	if f.ID == "" {
		f.ID = uuidUtil.NewV7()
	}
	return nil
}

func (f *File) AfterFind(_ *gorm.DB) error {
	if resolveURL == nil || f.Key == "" || f.StorageType == storageTypeExternal {
		return nil
	}
	if url := resolveURL(f.StorageType, f.Key); url != "" {
		f.URL = url
	}
	return nil
}

func (e *EchoFile) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuidUtil.NewV7()
	}
	return nil
}

func (t *TempFile) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuidUtil.NewV7()
	}
	return nil
}
