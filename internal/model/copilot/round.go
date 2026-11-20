// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// Round belongs to a live SSE request. The export excludes these temporary
// rows so restoring a backup cannot revive an outstanding write confirmation.
type Round struct {
	ID      string `gorm:"primaryKey"`
	Owner   string `gorm:"not null"`
	Expires int64  `gorm:"not null;index"`
	Answer  []byte
}

func (Round) TableName() string { return "copilot_rounds" }
