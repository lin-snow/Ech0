// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/storage/storetest"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

func TestMemStoreConformance(t *testing.T) {
	storetest.Run(t, func(t *testing.T) storage.Store {
		t.Helper()
		return storage.NewMemStore()
	})
}

func TestLocalStoreConformance(t *testing.T) {
	storetest.Run(t, func(t *testing.T) storage.Store {
		t.Helper()
		store, err := storage.OpenDriver(context.Background(), storage.Config{
			ID:     "local",
			Driver: storage.DriverLocal,
			Config: map[string]string{"root": t.TempDir()},
		})
		if err != nil {
			t.Fatalf("open local driver: %v", err)
		}
		return store
	})
}

func TestLocalStoreOpenRange(t *testing.T) {
	store, err := storage.OpenDriver(context.Background(), storage.Config{
		ID:     "local",
		Driver: storage.DriverLocal,
		Config: map[string]string{"root": t.TempDir()},
	})
	if err != nil {
		t.Fatalf("open local driver: %v", err)
	}
	ctx := context.Background()
	key := "2026/11/0192f0c4-0000-7000-8000-000000000005.bin"
	if _, err := store.Put(ctx, key, strings.NewReader("0123456789abcdef"), 16); err != nil {
		t.Fatalf("put: %v", err)
	}
	ranged, ok := store.(storage.RangeOpener)
	if !ok {
		t.Fatal("local store must implement RangeOpener")
	}
	rc, err := ranged.OpenRange(ctx, key, 4, 6)
	if err != nil {
		t.Fatalf("open range: %v", err)
	}
	defer func() { _ = rc.Close() }()
	gotBytes, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read range: %v", err)
	}
	got := string(gotBytes)
	if got != "456789" {
		t.Fatalf("range = %q, want %q", got, "456789")
	}
}

func TestKeyFor(t *testing.T) {
	id := uuidUtil.NewV7()
	key, err := storage.KeyFor(id, ".PNG")
	if err != nil {
		t.Fatalf("KeyFor: %v", err)
	}
	if len(key) < len(id)+len(".png") || key[len(key)-4:] != ".png" {
		t.Fatalf("key = %q, want <yyyy>/<mm>/<id>.png", key)
	}
	if _, err := storage.KeyFor("0192f0c4-0000-4000-8000-000000000000", ""); !errors.Is(err, storage.ErrInvalidFileID) {
		t.Fatalf("v4 id err = %v, want ErrInvalidFileID", err)
	}
}

func TestRegistry(t *testing.T) {
	reg := storage.NewRegistry()
	if _, _, err := reg.Default(); !errors.Is(err, storage.ErrUnknownStorage) {
		t.Fatalf("empty registry err = %v, want ErrUnknownStorage", err)
	}

	err := reg.Reload(context.Background(), []storage.Config{
		{ID: "mem-a", Driver: storage.DriverMem},
		{ID: "mem-b", Driver: storage.DriverMem, Default: true},
	})
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	id, _, err := reg.Default()
	if err != nil || id != "mem-b" {
		t.Fatalf("default = %q, %v; want mem-b", id, err)
	}
	if _, err := reg.For("missing"); !errors.Is(err, storage.ErrUnknownStorage) {
		t.Fatalf("unknown store err = %v", err)
	}

	// A failed reload must leave the previous map serving.
	if err := reg.Reload(context.Background(), []storage.Config{{ID: "bad", Driver: "nope"}}); err == nil {
		t.Fatal("reload with unknown driver must fail")
	}
	if _, err := reg.For("mem-a"); err != nil {
		t.Fatalf("previous registry lost after failed reload: %v", err)
	}

	// A CDN origin is per storage: public bytes go there, and a failed reload
	// must not resurrect an origin that is no longer configured.
	if got := reg.PublicURL("mem-a", "2026/11/x.png"); got != "" {
		t.Fatalf("PublicURL without cdn = %q, want empty", got)
	}
	if err := reg.Reload(context.Background(), []storage.Config{
		{ID: "mem-a", Driver: storage.DriverMem, Default: true, Config: map[string]string{"cdn_url": "https://cdn.example.com/"}},
	}); err != nil {
		t.Fatalf("reload with cdn: %v", err)
	}
	if got, want := reg.PublicURL("mem-a", "2026/11/x.png"), "https://cdn.example.com/2026/11/x.png"; got != want {
		t.Fatalf("PublicURL = %q, want %q", got, want)
	}
	if got := reg.PublicURL("mem-b", "2026/11/x.png"); got != "" {
		t.Fatalf("PublicURL for unconfigured storage = %q, want empty", got)
	}
}
