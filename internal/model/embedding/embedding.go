// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
)

type EchoEmbedding struct {
	EchoID      string `gorm:"type:char(36);primaryKey" json:"echo_id"`
	ContentHash string `gorm:"type:varchar(64);index"   json:"content_hash"`
	Model       string `gorm:"type:varchar(100)"        json:"model"`
	Dim         int    `gorm:"default:0"                json:"dim"`
	Content     string `gorm:"type:text"                json:"content"`
	Username    string `gorm:"type:varchar(100)"        json:"username"`
	EchoCreated int64  `gorm:"index"                    json:"echo_created"`
	CreatedAt   int64  `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime"           json:"updated_at"`
}

func (EchoEmbedding) TableName() string { return "echo_embeddings" }

type SearchResult struct {
	EchoID      string                   `json:"echo_id"`
	Content     string                   `json:"content"`
	Username    string                   `json:"username"`
	EchoCreated int64                    `json:"echo_created"`
	Distance    float64                  `json:"distance"`
	Files       []fileModel.File         `json:"files,omitempty"`
	Extension   *echoModel.EchoExtension `json:"extension,omitempty"`
}

type IndexState struct {
	Model string `json:"model"`
	Dim   int    `json:"dim"`
}
