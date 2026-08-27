// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package zip

import (
	"archive/zip"
	"context"
	"io"

	virefs "github.com/lin-snow/ech0/pkg/virefs"
)

func Unpack(ctx context.Context, r io.ReaderAt, size int64, dst virefs.FS, prefix string, opts ...virefs.PutOption) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	cleanedPrefix, err := virefs.CleanKey(prefix)
	if err != nil {
		return err
	}

	for _, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			continue
		}

		name, err := virefs.CleanKey(f.Name)
		if err != nil || name == "" {
			continue
		}

		dstKey := name
		if cleanedPrefix != "" {
			dstKey = cleanedPrefix + "/" + name
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		putErr := dst.Put(ctx, dstKey, rc, opts...)
		rc.Close()
		if putErr != nil {
			return putErr
		}
	}
	return nil
}
