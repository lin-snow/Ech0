// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package task

import (
	"context"

	"github.com/go-co-op/gocron/v2"
)

type Task interface {
	Name() string
	Schedule(ctx context.Context, s gocron.Scheduler) error
}

type StopHook interface {
	OnStop(ctx context.Context)
}
