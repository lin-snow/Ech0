// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package config

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/kvstore"
)

func TestEnsureJWTSecret_IgnoresLegacyEnvVariables(t *testing.T) {
	t.Setenv("JWT_SECRET", "legacy")
	t.Setenv("ECH0_JWT_SECRET", "legacy-alias")
	restoreJWTSecret(t)
	kv := kvstore.NewMemory()

	if err := EnsureJWTSecret(context.Background(), kv); err != nil {
		t.Fatalf("EnsureJWTSecret() error = %v", err)
	}
	stored, err := kv.Get(context.Background(), jwtSecretKey)
	if err != nil {
		t.Fatalf("persisted secret missing: %v", err)
	}
	if stored == "legacy" || stored == "legacy-alias" {
		t.Fatalf("env value must not be used as the signing secret, stored = %q", stored)
	}
	if got := string(Config().Security.JWTSecret); got != stored {
		t.Fatalf("pinned secret = %q, want the persisted %q", got, stored)
	}
}

func TestEnsureJWTSecret_GeneratesAndPersistsOnce(t *testing.T) {
	restoreJWTSecret(t)
	kv := kvstore.NewMemory()

	if err := EnsureJWTSecret(context.Background(), kv); err != nil {
		t.Fatalf("EnsureJWTSecret() error = %v", err)
	}
	stored, err := kv.Get(context.Background(), jwtSecretKey)
	if err != nil {
		t.Fatalf("persisted secret missing: %v", err)
	}
	if len(stored) != 64 {
		t.Fatalf("persisted secret = %q, want 64 hex chars", stored)
	}
	if got := string(Config().Security.JWTSecret); got != stored {
		t.Fatalf("pinned secret = %q, want the persisted %q", got, stored)
	}
}

func TestEnsureJWTSecret_ReusesStoredSecret(t *testing.T) {
	restoreJWTSecret(t)
	kv := kvstore.NewMemory()
	if err := kv.Set(context.Background(), jwtSecretKey, "stored-secret"); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	if err := EnsureJWTSecret(context.Background(), kv); err != nil {
		t.Fatalf("EnsureJWTSecret() error = %v", err)
	}
	if got := string(Config().Security.JWTSecret); got != "stored-secret" {
		t.Fatalf("pinned secret = %q, want %q", got, "stored-secret")
	}
	if stored, _ := kv.Get(context.Background(), jwtSecretKey); stored != "stored-secret" {
		t.Fatalf("stored secret was rewritten: %q", stored)
	}
}

func TestEnsureJWTSecret_RegeneratesBlankStoredValue(t *testing.T) {
	restoreJWTSecret(t)
	kv := kvstore.NewMemory()
	if err := kv.Set(context.Background(), jwtSecretKey, "   "); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	if err := EnsureJWTSecret(context.Background(), kv); err != nil {
		t.Fatalf("EnsureJWTSecret() error = %v", err)
	}
	stored, err := kv.Get(context.Background(), jwtSecretKey)
	if err != nil {
		t.Fatalf("persisted secret missing: %v", err)
	}
	if len(stored) != 64 {
		t.Fatalf("persisted secret = %q, want 64 hex chars", stored)
	}
}

func TestEnsureJWTSecret_PropagatesStoreErrors(t *testing.T) {
	restoreJWTSecret(t)

	if err := EnsureJWTSecret(context.Background(), failingStore{getErr: errors.New("boom")}); err == nil {
		t.Fatal("expected a read error to propagate")
	}
	if err := EnsureJWTSecret(context.Background(), failingStore{setErr: errors.New("boom")}); err == nil {
		t.Fatal("expected a write error to propagate")
	}
}

func TestEnsureJWTSecret_AdoptsValueWrittenByAnotherProcess(t *testing.T) {
	restoreJWTSecret(t)
	kv := &racingStore{}

	if err := EnsureJWTSecret(context.Background(), kv); err != nil {
		t.Fatalf("EnsureJWTSecret() error = %v", err)
	}
	if got := string(Config().Security.JWTSecret); got != "winner" {
		t.Fatalf("pinned secret = %q, want the concurrently written %q", got, "winner")
	}
}

type failingStore struct {
	getErr error
	setErr error
}

func (s failingStore) Get(context.Context, string) (string, error) {
	if s.getErr != nil {
		return "", s.getErr
	}
	return "", kvstore.ErrNotFound
}

func (s failingStore) Set(context.Context, string, string) error { return s.setErr }

func (s failingStore) Delete(context.Context, string) error { return nil }

// racingStore starts empty like a fresh database, then reports a different
// value than the one written — as if another process won the insert.
type racingStore struct {
	written bool
}

func (s *racingStore) Get(context.Context, string) (string, error) {
	if !s.written {
		return "", kvstore.ErrNotFound
	}
	return "winner", nil
}

func (s *racingStore) Set(context.Context, string, string) error {
	s.written = true
	return nil
}

func (s *racingStore) Delete(context.Context, string) error { return nil }

// restoreJWTSecret pins the process-wide config once and restores its signing
// secret after the test.
func restoreJWTSecret(t *testing.T) {
	t.Helper()
	prev := append([]byte(nil), Config().Security.JWTSecret...)
	t.Cleanup(func() { Config().Security.JWTSecret = prev })
}
