// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import (
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

// Reaction records one actor's reaction with one emoji on one echo. The
// unique index is the entire idempotency and anti-spam mechanism: writes use
// INSERT ... ON CONFLICT DO NOTHING and deletes are unconditional.
type Reaction struct {
	ID        string `gorm:"type:char(36);primaryKey"                                            json:"-"`
	EchoID    string `gorm:"type:char(36);not null;uniqueIndex:uk_echo_emoji_actor,priority:1"    json:"echo_id"`
	Emoji     string `gorm:"type:varchar(32);not null;uniqueIndex:uk_echo_emoji_actor,priority:2" json:"emoji"`
	Actor     string `gorm:"type:varchar(80);not null;uniqueIndex:uk_echo_emoji_actor,priority:3" json:"-"`
	CreatedAt int64  `gorm:"autoCreateTime"                                                     json:"created_at"`
}

func (r *Reaction) BeforeCreate(_ *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuidUtil.NewV7()
	}
	return nil
}

// TableName keeps the reactions scoped to echos and matches the raw SQL used
// for aggregate queries and hot ranking.
func (r *Reaction) TableName() string { return "echo_reactions" }

// ReactionEmoji is one selectable reaction. Clients store and send Key; Char
// is only for rendering, so glyph variation never leaks into storage.
type ReactionEmoji struct {
	Key  string `json:"key"`
	Char string `json:"emoji"`
}

// ReactionEmojis is the fixed reaction catalog, in display order. The Huma
// enum on the reaction routes must list the same keys; reaction_test.go
// enforces that they stay in sync.
var ReactionEmojis = []ReactionEmoji{
	{Key: "like", Char: "👍"},
	{Key: "heart", Char: "❤️"},
	{Key: "haha", Char: "😂"},
	{Key: "wow", Char: "😮"},
	{Key: "sad", Char: "😢"},
	{Key: "party", Char: "🎉"},
}

// ReactionEmojiByKey reports whether key is a supported reaction.
func ReactionEmojiByKey(key string) (ReactionEmoji, bool) {
	for _, emoji := range ReactionEmojis {
		if emoji.Key == key {
			return emoji, true
		}
	}
	return ReactionEmoji{}, false
}

// ReactionCount is the aggregate the API returns for one emoji on one echo.
type ReactionCount struct {
	Key     string `json:"key"`
	Emoji   string `json:"emoji"`
	Count   int    `json:"count"`
	Reacted bool   `json:"reacted"`
}
