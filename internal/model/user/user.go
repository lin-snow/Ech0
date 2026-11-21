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

// Account status is binary on purpose: there is no soft-delete state, so a
// user is either allowed to act or disabled until an administrator restores
// them. Permanent deletion removes the row entirely.
type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusDisabled Status = "DISABLED"
)

func (s Status) Valid() bool {
	return s == StatusActive || s == StatusDisabled
}

type User struct {
	ID       string `gorm:"type:char(36);primaryKey" json:"id"`
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Email    string `gorm:"size:255;index"            json:"email"`
	IsAdmin  bool   `gorm:"bool"                     json:"is_admin"`
	Status   Status `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	Avatar   string `gorm:"size:255"                 json:"avatar"`
	Locale   string `gorm:"size:16;default:zh-CN"    json:"locale"`
}

func (u *User) IsActive() bool {
	return u.Status == "" || u.Status == StatusActive
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuidUtil.NewV7()
	}
	if u.Status == "" {
		u.Status = StatusActive
	}
	return nil
}
