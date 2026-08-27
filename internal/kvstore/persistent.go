// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package kvstore

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Persistent struct {
	backend Backend
}

var _ Store = (*Persistent)(nil)

func NewPersistent(backend Backend) *Persistent {
	return &Persistent{backend: backend}
}

func (s *Persistent) Get(ctx context.Context, key string) (string, error) {
	v, err := s.backend.GetKeyValue(ctx, key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrNotFound
	}
	return v, err
}

func (s *Persistent) Set(ctx context.Context, key, value string) error {
	return s.backend.AddOrUpdateKeyValue(ctx, key, value)
}

func (s *Persistent) Delete(ctx context.Context, key string) error {
	return s.backend.DeleteKeyValue(ctx, key)
}
