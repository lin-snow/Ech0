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
	refresh    sync.Mutex
	mu         sync.RWMutex
	activeCfg  config.StorageConfig
	defaultCfg config.StorageConfig
	durableKV  kvstore.Store
	selector   *StorageSelector
}

func NewStorageManager(durableKV kvstore.Store) *Manager {
	defaultCfg := config.Config().Storage
	m := &Manager{
		defaultCfg: defaultCfg,
		activeCfg:  defaultCfg,
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
		activeCfg:  cfg,
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

func (m *Manager) GetStorageConfig(_ context.Context) config.StorageConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.activeCfg
}

// Serialize reads with replacements so a slower read cannot restore an
// obsolete credential after another request has already applied a rotation.
func (m *Manager) ReloadFromConfigAndDB(ctx context.Context) error {
	m.refresh.Lock()
	defer m.refresh.Unlock()
	setting := coreSetting.S3.Pristine()
	if m.durableKV != nil {
		var err error
		setting, err = coreSetting.S3.Get(ctx, m.durableKV)
		if err != nil {
			return err
		}
	}
	return m.replaceSelector(storageConfigFromSetting(setting, m.defaultCfg))
}

func (m *Manager) ApplyS3Setting(setting settingModel.S3Setting) error {
	m.refresh.Lock()
	defer m.refresh.Unlock()
	return m.replaceSelector(storageConfigFromSetting(setting, m.defaultCfg))
}

func (m *Manager) replaceSelector(cfg config.StorageConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cfg == m.activeCfg && m.selector != nil && (!cfg.ObjectEnabled || m.selector.ObjectEnabled()) {
		return nil
	}
	selector := NewStorageSelector(cfg)
	if cfg.ObjectEnabled && !selector.ObjectEnabled() {
		return errors.New("object storage enabled but initialization failed")
	}
	m.selector = selector
	m.activeCfg = cfg
	return nil
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
