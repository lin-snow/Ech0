// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/kvstore"
)

type (
	KVStore  = kvstore.Store
	AppCache = *cache.Cache
)
