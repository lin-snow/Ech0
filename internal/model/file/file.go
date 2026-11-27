// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	"strings"
	"time"
)

// Kind is how a file row locates its bytes. It is set once at creation and
// never edited: content is immutable, so a file never changes kind.
type Kind string

const (
	// KindOwned bytes live in a storage provider, addressed by storage_id +
	// object_key.
	KindOwned Kind = "owned"
	// KindBlob bytes live in file_blobs in the same database transaction.
	KindBlob Kind = "blob"
	// KindRef bytes live outside the instance, addressed by ref_url.
	KindRef Kind = "ref"
)

// Access is the rule that decides who may read the bytes. It is a fact about
// the file, not about its owner: public is readable by anyone holding the id,
// owner delegates to the owning domain, private is readable by its uploader
// (and administrators) only.
type Access string

const (
	AccessPublic  Access = "public"
	AccessOwner   Access = "owner"
	AccessPrivate Access = "private"
)

// Status marks whether the bytes are known to be complete. pending exists for
// direct uploads only; every other creation path inserts ready.
type Status string

const (
	StatusPending Status = "pending"
	StatusReady   Status = "ready"
)

// File is the single file resource. One row is one byte resource with one
// stable id; the id is the only thing a business table stores.
type File struct {
	ID     string `gorm:"type:char(36);primaryKey" json:"id"`
	Kind   string `gorm:"type:varchar(10);not null" json:"kind"`
	Status string `gorm:"type:varchar(10);not null" json:"status"`
	Access string `gorm:"type:varchar(10);not null" json:"access"`

	// Owned locator. NULL for blob and ref.
	StorageID *string `gorm:"type:char(36)" json:"storage_id,omitempty"`
	ObjectKey *string `gorm:"type:varchar(500)" json:"object_key,omitempty"`

	// Blob size mirrors file_blobs.data length so listings never read blobs.
	BlobBytes *int64 `gorm:"column:blob_bytes" json:"blob_bytes,omitempty"`

	// Ref locator. NULL for owned and blob.
	RefURL *string `gorm:"column:ref_url;type:text" json:"ref_url,omitempty"`

	Name        string `gorm:"type:varchar(255)" json:"name"`
	ContentType string `gorm:"type:varchar(100);not null" json:"content_type"`
	Size        int64  `gorm:"not null;default:0" json:"size"`
	Width       int    `gorm:"default:0" json:"width,omitempty"`
	Height      int    `gorm:"default:0" json:"height,omitempty"`

	// UploaderID is who may manage the file and who access=private trusts.
	// Administrator override is a role check, never a second column.
	UploaderID string `gorm:"column:uploader_id;type:char(36);not null;index" json:"uploader_id"`

	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime" json:"updated_at"`
}

// Blob holds the bytes of a kind=blob file. It shares the file's primary key,
// so "file without bytes" is unrepresentable and the foreign key removes both
// rows together.
type Blob struct {
	FileID string `gorm:"column:file_id;type:char(36);primaryKey" json:"file_id"`
	Data   []byte `gorm:"type:blob;not null" json:"-"`
}

func (Blob) TableName() string { return "file_blobs" }

// Link is a reference from an owning domain to a file. It carries role and
// order — what the domain needs — and nothing else: name, size and type belong
// to the file.
type Link struct {
	FileID    string `gorm:"column:file_id;type:char(36);primaryKey" json:"file_id"`
	OwnerKind string `gorm:"column:owner_kind;type:varchar(20);primaryKey" json:"owner_kind"`
	OwnerID   string `gorm:"column:owner_id;type:char(36);primaryKey" json:"owner_id"`
	Role      string `gorm:"type:varchar(20);primaryKey" json:"role"`
	SortOrder int    `gorm:"column:sort_order;not null;default:0;primaryKey" json:"sort_order"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
}

func (Link) TableName() string { return "file_links" }

// GCTask is one object waiting to be removed. It has no foreign key: the row it
// refers to is already gone, and the task must outlive it.
type GCTask struct {
	ID        string `gorm:"type:char(36);primaryKey" json:"id"`
	StorageID string `gorm:"column:storage_id;type:char(36);not null" json:"storage_id"`
	ObjectKey string `gorm:"column:object_key;type:varchar(500);not null" json:"object_key"`
	Attempts  int    `gorm:"not null;default:0" json:"attempts"`
	LastError string `gorm:"column:last_error;type:text" json:"last_error,omitempty"`
	NextTryAt int64  `gorm:"column:next_try_at;not null;index" json:"next_try_at"`
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"`
}

func (GCTask) TableName() string { return "file_gc" }

// StorageConfig is one configured storage instance. Its config map holds
// driver-specific keys; the file model never interprets them.
type StorageConfig struct {
	ID        string            `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	Driver    string            `gorm:"type:varchar(20);not null" json:"driver"`
	Name      string            `gorm:"type:varchar(100);not null" json:"name"`
	Config    map[string]string `gorm:"serializer:json;type:text;not null" json:"config"`
	IsDefault bool              `gorm:"column:is_default;not null;default:false" json:"is_default"`
	CreatedAt int64             `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64             `gorm:"autoUpdateTime" json:"updated_at"`
}

func (StorageConfig) TableName() string { return "storages" }

// The owner kinds and roles the core domains write. They are vocabulary, not
// policy: the file system stores them and never branches on them.
const (
	OwnerKindEcho     = "echo"
	OwnerKindUser     = "user"
	OwnerKindInstance = "instance"

	RoleAttachment = "attachment"
	RoleAvatar     = "avatar"
	RoleCover      = "cover"
	RoleLogo       = "logo"
	RoleFavicon    = "favicon"
	RoleOGImage    = "og_image"
)

// The media families a client groups by. They are projections of content_type,
// never stored.
const (
	CategoryImage    = "image"
	CategoryVideo    = "video"
	CategoryAudio    = "audio"
	CategoryPDF      = "pdf"
	CategoryMarkdown = "markdown"
	CategoryFile     = "file"
)

// CategoryOf derives the coarse media family a client groups by. It is a
// projection of content_type, never a stored column: two facts, one value.
func CategoryOf(contentType string) string {
	switch {
	case strings.HasPrefix(contentType, "image/"):
		return CategoryImage
	case strings.HasPrefix(contentType, "video/"):
		return CategoryVideo
	case strings.HasPrefix(contentType, "audio/"), contentType == "application/ogg":
		return CategoryAudio
	case contentType == "application/pdf":
		return CategoryPDF
	case strings.HasPrefix(contentType, "text/markdown"), contentType == "text/plain":
		return CategoryMarkdown
	default:
		return CategoryFile
	}
}

func (f File) IsOwned() bool { return f.Kind == string(KindOwned) }
func (f File) IsBlob() bool  { return f.Kind == string(KindBlob) }
func (f File) IsRef() bool   { return f.Kind == string(KindRef) }

// ContentPath is the single public address of a file's bytes. It is not stored:
// a stored URL is a second truth that rots when the route changes.
func ContentPath(fileID string) string { return "/api/files/" + fileID + "/content" }

// Age reports how long ago the row was created, used by the pending janitor.
func (f File) Age(now time.Time) time.Duration {
	return now.Sub(time.Unix(f.CreatedAt, 0))
}

// ListFilter is the repository query shape of the file browser and of owner
// listings. It lives in the model so the service can declare a repository
// interface without importing the repository package.
type ListFilter struct {
	UploaderID string
	OwnerKind  string
	OwnerID    string
	Role       string
	Kind       string
	Status     string
	Access     string
	Search     string
	Page       int
	PageSize   int
}
