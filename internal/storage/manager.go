// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/kvstore"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
)

type Manager struct {
	mu         sync.RWMutex
	defaultCfg config.StorageConfig
	durableKV  kvstore.Store
	selector   *StorageSelector
}

func NewStorageManager(durableKV kvstore.Store) *Manager {
	defaultCfg := config.Config().Storage
	m := &Manager{
		defaultCfg: defaultCfg,
		durableKV:  durableKV,
		selector:   NewStorageSelector(defaultCfg),
	}
	_ = m.ReloadFromConfigAndDB(context.Background())
	fileModel.RegisterURLResolver(m.ResolveURL)
	return m
}

func NewStorageManagerForTest(dataRoot string) *Manager {
	cfg := config.StorageConfig{DataRoot: dataRoot}
	return &Manager{
		defaultCfg: cfg,
		durableKV:  nil,
		selector:   NewStorageSelector(cfg),
	}
}

func (m *Manager) GetSelector() *StorageSelector {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.selector
}

func (m *Manager) ResolveURL(storageType, key string) string {
	return m.GetSelector().ResolveURL(StorageType(storageType), key)
}

func (m *Manager) GetStorageConfig(ctx context.Context) config.StorageConfig {
	return m.resolveStorageConfig(ctx)
}

func (m *Manager) ReloadFromConfigAndDB(ctx context.Context) error {
	return m.replaceSelector(m.resolveStorageConfig(ctx))
}

func (m *Manager) ApplyS3Setting(setting settingModel.S3Setting) error {
	return m.replaceSelector(storageConfigFromSetting(setting, m.defaultCfg))
}

func (m *Manager) replaceSelector(cfg config.StorageConfig) error {
	selector := NewStorageSelector(cfg)
	if cfg.ObjectEnabled && !selector.ObjectEnabled() {
		return errors.New("object storage enabled but initialization failed")
	}
	m.mu.Lock()
	m.selector = selector
	m.mu.Unlock()
	return nil
}

func (m *Manager) resolveStorageConfig(ctx context.Context) config.StorageConfig {
	return storageConfigFromSetting(m.currentS3Setting(ctx), m.defaultCfg)
}

func (m *Manager) currentS3Setting(ctx context.Context) settingModel.S3Setting {
	if m.durableKV == nil {
		return coreSetting.S3.Default()
	}
	s3, _ := coreSetting.Get(ctx, m.durableKV, coreSetting.S3)
	return s3
}

func storageConfigFromSetting(s3 settingModel.S3Setting, defaultCfg config.StorageConfig) config.StorageConfig {
	cfg := defaultCfg
	cfg.DataRoot = strings.TrimSpace(defaultCfg.DataRoot)
	if cfg.DataRoot == "" {
		cfg.DataRoot = "data/files"
	}

	cfg.ObjectEnabled = s3.Enable
	cfg.Provider = strings.TrimSpace(s3.Provider)
	cfg.Endpoint = trimEndpoint(s3.Endpoint)
	cfg.AccessKey = strings.TrimSpace(s3.AccessKey)
	cfg.SecretKey = strings.TrimSpace(s3.SecretKey)
	cfg.BucketName = strings.TrimSpace(s3.BucketName)
	cfg.Region = strings.TrimSpace(s3.Region)
	cfg.CDNURL = strings.TrimRight(strings.TrimSpace(s3.CDNURL), "/")
	cfg.PathPrefix = strings.Trim(strings.TrimSpace(s3.PathPrefix), "/")
	cfg.UseSSL = s3.UseSSL
	cfg.UsePathStyle = s3.UsePathStyle

	return cfg
}

func trimEndpoint(endpoint string) string {
	e := strings.TrimSpace(endpoint)
	e = strings.TrimPrefix(e, "http://")
	e = strings.TrimPrefix(e, "https://")
	return e
}
