// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package job

import (
	"context"
	"encoding/json"
	"fmt"
)

type TypedRun[P any] func(ctx context.Context, p P, report ReportFunc) (any, error)

func Adapt[P any](fn TypedRun[P]) Runner {
	return runnerFunc(func(ctx context.Context, raw []byte, report ReportFunc) (any, error) {
		var p P
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &p); err != nil {
				return nil, fmt.Errorf("decode %T payload: %w", p, err)
			}
		}
		return fn(ctx, p, report)
	})
}

type runnerFunc func(ctx context.Context, payload []byte, report ReportFunc) (any, error)

func (f runnerFunc) Run(ctx context.Context, payload []byte, report ReportFunc) (any, error) {
	return f(ctx, payload, report)
}
