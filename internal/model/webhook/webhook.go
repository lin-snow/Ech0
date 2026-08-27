// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type Webhook struct {
	ID          string `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string `                    json:"name"`
	URL         string `                    json:"url"`
	Secret      string `                    json:"-"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	LastStatus  string `                    json:"last_status"`
	LastTrigger int64  `                    json:"last_trigger"`
	CreatedAt   int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (w *Webhook) BeforeCreate(_ *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuidUtil.NewV7()
	}
	return nil
}
