// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/echo"
	actorUtil "github.com/lin-snow/ech0/internal/util/actor"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// ReactToEcho adds one emoji reaction for the current actor. It is idempotent:
// reacting twice with the same emoji changes nothing.
func (echoService *EchoService) ReactToEcho(ctx context.Context, id, emoji string) ([]model.ReactionCount, error) {
	return echoService.setReaction(ctx, id, emoji, true)
}

// UnreactEcho removes one emoji reaction for the current actor. It is
// idempotent: removing a reaction that is not there is not an error.
func (echoService *EchoService) UnreactEcho(ctx context.Context, id, emoji string) ([]model.ReactionCount, error) {
	return echoService.setReaction(ctx, id, emoji, false)
}

// GetReactionCatalog returns the supported emoji in display order.
func (echoService *EchoService) GetReactionCatalog() []model.ReactionEmoji {
	return model.ReactionEmojis
}

func (echoService *EchoService) setReaction(
	ctx context.Context,
	id, emoji string,
	add bool,
) ([]model.ReactionCount, error) {
	if _, ok := model.ReactionEmojiByKey(emoji); !ok {
		return nil, errors.New(commonModel.INVALID_PARAMS)
	}

	echo, err := echoService.echoRepository.GetEchosById(ctx, id)
	if err != nil {
		return nil, err
	}
	if echo == nil {
		return nil, errors.New(commonModel.ECHO_NOT_FOUND)
	}
	if echo.Private {
		allowed, err := echoService.canReadPrivate(ctx)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
		}
	}

	actor := reactionActor(ctx)
	if add {
		err = echoService.echoRepository.AddReaction(ctx, id, emoji, actor)
	} else {
		err = echoService.echoRepository.RemoveReaction(ctx, id, emoji, actor)
	}
	if err != nil {
		return nil, err
	}

	counts, err := echoService.echoRepository.ListReactions(ctx, []string{id}, actor)
	if err != nil {
		return nil, err
	}
	if counts[id] == nil {
		return []model.ReactionCount{}, nil
	}
	return counts[id], nil
}

// reactionActor resolves who is reacting: an authenticated user always wins,
// then the request-scoped visitor identity, then a shared anonymous fallback
// for contexts without an HTTP request (background jobs, tests).
func reactionActor(ctx context.Context) string {
	if userID := viewer.MustFromContext(ctx).UserID(); userID != "" {
		return "u:" + userID
	}
	if fallback := actorUtil.FromContext(ctx); fallback != "" {
		return fallback
	}
	return "ip:unknown"
}

// hydrateReactions attaches the derived reaction summary to echoes returned by
// a read path. The cached echo structs are copied before mutation so the shared
// cache never learns any actor-specific state.
func (echoService *EchoService) hydrateReactions(
	ctx context.Context,
	echos []model.Echo,
) ([]model.Echo, error) {
	if len(echos) == 0 {
		return echos, nil
	}

	ids := make([]string, 0, len(echos))
	for i := range echos {
		if echos[i].ID != "" {
			ids = append(ids, echos[i].ID)
		}
	}
	byEcho, err := echoService.echoRepository.ListReactions(ctx, ids, reactionActor(ctx))
	if err != nil {
		return nil, err
	}

	out := make([]model.Echo, len(echos))
	for i := range echos {
		out[i] = echos[i]
		out[i].Reactions = byEcho[echos[i].ID]
	}
	return out, nil
}

// hydrateEchoReactions is the single-echo variant of hydrateReactions.
func (echoService *EchoService) hydrateEchoReactions(
	ctx context.Context,
	echo *model.Echo,
) (*model.Echo, error) {
	if echo == nil {
		return nil, nil
	}
	hydrated, err := echoService.hydrateReactions(ctx, []model.Echo{*echo})
	if err != nil {
		return nil, err
	}
	clone := hydrated[0]
	return &clone, nil
}
