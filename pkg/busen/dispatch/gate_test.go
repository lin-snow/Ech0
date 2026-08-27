// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package dispatch_test

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/pkg/busen/dispatch"
)

func TestGateEnterClosedTransitions(t *testing.T) {
	g := dispatch.NewGate()

	if g.Closed() {
		t.Fatalf("new gate should be open")
	}
	if !g.Enter() {
		t.Fatalf("Enter on open gate should succeed")
	}
	g.Leave()

	g.Close()
	if !g.Closed() {
		t.Fatalf("gate should report closed after Close")
	}
	if g.Enter() {
		t.Fatalf("Enter after Close must return false")
	}
}

func TestGateWaitImmediateWhenIdle(t *testing.T) {
	g := dispatch.NewGate()

	if err := g.Wait(context.Background()); err != nil {
		t.Fatalf("Wait on idle gate = %v, want nil", err)
	}
}

func TestGateWaitUnblocksWhenActiveReachesZero(t *testing.T) {
	g := dispatch.NewGate()
	if !g.Enter() {
		t.Fatalf("Enter should succeed")
	}

	done := make(chan error, 1)
	go func() {
		done <- g.Wait(context.Background())
	}()

	g.Leave()

	if err := <-done; err != nil {
		t.Fatalf("Wait after Leave = %v, want nil", err)
	}
}

func TestGateWaitReturnsCtxErrOnCancel(t *testing.T) {
	g := dispatch.NewGate()
	if !g.Enter() {
		t.Fatalf("Enter should succeed")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- g.Wait(ctx)
	}()

	cancel()

	if err := <-done; err != context.Canceled {
		t.Fatalf("Wait after cancel = %v, want context.Canceled", err)
	}
}

func TestGateActiveCountGatesIdle(t *testing.T) {
	g := dispatch.NewGate()

	if !g.Enter() {
		t.Fatalf("first Enter should succeed")
	}
	if !g.Enter() {
		t.Fatalf("second Enter should succeed")
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := g.Wait(canceled); err != context.Canceled {
		t.Fatalf("Wait while active = %v, want context.Canceled", err)
	}

	g.Leave()
	if err := g.Wait(canceled); err != context.Canceled {
		t.Fatalf("Wait with one op still in flight = %v, want context.Canceled", err)
	}

	g.Leave()
	if err := g.Wait(context.Background()); err != nil {
		t.Fatalf("Wait after all ops drained = %v, want nil", err)
	}
}

func TestGateLeaveUnderflowIsSafeNoop(t *testing.T) {
	g := dispatch.NewGate()

	for range 3 {
		g.Leave()
	}

	if err := g.Wait(context.Background()); err != nil {
		t.Fatalf("Wait after no-op Leave = %v, want nil", err)
	}

	if !g.Enter() {
		t.Fatalf("Enter after no-op Leave should succeed")
	}
	g.Leave()
	if err := g.Wait(context.Background()); err != nil {
		t.Fatalf("Wait after Enter/Leave cycle = %v, want nil", err)
	}

	g.Leave()
	if err := g.Wait(context.Background()); err != nil {
		t.Fatalf("Wait after post-cycle surplus Leave = %v, want nil", err)
	}
}
