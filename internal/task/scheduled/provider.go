// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package scheduled

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewCleanup,
	NewSnapshot,
	NewVisitorSnapshot,
)
