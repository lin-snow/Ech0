// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package snapshot 是 Ech0 统一的快照资源:把整个 data/ 目录打成 zip 归档(导出端),
// 以及把外部上传的 zip 解开到目录(导入端)。它是 import / export 两端共用的同一产物,
// 只依赖 virefs 与配置,可被无 DI 的 CLI 直接调用。
//
// 快照格式即「data/ 的 zip」(排除 snapshots/、capsules/、tmp/)。数据库文件名固定为 ech0.db
// (见 config 默认),导入端在 ech0 importer 中定位。在线导出时应通过 WithConsistentDB
// 打入数据库的一致性副本(见其注释),冷目录打包则原样带走全部文件。
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

// ErrNoSnapshot 表示快照目录下尚无可下载的快照（需先执行一次导出）。与 artifact.ErrNone 是
// 同一个值，errors.Is 在两侧都成立；保留此名是为了不惊动既有调用方。
var ErrNoSnapshot = artifact.ErrNone

const (
	dataDir = artifact.DataDir
	// 布局常量的本地投影：真相在 artifact，这里只为可读性与测试引用。
	snapshotRelativeDir = artifact.SnapshotDir
	tmpRelativeDir      = artifact.TmpDir
	capsuleRelativeDir  = artifact.CapsuleDir
	// dbFileName 是快照内数据库文件的固定名称(与 config 默认的 data/ech0.db 对应)。
	dbFileName = "ech0.db"
	// dbStagingRelativeDir 是数据库一致性副本的暂存目录,位于 tmp 下(tmp 本身被快照排除)。
	dbStagingRelativeDir = artifact.TmpDir + "/db-export"
)

// Slot 返回快照产物槽位，供下载出口取回最新一份。
func Slot() artifact.Slot {
	return artifact.Snapshots()
}

// CreateOption 调整 Create 的打包行为。
type CreateOption func(*createConfig)

type createConfig struct {
	dbCopy func(dstPath string) error
}

// WithConsistentDB 注册「把数据库一致性副本写到指定路径」的函数(线上即 database.SnapshotTo)。
// 设置后,Create 不再直接拷贝运行中的 ech0.db 及其伴生文件(-wal/-shm/-journal)——带着并发
// 写入直接拷实时库文件可能得到撕裂状态——而是把该副本以 ech0.db 之名打进 zip。
// 未设置时保持原样打包:冷目录场景下 db 与伴生文件一起原样带走本就是一致的。
func WithConsistentDB(copyFn func(dstPath string) error) CreateOption {
	return func(cfg *createConfig) {
		cfg.dbCopy = copyFn
	}
}

// Create packs the data/ directory into a zip snapshot using VireFS and returns
// the path and file name of the produced archive. It excludes the snapshots/ and
// tmp/ subtrees and keeps only the latest snapshot locally.
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
		// 实时库文件已从 walk 中排除,此处以固定名打入一致性副本(由 dbOverlayFS 路由)。
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

// LatestPath 返回最新一份快照 zip 的路径，无可用快照时返回 ErrNoSnapshot。
// 供同步下载出口取回「上一次导出作业产出的快照」。
func LatestPath() (string, error) {
	return Slot().Latest()
}

// Unpack unpacks a snapshot zip file to the destination directory.
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

// stageConsistentDB 让 copyFn 把数据库一致性副本写到 tmp 下的暂存目录,返回以该目录
// 为根的 FS(供打包时顶替实时库文件)和清理函数。副本写不出来时导出必须失败——
// 静默回退去拷实时库文件会产出一份可能缺数据的「坏快照」。
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

// isDBArtifact 识别运行中数据库的主文件及其伴生文件(-wal/-shm/-journal)。
// 仅在启用一致性副本时排除;冷目录打包时它们本就该原样带走。
func isDBArtifact(cleanKey string) bool {
	switch cleanKey {
	case dbFileName, dbFileName + "-wal", dbFileName + "-shm", dbFileName + "-journal":
		return true
	}
	return false
}

// dbOverlayFS 在打包视图中把数据库文件的读取路由到暂存的一致性副本,其余 key 透传底层
// data FS。vizip.Pack 只使用 Get/Stat,故仅覆盖这两个方法。
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
