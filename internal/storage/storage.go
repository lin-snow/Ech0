// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import "strings"

type (
	Category    string
	StorageType string
)

const (
	CategoryImage    Category = "image"
	CategoryVideo    Category = "video"
	CategoryAudio    Category = "audio"
	CategoryPDF      Category = "pdf"
	CategoryMarkdown Category = "markdown"
	CategoryFile     Category = "file"
)

const (
	StorageTypeLocal    StorageType = "local"
	StorageTypeObject   StorageType = "object"
	StorageTypeExternal StorageType = "external"
)

func NormalizeCategory(raw string) Category {
	switch Category(strings.ToLower(strings.TrimSpace(raw))) {
	case CategoryImage:
		return CategoryImage
	case CategoryVideo:
		return CategoryVideo
	case CategoryAudio:
		return CategoryAudio
	case CategoryPDF:
		return CategoryPDF
	case CategoryMarkdown:
		return CategoryMarkdown
	default:
		return CategoryFile
	}
}

func (c Category) IsImageLike() bool {
	return c == CategoryImage
}

func NormalizeStorageType(raw string) StorageType {
	switch StorageType(strings.ToLower(strings.TrimSpace(raw))) {
	case "s3":
		return StorageTypeObject
	case StorageTypeObject:
		return StorageTypeObject
	case StorageTypeExternal:
		return StorageTypeExternal
	default:
		return StorageTypeLocal
	}
}

type URLResolver func(key string) string

type KeyGenerator interface {
	GenerateKey(category Category, userID string, originalFilename string) (string, error)
}

func TrimLeadingSlash(p string) string {
	return strings.TrimPrefix(p, "/")
}
