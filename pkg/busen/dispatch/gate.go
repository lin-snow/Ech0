// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package dispatch

import (
	"context"
	"sync"
)

type Gate struct {
	mu     sync.Mutex
	closed bool
	active int
	idle   chan struct{}
}

func NewGate() *Gate {
	g := &Gate{
		idle: make(chan struct{}),
	}
	close(g.idle)
	return g
}

func (g *Gate) Enter() bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.closed {
		return false
	}

	if g.active == 0 {
		g.idle = make(chan struct{})
	}
	g.active++
	return true
}

func (g *Gate) Leave() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.active == 0 {
		return
	}
	g.active--
	if g.active == 0 {
		close(g.idle)
	}
}

func (g *Gate) Closed() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.closed
}

func (g *Gate) Close() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.closed = true
}

func (g *Gate) Wait(ctx context.Context) error {
	g.mu.Lock()
	idle := g.idle
	g.mu.Unlock()

	select {
	case <-idle:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
