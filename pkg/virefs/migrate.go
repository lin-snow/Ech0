// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"fmt"
	"path"
	"strings"
)

type ConflictPolicy int

const (
	ConflictError ConflictPolicy = iota
	ConflictSkip
	ConflictOverwrite
)

type MigrateProgress struct {
	Key     string
	Copied  int
	Skipped int
	Total   int
}

type MigrateResult struct {
	Copied  int
	Skipped int
	Total   int
}

type MigrateOption func(*migrateConfig)

type migrateConfig struct {
	conflict ConflictPolicy
	dryRun   bool
	progress func(MigrateProgress)
	keyFunc  func(srcKey string) string
}

func WithConflictPolicy(p ConflictPolicy) MigrateOption {
	return func(c *migrateConfig) { c.conflict = p }
}

func WithDryRun() MigrateOption {
	return func(c *migrateConfig) { c.dryRun = true }
}

func WithProgressFunc(fn func(MigrateProgress)) MigrateOption {
	return func(c *migrateConfig) { c.progress = fn }
}

func WithMigrateKeyFunc(fn func(srcKey string) string) MigrateOption {
	return func(c *migrateConfig) { c.keyFunc = fn }
}

func Migrate(ctx context.Context, src FS, srcPrefix string, dst FS, dstPrefix string, opts ...MigrateOption) (*MigrateResult, error) {
	cfg := &migrateConfig{}
	for _, o := range opts {
		o(cfg)
	}

	result := &MigrateResult{}

	err := Walk(ctx, src, srcPrefix, func(key string, info FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir {
			return nil
		}

		result.Total++

		relKey := stripPrefix(key, srcPrefix)
		if cfg.keyFunc != nil {
			relKey = cfg.keyFunc(relKey)
		}

		var dstKey string
		if dstPrefix != "" {
			dstKey = path.Join(dstPrefix, relKey)
		} else {
			dstKey = relKey
		}

		if cfg.conflict != ConflictOverwrite {
			exists, err := dst.Exists(ctx, dstKey)
			if err != nil {
				return fmt.Errorf("migrate: check exists %q: %w", dstKey, err)
			}
			if exists {
				switch cfg.conflict {
				case ConflictError:
					return fmt.Errorf("migrate: %w: destination key %q already exists", ErrAlreadyExist, dstKey)
				case ConflictSkip:
					result.Skipped++
					if cfg.progress != nil {
						cfg.progress(MigrateProgress{
							Key: key, Copied: result.Copied,
							Skipped: result.Skipped, Total: result.Total,
						})
					}
					return nil
				}
			}
		}

		if !cfg.dryRun {
			if err := Copy(ctx, src, key, dst, dstKey); err != nil {
				return fmt.Errorf("migrate: copy %q -> %q: %w", key, dstKey, err)
			}
		}

		result.Copied++
		if cfg.progress != nil {
			cfg.progress(MigrateProgress{
				Key: key, Copied: result.Copied,
				Skipped: result.Skipped, Total: result.Total,
			})
		}
		return nil
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func stripPrefix(key, prefix string) string {
	if prefix == "" {
		return key
	}
	trimmed := strings.TrimPrefix(key, prefix)
	trimmed = strings.TrimPrefix(trimmed, "/")
	return trimmed
}
