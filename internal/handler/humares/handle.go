// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package humares

import (
	"context"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

func Wrap[I, T any](h func(context.Context, *I) (commonModel.Result[T], error)) func(context.Context, *I) (*Envelope[T], error) {
	return func(ctx context.Context, in *I) (*Envelope[T], error) {
		res, err := h(ctx, in)
		if err != nil {
			return nil, Err(ctx, err)
		}
		return &Envelope[T]{Body: localizeResult(ctx, res)}, nil
	}
}
