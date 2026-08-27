// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/pkg/busen"
)

func AsyncParallel() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.WithParallelism(ec.AgentParallelism),
		busen.WithBuffer(ec.AgentBuffer),
		busen.WithOverflow(MapOverflow(ec.DefaultOverflow)),
	}
}

func AsyncSequential() []busen.SubscribeOption {
	ec := config.Config().Event
	return []busen.SubscribeOption{
		busen.Async(),
		busen.Sequential(),
		busen.WithBuffer(ec.SystemBuffer),
		busen.WithOverflow(MapOverflow(ec.DefaultOverflow)),
	}
}
