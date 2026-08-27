// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func recvWithin[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for delivery")
		var zero T
		return zero
	}
}

func TestSafeTypeString(t *testing.T) {
	cases := []struct {
		name string
		in   reflect.Type
		want string
	}{
		{"nil", nil, ""},
		{"int", reflect.TypeFor[int](), "int"},
		{"string", reflect.TypeFor[string](), "string"},
		{"pointer", reflect.TypeOf(&struct{}{}), "*struct {}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, safeTypeString(tc.in))
		})
	}
}

func TestMapOverflow(t *testing.T) {
	cases := []struct {
		name   string
		policy string
		want   busen.OverflowPolicy
	}{
		{"fail_fast", "fail_fast", busen.OverflowFailFast},
		{"drop_newest", "drop_newest", busen.OverflowDropNewest},
		{"drop_oldest", "drop_oldest", busen.OverflowDropOldest},
		{"block", "block", busen.OverflowBlock},
		{"empty falls back to block", "", busen.OverflowBlock},
		{"unknown falls back to block", "nonsense", busen.OverflowBlock},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, MapOverflow(tc.policy))
		})
	}
}

type ping struct{ n int }

func TestNew_DeliversSynchronously(t *testing.T) {
	b := New()
	require.NotNil(t, b)
	t.Cleanup(func() { _ = b.Close(context.Background()) })

	var got int
	unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[ping]) error {
		got = e.Value.n
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(unsub)

	require.NoError(t, b.Publish(context.Background(), ping{n: 7}))
	assert.Equal(t, 7, got)
}

func TestProvideProvider_ReturnsSingleton(t *testing.T) {
	provider := ProvideProvider()
	b1 := provider()
	b2 := provider()
	require.NotNil(t, b1)
	assert.Same(t, b1, b2, "ProvideProvider must memoize a single bus instance")
	t.Cleanup(func() { _ = b1.Close(context.Background()) })
}

func TestAsyncParallel_EnablesAsyncDelivery(t *testing.T) {
	opts := AsyncParallel()
	require.NotEmpty(t, opts)

	b := New()
	t.Cleanup(func() { _ = b.Close(context.Background()) })

	done := make(chan int, 1)
	unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[ping]) error {
		done <- e.Value.n
		return nil
	}, opts...)
	require.NoError(t, err)
	t.Cleanup(unsub)

	require.NoError(t, b.Publish(context.Background(), ping{n: 42}))
	assert.Equal(t, 42, recvWithin(t, done))
}

func TestAsyncSequential_EnablesAsyncDelivery(t *testing.T) {
	opts := AsyncSequential()
	require.NotEmpty(t, opts)

	b := New()
	t.Cleanup(func() { _ = b.Close(context.Background()) })

	done := make(chan int, 1)
	unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[ping]) error {
		done <- e.Value.n
		return nil
	}, opts...)
	require.NoError(t, err)
	t.Cleanup(unsub)

	require.NoError(t, b.Publish(context.Background(), ping{n: 99}))
	assert.Equal(t, 99, recvWithin(t, done))
}

func TestNew_HooksHandleErrorPaths(t *testing.T) {
	t.Run("handler error is surfaced and logged", func(t *testing.T) {
		b := New()
		t.Cleanup(func() { _ = b.Close(context.Background()) })

		sentinel := errors.New("handler failed")
		unsub, err := b.Subscribe(func(_ context.Context, _ busen.Event[ping]) error {
			return sentinel
		})
		require.NoError(t, err)
		t.Cleanup(unsub)

		err = b.Publish(context.Background(), ping{n: 1})
		require.Error(t, err)
		assert.ErrorIs(t, err, sentinel)
	})

	t.Run("handler panic is recovered, not propagated", func(t *testing.T) {
		b := New()
		t.Cleanup(func() { _ = b.Close(context.Background()) })

		unsub, err := b.Subscribe(func(_ context.Context, _ busen.Event[ping]) error {
			panic("boom in handler")
		})
		require.NoError(t, err)
		t.Cleanup(unsub)

		var pubErr error
		assert.NotPanics(t, func() {
			pubErr = b.Publish(context.Background(), ping{n: 2})
		})
		assert.ErrorIs(t, pubErr, busen.ErrHandlerPanic)
	})
}
