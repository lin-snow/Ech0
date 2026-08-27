// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

const (
	USER_NOT_EXISTS_ID = ""
)

type User struct {
	ID       string `gorm:"type:char(36);primaryKey" json:"id"`
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Email    string `gorm:"size:255;index"            json:"email"`
	IsAdmin  bool   `gorm:"bool"                     json:"is_admin"`
	IsOwner  bool   `gorm:"bool"                     json:"is_owner"`
	Avatar   string `gorm:"size:255"                 json:"avatar"`
	Locale   string `gorm:"size:16;default:zh-CN"    json:"locale"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuidUtil.NewV7()
	}
	return nil
}
