// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/migrator/artifact"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/virefs"
	vizip "github.com/lin-snow/ech0/pkg/virefs/plugin/zip"
)

var ErrNoSnapshot = artifact.ErrNone

const (
	dataDir              = artifact.DataDir
	snapshotRelativeDir  = artifact.SnapshotDir
	tmpRelativeDir       = artifact.TmpDir
	capsuleRelativeDir   = artifact.CapsuleDir
	dbFileName           = "ech0.db"
	dbStagingRelativeDir = artifact.TmpDir + "/db-export"
)

func Slot() artifact.Slot {
	return artifact.Snapshots()
}

type CreateOption func(*createConfig)

type createConfig struct {
	dbCopy func(dstPath string) error
}

func WithConsistentDB(copyFn func(dstPath string) error) CreateOption {
	return func(cfg *createConfig) {
		cfg.dbCopy = copyFn
	}
}

func Create(opts ...CreateOption) (string, string, error) {
	var cfg createConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	slot := Slot()
	fileName := slot.Name(time.Now().UTC())
	snapshotDir := slot.Dir()
	snapshotPath := slot.Path(fileName)
	tempPath := slot.Path("." + fileName + ".tmp")

	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create snapshot dir: %w", err)
	}

	dataFS, err := virefs.NewLocalFS(dataDir)
	if err != nil {
		return "", "", fmt.Errorf("open data dir: %w", err)
	}

	packFS := virefs.FS(dataFS)
	if cfg.dbCopy != nil {
		stageFS, cleanup, stageErr := stageConsistentDB(cfg.dbCopy)
		if stageErr != nil {
			return "", "", stageErr
		}
		defer cleanup()
		packFS = &dbOverlayFS{FS: dataFS, stage: stageFS}
	}

	ctx := context.Background()
	var keys []string
	if err := virefs.Walk(ctx, dataFS, "", func(key string, info virefs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir {
			return nil
		}
		cleanKey := strings.Trim(strings.TrimSpace(key), "/")
		if cleanKey == "" {
			return nil
		}
		if shouldExcludeFromSnapshot(cleanKey) {
			return nil
		}
		if cfg.dbCopy != nil && isDBArtifact(cleanKey) {
			return nil
		}
		keys = append(keys, cleanKey)
		return nil
	}); err != nil {
		return "", "", fmt.Errorf("walk data dir: %w", err)
	}
	if cfg.dbCopy != nil {
		keys = append(keys, dbFileName)
	}

	f, err := os.Create(tempPath)
	if err != nil {
		return "", "", fmt.Errorf("create zip file: %w", err)
	}

	if err := vizip.Pack(ctx, packFS, keys, f); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close snapshot zip after pack error",
				slog.String("path", tempPath), slog.String("error", closeErr.Error()))
		}
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("pack zip: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("close zip file: %w", err)
	}

	if err := os.Rename(tempPath, snapshotPath); err != nil {
		_ = os.Remove(tempPath)
		return "", "", fmt.Errorf("finalize snapshot zip: %w", err)
	}

	if err := slot.KeepOnly(fileName); err != nil {
		return "", "", err
	}

	return snapshotPath, fileName, nil
}

func LatestPath() (string, error) {
	return Slot().Latest()
}

func Unpack(zipPath, destDir string) error {
	f, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close snapshot zip reader",
				slog.String("path", zipPath), slog.String("error", closeErr.Error()))
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat zip: %w", err)
	}

	dstFS, err := virefs.NewLocalFS(destDir, virefs.WithCreateRoot())
	if err != nil {
		return fmt.Errorf("open dest dir: %w", err)
	}

	return vizip.Unpack(context.Background(), f, info.Size(), dstFS, "")
}

func stageConsistentDB(copyFn func(dstPath string) error) (virefs.FS, func(), error) {
	stagingDir := filepath.Join(dataDir, dbStagingRelativeDir)
	if err := os.RemoveAll(stagingDir); err != nil {
		return nil, nil, fmt.Errorf("clean db staging dir: %w", err)
	}
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create db staging dir: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(stagingDir) }

	if err := copyFn(filepath.Join(stagingDir, dbFileName)); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("write consistent db copy: %w", err)
	}

	stageFS, err := virefs.NewLocalFS(stagingDir)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("open db staging dir: %w", err)
	}
	return stageFS, cleanup, nil
}

func isDBArtifact(cleanKey string) bool {
	switch cleanKey {
	case dbFileName, dbFileName + "-wal", dbFileName + "-shm", dbFileName + "-journal":
		return true
	}
	return false
}

type dbOverlayFS struct {
	virefs.FS
	stage virefs.FS
}

func (o *dbOverlayFS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == dbFileName {
		return o.stage.Get(ctx, dbFileName)
	}
	return o.FS.Get(ctx, key)
}

func (o *dbOverlayFS) Stat(ctx context.Context, key string) (*virefs.FileInfo, error) {
	if key == dbFileName {
		return o.stage.Stat(ctx, dbFileName)
	}
	return o.FS.Stat(ctx, key)
}

func shouldExcludeFromSnapshot(cleanKey string) bool {
	for _, dir := range artifact.Excluded() {
		prefix := strings.Trim(strings.TrimSpace(dir), "/")
		if cleanKey == prefix || strings.HasPrefix(cleanKey, prefix+"/") {
			return true
		}
	}
	return false
}
