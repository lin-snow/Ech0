// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package storetest is the conformance suite every storage.Store driver must
// pass. A driver is only "done" when these tests are green for it: the suite,
// not the driver's own tests, is the contract.
package storetest

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"

	"github.com/lin-snow/ech0/internal/storage"
)

// Run executes the suite against a fresh store.
func Run(t *testing.T, open func(t *testing.T) storage.Store) {
	t.Helper()

	t.Run("roundtrip", func(t *testing.T) {
		store := open(t)
		ctx := context.Background()
		body := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR not really a png")
		key := "2026/11/0192f0c4-0000-7000-8000-000000000001.png"

		got, err := store.Put(ctx, key, bytes.NewReader(body), int64(len(body)))
		if err != nil {
			t.Fatalf("put: %v", err)
		}
		sum := sha256.Sum256(body)
		if got.Size != int64(len(body)) || got.SHA256 != hex.EncodeToString(sum[:]) {
			t.Fatalf("put result = %+v, want size %d", got, len(body))
		}

		rc, info, err := store.Open(ctx, key)
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		defer func() { _ = rc.Close() }()
		read, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if !bytes.Equal(read, body) {
			t.Fatalf("body mismatch: %q != %q", read, body)
		}
		if info.Size != int64(len(body)) {
			t.Fatalf("info.Size = %d, want %d", info.Size, len(body))
		}
	})

	t.Run("size mismatch leaves nothing behind", func(t *testing.T) {
		store := open(t)
		ctx := context.Background()
		key := "2026/11/0192f0c4-0000-7000-8000-000000000002.bin"

		if _, err := store.Put(ctx, key, bytes.NewReader([]byte("short")), 50); !errors.Is(err, storage.ErrSizeMismatch) {
			t.Fatalf("put err = %v, want ErrSizeMismatch", err)
		}
		if _, err := store.Stat(ctx, key); !errors.Is(err, storage.ErrNotFound) {
			t.Fatalf("stat after failed put = %v, want ErrNotFound", err)
		}
	})

	t.Run("missing object", func(t *testing.T) {
		store := open(t)
		if _, _, err := store.Open(context.Background(), "2026/11/nope"); !errors.Is(err, storage.ErrNotFound) {
			t.Fatalf("open err = %v, want ErrNotFound", err)
		}
	})

	t.Run("delete is idempotent", func(t *testing.T) {
		store := open(t)
		ctx := context.Background()
		key := "2026/11/0192f0c4-0000-7000-8000-000000000003.bin"
		if _, err := store.Put(ctx, key, bytes.NewReader([]byte("x")), 1); err != nil {
			t.Fatalf("put: %v", err)
		}
		if err := store.Delete(ctx, key); err != nil {
			t.Fatalf("first delete: %v", err)
		}
		if err := store.Delete(ctx, key); err != nil {
			t.Fatalf("second delete: %v", err)
		}
	})

	t.Run("large body streams intact", func(t *testing.T) {
		store := open(t)
		ctx := context.Background()
		key := "2026/11/0192f0c4-0000-7000-8000-000000000004.bin"
		body := make([]byte, 1<<20)
		if _, err := rand.Read(body); err != nil {
			t.Fatalf("rand: %v", err)
		}
		want := sha256.Sum256(body)
		if _, err := store.Put(ctx, key, bytes.NewReader(body), int64(len(body))); err != nil {
			t.Fatalf("put: %v", err)
		}
		rc, _, err := store.Open(ctx, key)
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		defer func() { _ = rc.Close() }()
		hash := sha256.New()
		n, err := io.Copy(hash, rc)
		if err != nil {
			t.Fatalf("copy: %v", err)
		}
		if n != int64(len(body)) || hex.EncodeToString(hash.Sum(nil)) != hex.EncodeToString(want[:]) {
			t.Fatalf("large body mismatch: n=%d", n)
		}
	})
}
