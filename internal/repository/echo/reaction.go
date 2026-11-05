// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/echo"
	"gorm.io/gorm/clause"
)

// AddReaction is idempotent: the unique index on (echo_id, emoji, actor) makes
// a repeated insert a no-op instead of incrementing anything.
func (echoRepository *EchoRepository) AddReaction(ctx context.Context, echoID, emoji, actor string) error {
	return echoRepository.getDB(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.Reaction{EchoID: echoID, Emoji: emoji, Actor: actor}).Error
}

// RemoveReaction is idempotent: removing a reaction that does not exist is not
// an error.
func (echoRepository *EchoRepository) RemoveReaction(ctx context.Context, echoID, emoji, actor string) error {
	return echoRepository.getDB(ctx).
		Where("echo_id = ? AND emoji = ? AND actor = ?", echoID, emoji, actor).
		Delete(&model.Reaction{}).Error
}

// ListReactions returns the hydrated reaction summary for the given echoes in
// one grouped query. The result is keyed by echo id and already ordered by the
// reaction catalog, so callers only attach it to the echo DTO.
func (echoRepository *EchoRepository) ListReactions(
	ctx context.Context,
	echoIDs []string,
	actor string,
) (map[string][]model.ReactionCount, error) {
	result := make(map[string][]model.ReactionCount, len(echoIDs))
	if len(echoIDs) == 0 {
		return result, nil
	}

	type reactionRow struct {
		EchoID  string
		Emoji   string
		Total   int
		Reacted int
	}
	var rows []reactionRow
	if err := echoRepository.getDB(ctx).
		Model(&model.Reaction{}).
		Select("echo_id, emoji, COUNT(*) AS total, MAX(CASE WHEN actor = ? THEN 1 ELSE 0 END) AS reacted", actor).
		Where("echo_id IN ?", echoIDs).
		Group("echo_id, emoji").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	byEcho := make(map[string]map[string]reactionRow, len(echoIDs))
	for _, row := range rows {
		if byEcho[row.EchoID] == nil {
			byEcho[row.EchoID] = make(map[string]reactionRow, len(model.ReactionEmojis))
		}
		byEcho[row.EchoID][row.Emoji] = row
	}

	for _, echoID := range echoIDs {
		emojiRows := byEcho[echoID]
		if len(emojiRows) == 0 {
			continue
		}
		counts := make([]model.ReactionCount, 0, len(emojiRows))
		for _, emoji := range model.ReactionEmojis {
			row, ok := emojiRows[emoji.Key]
			if !ok {
				continue
			}
			counts = append(counts, model.ReactionCount{
				Key:     emoji.Key,
				Emoji:   emoji.Char,
				Count:   row.Total,
				Reacted: row.Reacted > 0,
			})
		}
		result[echoID] = counts
	}

	return result, nil
}
