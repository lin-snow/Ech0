// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"
)

// MemStore is the in-memory driver. It exists so the conformance suite and the
// file-service tests run without touching a disk, and so a deployment can be
// started with zero durable storage in tests. It implements the same contract
// as the durable drivers, not a convenience subset.
type MemStore struct {
	mu      sync.RWMutex
	objects map[string]memObject
}

type memObject struct {
	data        []byte
	contentType string
	modTime     time.Time
}

var (
	_ Store       = (*MemStore)(nil)
	_ RangeOpener = (*MemStore)(nil)
)

func NewMemStore() *MemStore {
	return &MemStore{objects: map[string]memObject{}}
}

func (m *MemStore) Put(ctx context.Context, key string, r io.Reader, size int64) (PutResult, error) {
	if err := ctx.Err(); err != nil {
		return PutResult{}, err
	}
	if size < 0 {
		return PutResult{}, ErrSizeMismatch
	}
	data, err := io.ReadAll(io.LimitReader(r, size+1))
	if err != nil {
		return PutResult{}, err
	}
	if int64(len(data)) != size {
		return PutResult{}, ErrSizeMismatch
	}
	sum := sha256.Sum256(data)
	prefix := data
	if len(prefix) > 512 {
		prefix = prefix[:512]
	}
	contentType := ContentTypeOf(prefix)

	m.mu.Lock()
	m.objects[key] = memObject{data: data, contentType: contentType, modTime: time.Now().UTC()}
	m.mu.Unlock()

	return PutResult{Size: size, SHA256: hex.EncodeToString(sum[:]), ContentType: contentType}, nil
}

func (m *MemStore) Open(ctx context.Context, key string) (io.ReadCloser, BlobInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, BlobInfo{}, err
	}
	m.mu.RLock()
	obj, ok := m.objects[key]
	m.mu.RUnlock()
	if !ok {
		return nil, BlobInfo{}, ErrNotFound
	}
	info := BlobInfo{Size: int64(len(obj.data)), ContentType: obj.contentType, ModTime: obj.modTime}
	return io.NopCloser(bytes.NewReader(obj.data)), info, nil
}

// OpenRange serves a byte window without copying the object, which is what the
// gateway needs for a Range request.
func (m *MemStore) OpenRange(ctx context.Context, key string, offset, length int64) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	obj, ok := m.objects[key]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	start := offset
	if start < 0 {
		start = 0
	}
	if start > int64(len(obj.data)) {
		start = int64(len(obj.data))
	}
	end := start + length
	if length < 0 || end > int64(len(obj.data)) {
		end = int64(len(obj.data))
	}
	return io.NopCloser(bytes.NewReader(obj.data[start:end])), nil
}

func (m *MemStore) Stat(ctx context.Context, key string) (BlobInfo, error) {
	if err := ctx.Err(); err != nil {
		return BlobInfo{}, err
	}
	m.mu.RLock()
	obj, ok := m.objects[key]
	m.mu.RUnlock()
	if !ok {
		return BlobInfo{}, ErrNotFound
	}
	return BlobInfo{Size: int64(len(obj.data)), ContentType: obj.contentType, ModTime: obj.modTime}, nil
}

func (m *MemStore) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.objects, key)
	m.mu.Unlock()
	return nil
}

// errNoop keeps errors imported for the sentinel translation tests without
// widening the public surface.
var _ = errors.Is
