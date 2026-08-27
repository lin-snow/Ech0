// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type EchoFile = fileModel.EchoFile

type Echo struct {
	ID        string         `gorm:"type:char(36);primaryKey"                      json:"id"`
	Content   string         `gorm:"type:text;not null"                            json:"content"`
	Username  string         `gorm:"type:varchar(100)"                             json:"username,omitempty"`
	EchoFiles []EchoFile     `gorm:"foreignKey:EchoID;constraint:OnDelete:CASCADE" json:"echo_files,omitempty"`
	Layout    string         `gorm:"type:varchar(50);default:'waterfall'"          json:"layout,omitempty"`
	Private   bool           `gorm:"default:false;index:idx_echos_private_created,priority:1" json:"private"`
	UserID    string         `gorm:"type:char(36);not null;index"                  json:"user_id"`
	Extension *EchoExtension `gorm:"foreignKey:EchoID;constraint:OnDelete:CASCADE" json:"extension,omitempty"`
	Tags      []Tag          `gorm:"many2many:echo_tags;"                          json:"tags,omitempty"`
	FavCount  int            `gorm:"default:0"                                     json:"fav_count"`
	CreatedAt int64          `gorm:"autoCreateTime;index:idx_echos_private_created,priority:2" json:"created_at"`
}

type EchoExtension struct {
	ID        string         `gorm:"type:char(36);primaryKey"      json:"id"`
	EchoID    string         `gorm:"type:char(36);not null;uniqueIndex" json:"echo_id"`
	Type      string         `gorm:"type:varchar(100);not null"    json:"type"`
	Payload   map[string]any `gorm:"serializer:json;type:text;not null" json:"payload"`
	CreatedAt int64          `gorm:"autoCreateTime"                         json:"created_at"`
	UpdatedAt int64          `gorm:"autoUpdateTime"                         json:"updated_at"`
}

type Tag struct {
	ID         string `gorm:"type:char(36);primaryKey"              json:"id"`
	Name       string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	UsageCount int    `gorm:"default:0"                             json:"usage_count"`
	CreatedAt  int64  `gorm:"autoCreateTime"                        json:"created_at"`
}

type EchoTag struct {
	EchoID string `gorm:"type:char(36);primaryKey"`
	TagID  string `gorm:"type:char(36);primaryKey;index"`
}

func (e *Echo) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuidUtil.NewV7()
	}
	return nil
}

func (t *Tag) BeforeCreate(_ *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuidUtil.NewV7()
	}
	return nil
}

func (e *EchoExtension) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuidUtil.NewV7()
	}
	return nil
}

const (
	Extension_MUSIC      = "MUSIC"
	Extension_VIDEO      = "VIDEO"
	Extension_GITHUBPROJ = "GITHUBPROJ"
	Extension_WEBSITE    = "WEBSITE"
	Extension_LOCATION   = "LOCATION"
	Extension_TWEET      = "TWEET"

	LayoutWaterfall  = "waterfall"
	LayoutGrid       = "grid"
	LayoutHorizontal = "horizontal"
	LayoutCarousel   = "carousel"
	LayoutStack      = "stack"
	LayoutNone       = "none"
)
