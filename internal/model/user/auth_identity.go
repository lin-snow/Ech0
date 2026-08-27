// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type UserLocalAuth struct {
	UserID       string `gorm:"type:char(36);primaryKey"`
	PasswordHash string `gorm:"size:255;not null"`
	PasswordAlgo string `gorm:"size:32;not null;default:md5"`
	UpdatedAt    int64  `gorm:"autoUpdateTime"`
}

func (UserLocalAuth) TableName() string {
	return "user_local_auth"
}

type UserExternalIdentity struct {
	ID        string `gorm:"type:char(36);primaryKey"`
	UserID    string `gorm:"type:char(36);not null;index"`
	Provider  string `gorm:"size:64;not null;index:idx_identity_provider_subject,priority:1"`
	Subject   string `gorm:"size:255;not null;index:idx_identity_provider_subject,priority:3"`
	Issuer    string `gorm:"size:255;default:'';index:idx_identity_provider_subject,priority:2"`
	Protocol  string `gorm:"size:32;not null;default:oauth2"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}

func (UserExternalIdentity) TableName() string {
	return "user_external_identities"
}

func (e *UserExternalIdentity) BeforeCreate(_ *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuidUtil.NewV7()
	}
	return nil
}

type WebAuthnCredential struct {
	ID             string `gorm:"type:char(36);primaryKey"`
	UserID         string `gorm:"type:char(36);not null;index"`
	CredentialID   string `gorm:"size:255;not null;uniqueIndex"`
	CredentialJSON string `gorm:"type:text;not null"`
	PublicKey      string `gorm:"type:text"`
	SignCount      uint32 `gorm:"not null;default:0"`
	LastUsedAt     int64  `gorm:"index"`
	DeviceName     string `gorm:"size:128"`
	AAGUID         string `gorm:"size:64"`
	CreatedAt      int64  `gorm:"autoCreateTime"`
	UpdatedAt      int64  `gorm:"autoUpdateTime"`
}

func (WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

func (w *WebAuthnCredential) BeforeCreate(_ *gorm.DB) error {
	if w.ID == "" {
		w.ID = uuidUtil.NewV7()
	}
	return nil
}
