// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"context"
	"log/slog"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/kvstore"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/virefs"
)

func ProvideStorageManager(durableKV kvstore.Store) *Manager { return NewStorageManager(durableKV) }

var (
	ManagerSet  = wire.NewSet(ProvideStorageManager)
	ProviderSet = wire.NewSet(ManagerSet)
)

func NewFS(cfg config.StorageConfig) virefs.FS {
	schema := NewFileSchema()
	if cfg.ObjectEnabled {
		return buildS3FS(cfg, schema)
	}
	return buildLocalFS(cfg, schema)
}

func NewURLResolver(cfg config.StorageConfig) URLResolver {
	schema := NewFileSchema()
	if cfg.ObjectEnabled {
		return buildS3URLResolver(cfg, schema)
	}
	return buildLocalURLResolver(schema)
}

func buildLocalFS(cfg config.StorageConfig, schema *virefs.Schema) virefs.FS {
	root := cfg.DataRoot
	if root == "" {
		root = "data/files"
	}
	fs, err := virefs.NewLocalFS(root,
		virefs.WithCreateRoot(),
		virefs.WithAtomicWrite(),
		virefs.WithLocalKeyFunc(schema.Resolve),
	)
	if err != nil {
		logUtil.Warn("create local fs failed, fallback to defaults", slog.String("module", "storage"), logUtil.Err(err))
		fs, _ = virefs.NewLocalFS("data/files",
			virefs.WithCreateRoot(),
			virefs.WithAtomicWrite(),
			virefs.WithLocalKeyFunc(schema.Resolve),
		)
	}
	return fs
}

func buildLocalURLResolver(schema *virefs.Schema) URLResolver {
	pathResolver := buildLocalPathURLResolver()
	return func(key string) string {
		return pathResolver(schema.Resolve(key))
	}
}

func buildLocalPathURLResolver() URLResolver {
	return func(path string) string {
		clean := strings.Trim(strings.TrimSpace(path), "/")
		if clean == "" {
			return "/api/files"
		}
		return "/api/files/" + clean
	}
}

func buildS3FS(cfg config.StorageConfig, schema *virefs.Schema) virefs.FS {
	var opts []virefs.ObjectOption
	if cfg.PathPrefix != "" {
		opts = append(opts, virefs.WithPrefix(strings.Trim(cfg.PathPrefix, "/")+"/"))
	}
	opts = append(opts, virefs.WithObjectKeyFunc(schema.Resolve))

	fs, err := virefs.NewObjectFSFromConfig(context.Background(), virefsS3ConfigFromStorage(cfg), opts...)
	if err != nil {
		logUtil.Warn("create s3 fs failed, fallback to local", slog.String("module", "storage"), logUtil.Err(err))
		return buildLocalFS(cfg, schema)
	}
	return fs
}

func virefsS3ConfigFromStorage(cfg config.StorageConfig) *virefs.S3Config {
	s3cfg := &virefs.S3Config{
		Provider:  mapProvider(cfg.Provider),
		Endpoint:  normalizeEndpoint(cfg.Endpoint, cfg.UseSSL),
		Region:    resolveObjectRegion(cfg.Provider, cfg.Region),
		Bucket:    cfg.BucketName,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}
	if cfg.UsePathStyle {
		s3cfg.UsePathStyle = aws.Bool(true)
	}
	return s3cfg
}

func buildS3URLResolver(cfg config.StorageConfig, schema *virefs.Schema) URLResolver {
	pathResolver := buildS3PathURLResolver(cfg)
	return func(key string) string {
		return pathResolver(schema.Resolve(key))
	}
}

func buildS3PathURLResolver(cfg config.StorageConfig) URLResolver {
	prefix := ""
	if cfg.PathPrefix != "" {
		prefix = strings.Trim(cfg.PathPrefix, "/") + "/"
	}

	cdnURL := strings.TrimSpace(cfg.CDNURL)
	if cdnURL != "" {
		if !strings.HasPrefix(strings.ToLower(cdnURL), "http://") &&
			!strings.HasPrefix(strings.ToLower(cdnURL), "https://") {
			protocol := "http"
			if cfg.UseSSL {
				protocol = "https"
			}
			cdnURL = protocol + "://" + cdnURL
		}
		cdnURL = strings.TrimRight(cdnURL, "/")
		return func(path string) string {
			clean := strings.Trim(strings.TrimSpace(path), "/")
			return cdnURL + "/" + prefix + clean
		}
	}

	endpoint := normalizeEndpoint(cfg.Endpoint, cfg.UseSSL)
	baseURL := strings.TrimRight(endpoint, "/") + "/" + cfg.BucketName
	if !addressesPathStyle(cfg) {
		if vh, ok := virtualHostedBaseURL(endpoint, cfg.BucketName); ok {
			baseURL = vh
		}
	}
	return func(path string) string {
		clean := strings.Trim(strings.TrimSpace(path), "/")
		return baseURL + "/" + prefix + clean
	}
}

func addressesPathStyle(cfg config.StorageConfig) bool {
	if cfg.UsePathStyle {
		return true
	}
	switch mapProvider(cfg.Provider) {
	case virefs.ProviderMinIO, virefs.ProviderR2:
		return true
	default:
		return false
	}
}

func virtualHostedBaseURL(endpoint, bucket string) (string, bool) {
	if endpoint == "" || bucket == "" {
		return "", false
	}
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return "", false
	}
	u.Host = bucket + "." + u.Host
	return strings.TrimRight(u.String(), "/"), true
}

func normalizeEndpoint(endpoint string, useSSL bool) string {
	if endpoint == "" {
		return endpoint
	}
	lower := strings.ToLower(endpoint)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return endpoint
	}
	if useSSL {
		return "https://" + endpoint
	}
	return "http://" + endpoint
}

func mapProvider(raw string) virefs.Provider {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "minio":
		return virefs.ProviderMinIO
	case "r2":
		return virefs.ProviderR2
	default:
		return virefs.ProviderAWS
	}
}

func resolveObjectRegion(providerRaw string, regionRaw string) string {
	region := strings.TrimSpace(regionRaw)
	if region != "" {
		return region
	}
	switch strings.ToLower(strings.TrimSpace(providerRaw)) {
	case "r2", "other":
		return "auto"
	default:
		return "us-east-1"
	}
}
