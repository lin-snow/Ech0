// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"errors"
	"path"
)

var ErrSkipDir = errors.New("skip this directory")

type WalkFunc func(key string, info FileInfo, err error) error

func Walk(ctx context.Context, fsys FS, prefix string, fn WalkFunc) error {
	result, err := fsys.List(ctx, prefix)
	if err != nil {
		return fn(prefix, FileInfo{}, err)
	}
	for _, fi := range result.Files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if fi.IsDir {
			if err := fn(fi.Key, fi, nil); err != nil {
				if errors.Is(err, ErrSkipDir) {
					continue
				}
				return err
			}
			subPrefix := fi.Key
			if prefix != "" && !hasPathPrefix(fi.Key, prefix) {
				subPrefix = path.Join(prefix, fi.Key)
			}
			if err := Walk(ctx, fsys, subPrefix, fn); err != nil {
				return err
			}
		} else {
			if err := fn(fi.Key, fi, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

func hasPathPrefix(key, prefix string) bool {
	if prefix == "" {
		return true
	}
	return len(key) > len(prefix) && key[:len(prefix)] == prefix && key[len(prefix)] == '/'
}
