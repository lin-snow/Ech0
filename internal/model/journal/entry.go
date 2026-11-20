// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type Entry struct {
	ID      int64  `gorm:"primaryKey;autoIncrement"`
	Source  string `gorm:"not null;uniqueIndex:idx_journal_seq"`
	Seq     uint64 `gorm:"not null;uniqueIndex:idx_journal_seq"`
	Created int64  `gorm:"not null;index"`
	Level   string `gorm:"not null;index"`
	Search  string `gorm:"not null"`
	Raw     string `gorm:"not null"`
}

func (Entry) TableName() string { return "ech0_journal" }
