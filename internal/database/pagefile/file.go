// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package pagefile lets a snapshot worker open a native SQLite image a block at
// a time. Each worker owns a private overlay. The caller publishes changed blocks
// and the source cursor together after SQLite closes the only connection.
package pagefile

import (
	"errors"
	"io"
)

const BlockSize = 64 << 10

type File struct {
	Size int64
	// BaseSize bounds reads from the last published version.
	BaseSize int64
	Dirty    map[int64][]byte
	load     func(int64) ([]byte, error)
}

func New(size int64, load func(int64) ([]byte, error)) *File {
	return &File{Size: size, BaseSize: size, Dirty: make(map[int64][]byte), load: load}
}

func (f *File) block(index int64) ([]byte, error) {
	if data, ok := f.Dirty[index]; ok {
		return data, nil
	}
	if index*BlockSize >= f.BaseSize {
		return make([]byte, BlockSize), nil
	}
	data, err := f.load(index)
	if err != nil {
		return nil, err
	}
	if len(data) != BlockSize {
		return nil, errors.New("invalid snapshot block")
	}
	return data, nil
}

func (f *File) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, errors.New("negative snapshot offset")
	}
	n := 0
	for n < len(p) && off < f.Size {
		block, err := f.block(off / BlockSize)
		if err != nil {
			return n, err
		}
		count := min(int64(len(p)-n), BlockSize-off%BlockSize, f.Size-off)
		copy(p[n:], block[off%BlockSize:off%BlockSize+count])
		off += count
		n += int(count)
	}
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *File) WriteAt(p []byte, off int64) (int, error) {
	if off < 0 {
		return 0, errors.New("negative snapshot offset")
	}
	n := 0
	for n < len(p) {
		index := off / BlockSize
		data, ok := f.Dirty[index]
		if !ok {
			block, err := f.block(index)
			if err != nil {
				return n, err
			}
			data = append([]byte(nil), block...)
			f.Dirty[index] = data
		}
		wrote := copy(data[off%BlockSize:], p[n:])
		n += wrote
		off += int64(wrote)
		f.Size = max(f.Size, off)
	}
	return n, nil
}

func (f *File) Truncate(size int64) error {
	if size < 0 {
		return errors.New("negative snapshot size")
	}
	if size < f.Size {
		if tail := size % BlockSize; tail != 0 {
			block, err := f.block(size / BlockSize)
			if err != nil {
				return err
			}
			data := append([]byte(nil), block...)
			clear(data[tail:])
			f.Dirty[size/BlockSize] = data
		}
		for index := range f.Dirty {
			if index*BlockSize >= size {
				delete(f.Dirty, index)
			}
		}
		f.BaseSize = min(f.BaseSize, size)
	}
	f.Size = size
	return nil
}
