// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package app

import "context"

type Component interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Namer interface {
	Name() string
}
