// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/role"
	"gorm.io/gorm"
)

const (
	USER_NOT_EXISTS_ID = ""
)

// Role is the user's standing in the space. The type lives in pkg/role so the
// request principal can carry it without importing the domain model.
type Role = role.Role

const (
	RoleOwner  = role.Owner
	RoleAdmin  = role.Admin
	RoleMember = role.Member
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

// AccountState is the live authorization snapshot of one account. It is read
// on every authenticated request so that a demotion, an ownership transfer or
// a disable applies to the next request rather than to the next token refresh.
type AccountState struct {
	Found  bool
	Role   Role
	Status Status
}

func (s AccountState) Active() bool {
	return s.Found && (s.Status == "" || s.Status == StatusActive)
}

type User struct {
	ID           string `gorm:"type:char(36);primaryKey" json:"id"`
	Username     string `gorm:"size:255;not null;unique" json:"username"`
	Email        string `gorm:"size:255;index"            json:"email"`
	Role         Role   `gorm:"size:16;not null;default:member;index" json:"role"`
	Status       Status `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	Avatar       string `gorm:"size:255"                 json:"avatar"`
	AvatarFileID string `gorm:"type:char(36);index"       json:"avatar_file_id,omitempty"`
	Locale       string `gorm:"size:16;default:zh-CN"    json:"locale"`
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
	if u.Role == "" {
		u.Role = RoleMember
	}
	return nil
}
