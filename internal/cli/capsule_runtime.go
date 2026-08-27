// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cli

import (
	"fmt"

	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/kvstore"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type capsuleRuntime struct {
	db      *gorm.DB
	kv      kvstore.Store
	cache   cache.ICache[string, any]
	storage *storage.Manager
	tx      transaction.Transactor
}

func newCapsuleRuntime() (rt *capsuleRuntime, err error) {
	defer func() {
		if r := recover(); r != nil {
			rt, err = nil, fmt.Errorf("initialise runtime: %v", r)
		}
	}()

	dbProvider := database.ProvideDBProvider()
	db := dbProvider()
	if db == nil {
		return nil, fmt.Errorf("database is not available")
	}

	appCache, err := cache.ProvideCache()
	if err != nil {
		return nil, fmt.Errorf("initialise cache: %w", err)
	}

	durableKV := kvstore.NewPersistent(keyvalueRepository.NewKeyValueRepository(dbProvider, appCache))

	return &capsuleRuntime{
		db:      db,
		kv:      durableKV,
		cache:   appCache,
		storage: storage.NewStorageManager(durableKV),
		tx:      transaction.NewGormTransactor(dbProvider),
	}, nil
}

func (rt *capsuleRuntime) selector() *storage.StorageSelector {
	return rt.storage.GetSelector()
}
