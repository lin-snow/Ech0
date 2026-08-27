// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package setting

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/lin-snow/ech0/internal/kvstore"
)

type Spec[T any] struct {
	Key       string
	Default   func() T
	Normalize func(*T)
	Migrate   func(context.Context, kvstore.Store) (T, bool)
}

func Get[T any](ctx context.Context, kv kvstore.Store, spec Spec[T]) (T, error) {
	raw, err := kv.Get(ctx, spec.Key)
	if err != nil {
		v := spec.Default()
		if spec.Normalize != nil {
			spec.Normalize(&v)
		}
		if errors.Is(err, kvstore.ErrNotFound) {
			return v, nil
		}
		return v, err
	}

	var v T
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		v = spec.Default()
		if spec.Normalize != nil {
			spec.Normalize(&v)
		}
		return v, err
	}
	if spec.Normalize != nil {
		spec.Normalize(&v)
	}
	return v, nil
}

func Set[T any](ctx context.Context, kv kvstore.Store, spec Spec[T], value T) error {
	if spec.Normalize != nil {
		spec.Normalize(&value)
	}
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return kv.Set(ctx, spec.Key, string(buf))
}

type seedable interface {
	seed(ctx context.Context, kv kvstore.Store) error
}

func (s Spec[T]) seed(ctx context.Context, kv kvstore.Store) error {
	if _, err := kv.Get(ctx, s.Key); err == nil {
		return nil
	} else if !errors.Is(err, kvstore.ErrNotFound) {
		return err
	}

	var v T
	if s.Migrate != nil {
		if migrated, ok := s.Migrate(ctx, kv); ok {
			v = migrated
		} else {
			v = s.Default()
		}
	} else {
		v = s.Default()
	}
	if s.Normalize != nil {
		s.Normalize(&v)
	}

	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return kv.Set(ctx, s.Key, string(buf))
}

func Seed(ctx context.Context, kv kvstore.Store) error {
	for _, s := range registry {
		if err := s.seed(ctx, kv); err != nil {
			return err
		}
	}
	return nil
}
