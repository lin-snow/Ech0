// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// RevokedToken contains the JWT identifier only. Expiry has Unix-second precision.
type RevokedToken struct {
	JTI       string `gorm:"primaryKey"`
	ExpiresAt int64  `gorm:"not null;index"`
}
