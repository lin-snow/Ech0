// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/lin-snow/ech0/internal/config"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/virefs"
)

const (
	s3SnapshotPrefix    = "snapshots/"
	s3SnapshotKeepCount = 3
)

func BuildS3FS(cfg config.StorageConfig) (virefs.FS, error) {
	if !cfg.ObjectEnabled {
		return nil, fmt.Errorf("object storage is not enabled")
	}

	var opts []virefs.ObjectOption
	if cfg.PathPrefix != "" {
		opts = append(opts, virefs.WithPrefix(strings.Trim(cfg.PathPrefix, "/")+"/"))
	}

	fs, err := virefs.NewObjectFSFromConfig(context.Background(), virefsS3Config(cfg), opts...)
	if err != nil {
		return nil, fmt.Errorf("create snapshot s3 fs: %w", err)
	}
	return fs, nil
}

func virefsS3Config(cfg config.StorageConfig) *virefs.S3Config {
	provider := mapS3Provider(cfg.Provider)
	region := resolveS3Region(cfg.Provider, cfg.Region)
	endpoint := normalizeS3Endpoint(cfg.Endpoint, cfg.UseSSL)
	s3cfg := &virefs.S3Config{
		Provider:  provider,
		Endpoint:  endpoint,
		Region:    region,
		Bucket:    cfg.BucketName,
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
	}
	if cfg.UsePathStyle {
		s3cfg.UsePathStyle = aws.Bool(true)
	}
	return s3cfg
}

func UploadToS3(ctx context.Context, snapshotFilePath, fileName string, cfg config.StorageConfig) error {
	f, err := os.Open(snapshotFilePath)
	if err != nil {
		return fmt.Errorf("open snapshot file for s3 upload: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close snapshot file after s3 upload",
				slog.String("path", snapshotFilePath), logUtil.Err(closeErr))
		}
	}()

	s3FS, err := BuildS3FS(cfg)
	if err != nil {
		return err
	}

	s3Key := s3SnapshotPrefix + fileName
	if err := s3FS.Put(ctx, s3Key, f); err != nil {
		return fmt.Errorf("put snapshot to s3: %w", err)
	}

	logUtil.GetLogger().Info("Snapshot uploaded to S3",
		slog.String("key", s3Key))

	if err := cleanupOldS3Snapshots(ctx, s3FS, s3SnapshotKeepCount); err != nil {
		logUtil.GetLogger().Warn("Failed to cleanup old S3 snapshots",
			logUtil.Err(err))
	}

	return nil
}

func cleanupOldS3Snapshots(ctx context.Context, s3FS virefs.FS, keepCount int) error {
	result, err := s3FS.List(ctx, s3SnapshotPrefix)
	if err != nil {
		return fmt.Errorf("list s3 snapshots: %w", err)
	}

	var snapshotFiles []string
	for _, item := range result.Files {
		key := strings.Trim(item.Key, "/")
		if item.IsDir || key == "" {
			continue
		}
		if strings.HasPrefix(key, s3SnapshotPrefix) && strings.HasSuffix(key, ".zip") {
			snapshotFiles = append(snapshotFiles, key)
		}
	}

	if len(snapshotFiles) <= keepCount {
		return nil
	}

	sort.Strings(snapshotFiles)

	toDelete := snapshotFiles[:len(snapshotFiles)-keepCount]
	for _, key := range toDelete {
		if err := s3FS.Delete(ctx, key); err != nil {
			logUtil.GetLogger().Warn("Failed to delete old S3 snapshot",
				slog.String("key", key), logUtil.Err(err))
		} else {
			logUtil.GetLogger().Info("Deleted old S3 snapshot",
				slog.String("key", key))
		}
	}

	return nil
}

func mapS3Provider(raw string) virefs.Provider {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "minio":
		return virefs.ProviderMinIO
	case "r2":
		return virefs.ProviderR2
	default:
		return virefs.ProviderAWS
	}
}

func resolveS3Region(providerRaw string, regionRaw string) string {
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

func normalizeS3Endpoint(endpoint string, useSSL bool) string {
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
