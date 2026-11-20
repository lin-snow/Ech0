// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/config"
)

func ProvideCache() (*Cache, error) {
	backend, err := NewRistrettoCache[string, any](1000000, 1000000, 100)
	if err != nil {
		return nil, err
	}
	c := New(backend)
	c.sharedDatabase = config.Config().Database.Type == "libsql"
	return c, nil
}

var ProviderSet = wire.NewSet(ProvideCache)
