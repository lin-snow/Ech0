// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package setting

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/kvstore"
)

// Spec declares one stored settings block: its key, the value that applies
// before anything is stored, and the rules every write must satisfy.
type Spec[T any] struct {
	Key       string
	Default   func() T
	Normalize func(*T)
	// Validate guards operator writes. Restores skip it on purpose: the source
	// instance already ran with that configuration.
	Validate func(T) error
}

// Get reads the stored block over its defaults. Fields absent from the stored
// JSON keep their default, so adding a field needs no migration and an older
// blob can never pin a new field to its zero value.
func (s Spec[T]) Get(ctx context.Context, kv kvstore.Store) (T, error) {
	v := s.Default()
	raw, err := kv.Get(ctx, s.Key)
	if err == nil {
		if unmarshalErr := json.Unmarshal([]byte(raw), &v); unmarshalErr != nil {
			v = s.Default()
			err = unmarshalErr
		}
	} else if errors.Is(err, kvstore.ErrNotFound) {
		err = nil
	}
	s.normalize(&v)
	return v, err
}

// Set normalizes and validates before persisting. Every operator-facing writer
// goes through it, including import paths that bypass the settings service.
func (s Spec[T]) Set(ctx context.Context, kv kvstore.Store, value T) error {
	s.normalize(&value)
	if s.Validate != nil {
		if err := s.Validate(value); err != nil {
			return err
		}
	}
	return s.write(ctx, kv, value)
}

// Restore persists a block decoded from another instance's data (snapshot or
// capsule import). It normalizes but does not validate, so a configuration the
// source instance was running with always lands.
func (s Spec[T]) Restore(ctx context.Context, kv kvstore.Store, value T) error {
	s.normalize(&value)
	return s.write(ctx, kv, value)
}

// Unset removes the stored block, so the defaults apply again.
func (s Spec[T]) Unset(ctx context.Context, kv kvstore.Store) error {
	if err := kv.Delete(ctx, s.Key); err != nil && !errors.Is(err, kvstore.ErrNotFound) {
		return err
	}
	return nil
}

func (s Spec[T]) write(ctx context.Context, kv kvstore.Store, value T) error {
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return kv.Set(ctx, s.Key, string(buf))
}

// DefaultValue is the normalized default: what Get yields before anything is
// stored.
func (s Spec[T]) DefaultValue() T {
	v := s.Default()
	s.normalize(&v)
	return v
}

func (s Spec[T]) normalize(v *T) {
	if s.Normalize != nil {
		s.Normalize(v)
	}
}

// Migrate applies the one-shot data moves. Stored blocks are deliberately not
// materialized: an absent block reads as its default, so a changed default
// reaches every instance that never overrode it.
func Migrate(ctx context.Context, kv kvstore.Store) error {
	for _, migrate := range migrations {
		if err := migrate(ctx, kv); err != nil {
			return err
		}
	}
	return nil
}
