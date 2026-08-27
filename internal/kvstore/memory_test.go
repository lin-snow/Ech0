// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package kvstore

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryGetMissReturnsNotFound(t *testing.T) {
	m := NewMemory()
	if _, err := m.Get(context.Background(), "absent"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestMemorySetGetDelete(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()

	if err := m.Set(ctx, "k", "v1"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if got, err := m.Get(ctx, "k"); err != nil || got != "v1" {
		t.Fatalf("get after set: got %q err %v", got, err)
	}

	if err := m.Set(ctx, "k", "v2"); err != nil {
		t.Fatalf("set upsert: %v", err)
	}
	if got, _ := m.Get(ctx, "k"); got != "v2" {
		t.Fatalf("expected upsert to v2, got %q", got)
	}

	if err := m.Delete(ctx, "k"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := m.Get(ctx, "k"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
