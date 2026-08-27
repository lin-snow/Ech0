// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package zip

import (
	"archive/zip"
	"context"
	"io"

	virefs "github.com/lin-snow/ech0/pkg/virefs"
)

func Pack(ctx context.Context, fsys virefs.FS, keys []string, w io.Writer) (retErr error) {
	zw := zip.NewWriter(w)
	defer func() {
		if cerr := zw.Close(); retErr == nil {
			retErr = cerr
		}
	}()

	for _, raw := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}

		key, err := virefs.CleanKey(raw)
		if err != nil {
			return err
		}

		header := &zip.FileHeader{
			Name:   key,
			Method: zip.Deflate,
		}

		if info, err := fsys.Stat(ctx, key); err == nil {
			header.UncompressedSize64 = uint64(info.Size)
			header.Modified = info.LastModified
		}

		ew, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		rc, err := fsys.Get(ctx, key)
		if err != nil {
			return err
		}

		_, copyErr := io.Copy(ew, rc)
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}
