// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package export

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/pkg/virefs"
	vizip "github.com/lin-snow/ech0/pkg/virefs/plugin/zip"
	"gorm.io/gorm"
)

type Deps struct {
	DB       *gorm.DB
	Selector *storage.StorageSelector
	KV       kvstore.Store
}

type Options struct {
	Output         string
	IncludePrivate bool
	Zip            bool
	Generator      string
}

type Result struct {
	Path                              string
	Echoes, Files, Comments, Connects int
	SkippedPrivate                    int
	ExternalFiles                     int
}

func (d Deps) validate() error {
	switch {
	case d.DB == nil:
		return errors.New("capsule export: database is required")
	case d.Selector == nil:
		return errors.New("capsule export: storage selector is required")
	case d.KV == nil:
		return errors.New("capsule export: kv store is required")
	}
	return nil
}

func Run(ctx context.Context, deps Deps, opts Options) (*Result, error) {
	if err := deps.validate(); err != nil {
		return nil, err
	}
	output := strings.TrimSpace(opts.Output)
	if output == "" {
		return nil, errors.New("capsule export: output path is empty")
	}

	data, err := collect(ctx, deps, opts)
	if err != nil {
		return nil, err
	}

	if !opts.Zip {
		if err := ensureEmptyDir(output); err != nil {
			return nil, err
		}
	}

	parent := filepath.Dir(output)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("capsule export: create output parent %q: %w", parent, err)
	}
	stageDir, err := os.MkdirTemp(parent, ".ech0-capsule-*")
	if err != nil {
		return nil, fmt.Errorf("capsule export: create staging dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(stageDir) }()

	stage, err := virefs.NewLocalFS(stageDir, virefs.WithCreateRoot(), virefs.WithAtomicWrite())
	if err != nil {
		return nil, fmt.Errorf("capsule export: open staging dir %q: %w", stageDir, err)
	}

	keys, err := writeCapsule(ctx, deps, stage, data, opts)
	if err != nil {
		return nil, err
	}

	result := &Result{
		Path:           output,
		Echoes:         len(data.echoes),
		Files:          len(data.files),
		Comments:       len(data.comments),
		Connects:       len(data.connects),
		SkippedPrivate: data.skippedPrivate,
		ExternalFiles:  data.externalFiles,
	}

	if opts.Zip {
		path, err := packZip(ctx, stage, keys, output)
		if err != nil {
			return nil, err
		}
		result.Path = path
		return result, nil
	}

	if err := os.Remove(output); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("capsule export: clear output %q: %w", output, err)
	}
	if err := os.Rename(stageDir, output); err != nil {
		return nil, fmt.Errorf("capsule export: move capsule into place: %w", err)
	}
	return result, nil
}

func ensureEmptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil
	case err != nil:
		return fmt.Errorf("capsule export: inspect output %q: %w", dir, err)
	case len(entries) > 0:
		return fmt.Errorf("capsule export: output %q already exists and is not empty", dir)
	}
	return nil
}

func packZip(ctx context.Context, stage virefs.FS, keys []string, output string) (string, error) {
	out := output
	if !strings.EqualFold(filepath.Ext(out), ".zip") {
		out += ".zip"
	}
	if dir := filepath.Dir(out); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("capsule export: create output dir %q: %w", dir, err)
		}
	}

	f, err := os.Create(out)
	if err != nil {
		return "", fmt.Errorf("capsule export: create %q: %w", out, err)
	}
	if err := vizip.Pack(ctx, stage, keys, f); err != nil {
		_ = f.Close()
		_ = os.Remove(out)
		return "", fmt.Errorf("capsule export: pack %q: %w", out, err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("capsule export: close %q: %w", out, err)
	}
	return out, nil
}
