// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package kvstore

import (
	"context"
	"sync"
)

type Memory struct {
	mu sync.RWMutex
	m  map[string]string
}

var _ Store = (*Memory)(nil)

func NewMemory() *Memory {
	return &Memory{m: make(map[string]string)}
}

func (s *Memory) Get(_ context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (s *Memory) Set(_ context.Context, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
	return nil
}

func (s *Memory) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
	return nil
}
