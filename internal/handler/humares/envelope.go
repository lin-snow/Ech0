// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"context"

	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

type Envelope[T any] struct {
	Body commonModel.Result[T]
}

func localizeResult[T any](ctx context.Context, body commonModel.Result[T]) commonModel.Result[T] {
	if body.MessageKey == "" {
		body.MessageKey = commonModel.MessageKeyFromMessage(body.Message)
	}
	if body.MessageKey != "" {
		body.Message = i18nUtil.Localize(localizerFrom(ctx), body.MessageKey, body.Message, body.MessageParams)
	}
	return body
}
