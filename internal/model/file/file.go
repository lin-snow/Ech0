// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

// storageTypeExternal marks files that live outside Ech0's managed storage
// (a user-supplied external URL). Their URL is authoritative and is never
// recomputed — there is no key to rebuild it from.
const storageTypeExternal = "external"

// resolveURL recomputes a managed file's public URL from its StorageType + Key
// using the *current* storage config. It is injected once at startup via
// RegisterURLResolver (by the storage layer). When nil — e.g. tests or the CLI
// without storage wiring — the stored URL snapshot is kept as-is.
var resolveURL func(storageType, key string) string

// RegisterURLResolver wires the storage URL resolver into the File model so
// AfterFind can refresh URLs on read. Called once from the composition root;
// safe to call again to replace (or with nil to disable, e.g. in tests).
func RegisterURLResolver(fn func(storageType, key string) string) {
	resolveURL = fn
}

type File struct {
	ID string `gorm:"type:char(36);primaryKey" json:"id"`

	// 存储键（本地文件名或对象存储 object key）
	Key string `gorm:"type:varchar(500);not null;uniqueIndex:idx_file_route,priority:4" json:"key"`

	StorageType string `gorm:"type:varchar(20);not null;uniqueIndex:idx_file_route,priority:1" json:"storage_type"` // local|object|external
	Provider    string `gorm:"type:varchar(50);uniqueIndex:idx_file_route,priority:2" json:"provider,omitempty"`    // object 提供商，如 aws/r2/minio/external
	Bucket      string `gorm:"type:varchar(120);uniqueIndex:idx_file_route,priority:3" json:"bucket,omitempty"`     // local/external 可空

	URL         string `gorm:"type:text" json:"url"` // 前端直链快照
	Name        string `gorm:"type:varchar(255)" json:"name"`
	ContentType string `gorm:"type:varchar(100)" json:"content_type,omitempty"`
	Size        int64  `gorm:"default:0" json:"size"`
	Width       int    `gorm:"default:0" json:"width,omitempty"`
	Height      int    `gorm:"default:0" json:"height,omitempty"`

	Category  string `gorm:"type:varchar(20);index" json:"category"` // image|video|audio|pdf|markdown|file，见 storage.Category
	UserID    string `gorm:"type:char(36);index;not null" json:"user_id"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
}

// EchoFile links a File to an Echo with ordering support.
type EchoFile struct {
	ID        string `gorm:"type:char(36);primaryKey"                        json:"id"`
	EchoID    string `gorm:"type:char(36);uniqueIndex:idx_echo_file;not null" json:"echo_id"`
	FileID    string `gorm:"type:char(36);uniqueIndex:idx_echo_file;not null" json:"file_id"`
	File      File   `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	SortOrder int    `gorm:"default:0"                                   json:"sort_order"`
}

// TempFile tracks uploaded files that are pending business confirmation.
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
		f.ID = uuidUtil.MustNewV7()
	}
	return nil
}

// AfterFind refreshes URL for managed (local/object) files from the current
// storage config, so rotating the CDN domain or switching S3 settings never
// strands existing rows behind the write-time snapshot. External files keep
// their stored URL (the source of truth). Fires for both direct loads and
// preloaded associations, making it the single read-path choke point.
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
		e.ID = uuidUtil.MustNewV7()
	}
	return nil
}

func (t *TempFile) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuidUtil.MustNewV7()
	}
	return nil
}
