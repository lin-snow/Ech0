// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package visitor

type DailyStat struct {
	Date string `gorm:"type:char(10);primaryKey" json:"date"`
	PV   int64  `gorm:"default:0" json:"pv"`
	UV   int64  `gorm:"default:0" json:"uv"`
}

type VisitorSeen struct {
	Hash   string `gorm:"primaryKey"`
	Day    string `gorm:"not null;index"`
	LastPV int64  `gorm:"not null"`
}
