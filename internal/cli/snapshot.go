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
		tuiUtil.CLIBoxHeader{Icon: "🗄️", Title: "Snapshot", Value: path},
		tuiUtil.CLIInfoItem{Title: "Size", Msg: strconv.FormatInt(outcome.Size, 10) + " bytes"},
	)
	return nil
}

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

	folder := "ech0_" + uuidUtil.NewV7()
	relativeTmpDir := filepath.ToSlash(filepath.Join(migrator.TmpRelativeDir, folder))
	extractDir := filepath.Join("data", relativeTmpDir)
	if err := os.MkdirAll(extractDir, 0o755); err != nil {
		return fmt.Errorf("create extract dir: %w", err)
	}
	if err := snapshot.Unpack(path, extractDir); err != nil {
		_ = os.RemoveAll(extractDir)
		return fmt.Errorf("unpack snapshot: %w", err)
	}

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
