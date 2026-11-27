// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cli

import (
	"context"
	"fmt"

	"github.com/lin-snow/ech0/internal/cache"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/kvstore"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

type capsuleRuntime struct {
	db      *gorm.DB
	kv      kvstore.Store
	cache   *cache.Cache
	storage *storage.Registry
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

	storages, err := openStorageRegistry(context.Background(), dbProvider)
	if err != nil {
		return nil, err
	}

	return &capsuleRuntime{
		db:      db,
		kv:      durableKV,
		cache:   appCache,
		storage: storages,
		tx:      transaction.NewGormTransactor(dbProvider),
	}, nil
}

// openStorageRegistry loads the storage table once for a CLI command. The
// table is the truth; configuration only provides the initial row.
func openStorageRegistry(ctx context.Context, dbProvider func() *gorm.DB) (*storage.Registry, error) {
	repo := fileRepository.NewFileRepository(dbProvider)
	rows, err := repo.ListStorages(ctx)
	if err != nil {
		return nil, fmt.Errorf("load storages: %w", err)
	}
	if len(rows) == 0 {
		rows = seedStorages()
		for i := range rows {
			if err := repo.SaveStorage(ctx, &rows[i]); err != nil {
				return nil, fmt.Errorf("seed storages: %w", err)
			}
		}
	}

	cfgs := make([]storage.Config, 0, len(rows))
	for i := range rows {
		cfgs = append(cfgs, storage.Config{
			ID:      rows[i].ID,
			Driver:  rows[i].Driver,
			Name:    rows[i].Name,
			Default: rows[i].IsDefault,
			Config:  rows[i].Config,
		})
	}

	reg := storage.NewRegistry()
	if err := reg.Reload(ctx, cfgs); err != nil {
		return nil, fmt.Errorf("open storages: %w", err)
	}
	return reg, nil
}

func seedStorages() []fileModel.StorageConfig {
	cfg := config.Config().Storage
	if cfg.ObjectEnabled && cfg.Endpoint != "" {
		return []fileModel.StorageConfig{{
			ID:        uuidUtil.NewV7(),
			Driver:    storage.DriverS3,
			Name:      "default",
			IsDefault: true,
			Config: map[string]string{
				"endpoint":       cfg.Endpoint,
				"access_key":     cfg.AccessKey,
				"secret_key":     cfg.SecretKey,
				"bucket":         cfg.BucketName,
				"region":         cfg.Region,
				"use_ssl":        fmt.Sprint(cfg.UseSSL),
				"use_path_style": fmt.Sprint(cfg.UsePathStyle),
				"path_prefix":    cfg.PathPrefix,
			},
		}}
	}
	return []fileModel.StorageConfig{{
		ID:        uuidUtil.NewV7(),
		Driver:    storage.DriverLocal,
		Name:      "default",
		IsDefault: true,
		Config:    map[string]string{"root": cfg.DataRoot},
	}}
}
