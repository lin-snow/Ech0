// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"testing"

	"github.com/lin-snow/ech0/internal/config"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
)

func TestStorageConfigFromSetting(t *testing.T) {
	defaultCfg := config.StorageConfig{
		DataRoot:  "data/files",
		Endpoint:  "env.example.com",
		AccessKey: "env-ak-must-be-ignored",
		SecretKey: "env-sk-must-be-ignored",
	}
	s3 := settingModel.S3Setting{
		Enable:       true,
		Provider:     "aws",
		Endpoint:     "https://s3.amazonaws.com",
		BucketName:   "db-bucket",
		Region:       "us-east-1",
		UseSSL:       true,
		CDNURL:       "https://cdn.example.com/",
		PathPrefix:   "uploads/",
		UsePathStyle: true,
	}

	cfg := storageConfigFromSetting(s3, defaultCfg)

	if !cfg.ObjectEnabled {
		t.Fatalf("expected object storage enabled")
	}
	if cfg.Provider != "aws" {
		t.Fatalf("expected provider aws, got %s", cfg.Provider)
	}
	if cfg.Endpoint != "s3.amazonaws.com" {
		t.Fatalf("expected scheme-stripped endpoint, got %s", cfg.Endpoint)
	}
	if cfg.BucketName != "db-bucket" {
		t.Fatalf("expected db bucket, got %s", cfg.BucketName)
	}
	if cfg.CDNURL != "https://cdn.example.com" {
		t.Fatalf("expected trailing-slash-trimmed CDN, got %s", cfg.CDNURL)
	}
	if cfg.PathPrefix != "uploads" {
		t.Fatalf("expected trimmed path prefix, got %s", cfg.PathPrefix)
	}
	if cfg.DataRoot != "data/files" {
		t.Fatalf("expected DataRoot carried from defaultCfg, got %s", cfg.DataRoot)
	}
	if !cfg.UsePathStyle {
		t.Fatalf("expected UsePathStyle carried from setting")
	}
	if cfg.AccessKey != "" || cfg.SecretKey != "" {
		t.Fatalf("S3 fields must come from setting only (no env fallback), got ak=%q sk=%q", cfg.AccessKey, cfg.SecretKey)
	}
}

func TestStorageManager_ApplyS3Setting_InvalidConfigKeepsOldSelector(t *testing.T) {
	m := NewStorageManager(nil)
	before := m.GetSelector()
	if before == nil {
		t.Fatalf("expected initial selector")
	}

	err := m.ApplyS3Setting(settingModel.S3Setting{
		Enable:     true,
		Provider:   "minio",
		Endpoint:   "127.0.0.1:9000",
		AccessKey:  "",
		SecretKey:  "",
		BucketName: "",
		Region:     "",
		UseSSL:     false,
	})
	if err == nil {
		t.Fatalf("expected apply to fail for invalid object config")
	}

	after := m.GetSelector()
	if after != before {
		t.Fatalf("expected selector pointer unchanged after failed apply")
	}
	if after.ObjectEnabled() {
		t.Fatalf("expected object storage disabled after failed apply")
	}
}
