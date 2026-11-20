// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"archive/zip"
	"bytes"
	"hash/crc32"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// sparseArchive lets archive/zip exercise ZIP64 offsets without allocating or
// writing a multi-gigabyte fixture.
type sparseArchive struct {
	header, footer []byte
	size           int64
}

func (a sparseArchive) ReadAt(p []byte, off int64) (int, error) {
	n := 0
	for len(p) > n {
		pos := off + int64(n)
		switch {
		case pos < int64(len(a.header)):
			n += copy(p[n:], a.header[pos:])
		case pos < int64(len(a.header))+a.size:
			count := min(len(p)-n, int(int64(len(a.header))+a.size-pos))
			clear(p[n : n+count])
			n += count
		case pos < int64(len(a.header)+len(a.footer))+a.size:
			n += copy(p[n:], a.footer[pos-int64(len(a.header))-a.size:])
		default:
			return n, io.EOF
		}
	}
	return n, nil
}

func TestIncrementalZIP(t *testing.T) {
	data := []byte("native SQLite image")
	contents := append(snapshotHeader(uint64(len(data))), data...)
	contents = append(contents, snapshotFooter(uint64(len(data)), crc32.ChecksumIEEE(data))...)
	archive, err := zip.NewReader(bytes.NewReader(contents), int64(len(contents)))
	require.NoError(t, err)
	require.Len(t, archive.File, 1)
	f, err := archive.File[0].Open()
	require.NoError(t, err)
	actual, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	require.Equal(t, data, actual)
	n := uint64(math.MaxUint32 + 1000)
	sparse := sparseArchive{snapshotHeader(n), snapshotFooter(n, 0), int64(n)}
	archive, err = zip.NewReader(sparse, int64(len(sparse.header)+len(sparse.footer))+sparse.size)
	require.NoError(t, err)
	require.Len(t, archive.File, 1)
	require.Equal(t, zipEntry, archive.File[0].Name)
	require.Equal(t, n, archive.File[0].UncompressedSize64)
}
