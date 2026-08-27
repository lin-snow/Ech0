// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type PageQueryDto struct {
	Page     int    `json:"page"     form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Search   string `json:"search"   form:"search"`
}

type EchoQueryDto struct {
	Page      int      `json:"page"`
	PageSize  int      `json:"pageSize"`
	Search    string   `json:"search"`
	TagIDs    []string `json:"tagIds"`
	SortBy    string   `json:"sortBy"`
	SortOrder string   `json:"sortOrder"`
	DateFrom  int64    `json:"dateFrom"`
	DateTo    int64    `json:"dateTo"`
	Private   *bool    `json:"private,omitempty"`
	UserID    string   `json:"-"`
}

type FileDto struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Key         string `json:"key"`
	StorageType string `json:"storage_type,omitempty"`
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Category    string `json:"category,omitempty"`
	Size        int64  `json:"size,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

type FileDeleteDto struct {
	ID string `json:"id" binding:"required"`
}

type PresignDto struct {
	ID          string `json:"id"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Key         string `json:"key"`
	PresignURL  string `json:"presign_url"`
	FileURL     string `json:"file_url"`
}

type GetPresignURLDto struct {
	FileName    string `json:"file_name" binding:"required"`
	ContentType string `json:"content_type"`
	StorageType string `json:"storage_type,omitempty"`
}

type CreateExternalFileDto struct {
	URL         string `json:"url" binding:"required"`
	ContentType string `json:"content_type"`
	Category    string `json:"category"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Name        string `json:"name"`
}

type UpdateFileMetaDto struct {
	Size        int64  `json:"size" binding:"required,min=0"`
	Width       *int   `json:"width,omitempty"`
	Height      *int   `json:"height,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type FileListQueryDto struct {
	Page        int    `json:"page" form:"page"`
	PageSize    int    `json:"pageSize" form:"pageSize"`
	Search      string `json:"search" form:"search"`
	StorageType string `json:"storage_type" form:"storage_type"`
}

type FileListItemDto struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	StorageType string `json:"storage_type"`
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

type FileListResultDto struct {
	Total int64             `json:"total"`
	Items []FileListItemDto `json:"items"`
}

type FileTreeQueryDto struct {
	StorageType string `json:"storage_type" form:"storage_type" binding:"required"`
	Prefix      string `json:"prefix" form:"prefix"`
}

type FilePathStreamQueryDto struct {
	StorageType string `json:"storage_type" form:"storage_type" binding:"required"`
	Path        string `json:"path" form:"path" binding:"required"`
	Name        string `json:"name" form:"name"`
	ContentType string `json:"content_type" form:"content_type"`
}

type FileTreeNodeDto struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	NodeType    string `json:"node_type"`
	HasChildren bool   `json:"has_children"`
	FileID      string `json:"file_id,omitempty"`
	Size        int64  `json:"size,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	ModifiedAt  int64  `json:"modified_at,omitempty"`
}

type FileTreeResultDto struct {
	Items []FileTreeNodeDto `json:"items"`
}

type GetWebsiteTitleDto struct {
	WebSiteURL string `json:"website_url" form:"website_url" binding:"required"`
}
