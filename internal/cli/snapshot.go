// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lin-snow/ech0/internal/migrator"
	"github.com/lin-snow/ech0/internal/migrator/snapshot"
	migratorModel "github.com/lin-snow/ech0/internal/model/migrator"
	tuiUtil "github.com/lin-snow/ech0/internal/util/tui"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

// 快照与胶囊是两种不同的产物：快照是整库 + data 目录的 zip 备份（破坏性整库替换、
// 不可读、不跨实现），胶囊是可读可编辑的内容交换格式。二者共用动词但语义不同，
// 因此格式是子命令而不是 flag 值（spec §9 / design Q11）。

// DoExportSnapshot 产出一份整库快照 zip。
//
// 复用 Web 端导出作业的同一个引擎：配置了对象存储时会额外后台上传，
// 本地产物始终落在 data/files/snapshots/ 下。
func DoExportSnapshot(output string) error {
	rt, err := newCapsuleRuntime()
	if err != nil {
		return err
	}

	outcome, err := migrator.NewExportEngine(rt.storage).Export(
		context.Background(),
		func(phase string, _ any) {
			if phase != "" {
				fmt.Fprintf(os.Stderr, "… %s\n", phase)
			}
		},
	)
	if err != nil {
		return err
	}

	path := outcome.ArtifactPath
	if strings.TrimSpace(output) != "" {
		if path, err = copyArtifact(outcome.ArtifactPath, output); err != nil {
			return err
		}
	}

	tuiUtil.PrintCLIWithBox(
		tuiUtil.CLIInfoItem{Title: "🗄️  Snapshot", Msg: path},
		tuiUtil.CLIInfoItem{Title: "Size", Msg: strconv.FormatInt(outcome.Size, 10) + " bytes"},
	)
	return nil
}

// DoImportSnapshot 用一份快照替换当前实例的内容。
//
// 破坏性操作，必须显式 --yes。落库语义与 Web 端「全局迁移」完全一致（同一个引擎），
// 包括「已存在主键则跳过」的批量插入与迁移后配置应用。
func DoImportSnapshot(path string, yes bool) error {
	if !yes {
		return errors.New("importing a snapshot rewrites instance data; pass --yes to confirm")
	}
	if !strings.HasSuffix(strings.ToLower(path), ".zip") {
		return fmt.Errorf("expected a snapshot .zip, got %q", path)
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}

	rt, err := newCapsuleRuntime()
	if err != nil {
		return err
	}

	// 引擎读的是已解包目录，且 resolveTmpDir 只接受 data/files/tmp 之下的相对路径，
	// 故这里复刻 Web 上传通道的落点约定，而不是随便找个临时目录。
	folder := "ech0_" + uuidUtil.MustNewV7()
	relativeTmpDir := filepath.ToSlash(filepath.Join(migrator.TmpRelativeDir, folder))
	extractDir := filepath.Join("data", relativeTmpDir)
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return fmt.Errorf("create extract dir: %w", err)
	}
	if err := snapshot.Unpack(path, extractDir); err != nil {
		_ = os.RemoveAll(extractDir)
		return fmt.Errorf("unpack snapshot: %w", err)
	}

	// Import 自带 tmp 清理（CleanupTmpDirFromPayload）。
	engine := migrator.NewImportEngine(rt.kv, rt.storage, rt.cache)
	if _, err := engine.Import(
		context.Background(),
		migratorModel.MigrationPayload{
			SourceType:    migratorModel.MigrationSourceEch0,
			SourcePayload: map[string]any{"tmp_dir": relativeTmpDir},
		},
		func(phase string, _ any) {
			if phase != "" {
				fmt.Fprintf(os.Stderr, "… %s\n", phase)
			}
		},
	); err != nil {
		return err
	}

	tuiUtil.PrintCLIInfo("📥 Snapshot imported", path)
	return nil
}

// copyArtifact 把引擎产物复制到用户指定位置（引擎只认自己的 snapshots 目录）。
func copyArtifact(src, dst string) (string, error) {
	if !strings.HasSuffix(strings.ToLower(dst), ".zip") {
		dst += ".zip"
	}
	if dir := filepath.Dir(dst); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create output dir: %w", err)
		}
	}

	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	return dst, out.Sync()
}
