// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus_test

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmit_DeliversToSubscriber(t *testing.T) {
	b := helpers.NewTestBus(t)

	var got event.EchoCreated
	unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[event.EchoCreated]) error {
		got = e.Value
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(unsub)

	want := event.EchoCreated{Echo: echoModel.Echo{ID: "echo-1"}}
	require.NoError(t, eventbus.Emit(context.Background(), b, want))
	assert.Equal(t, "echo-1", got.Echo.ID)
}

func TestEmit_KeyedEventCarriesOrderingKey(t *testing.T) {
	t.Run("keyed event attaches OrderingKey", func(t *testing.T) {
		b := helpers.NewTestBus(t)
		var key string
		unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[event.EchoCreated]) error {
			key = e.Key
			return nil
		})
		require.NoError(t, err)
		t.Cleanup(unsub)

		require.NoError(t, eventbus.Emit(context.Background(), b,
			event.EchoCreated{Echo: echoModel.Echo{ID: "echo-key"}}))
		assert.Equal(t, "echo-key", key, "Keyed event must publish with WithKey(OrderingKey)")
	})

	t.Run("keyed event with empty key publishes without key", func(t *testing.T) {
		b := helpers.NewTestBus(t)
		var key string
		unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[event.ResourceUploaded]) error {
			key = e.Key
			return nil
		})
		require.NoError(t, err)
		t.Cleanup(unsub)

		require.NoError(t, eventbus.Emit(context.Background(), b, event.ResourceUploaded{FileName: "a.png"}))
		assert.Empty(t, key)
	})

	t.Run("non-keyed event publishes without key", func(t *testing.T) {
		b := helpers.NewTestBus(t)
		var (
			key       string
			delivered bool
		)
		unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[event.SystemSnapshot]) error {
			key = e.Key
			delivered = true
			return nil
		})
		require.NoError(t, err)
		t.Cleanup(unsub)

		require.NoError(t, eventbus.Emit(context.Background(), b, event.SystemSnapshot{Info: "ok"}))
		assert.True(t, delivered)
		assert.Empty(t, key)
	})
}

func TestEmit_PropagatesHandlerError(t *testing.T) {
	b := helpers.NewTestBus(t)

	sentinel := assert.AnError
	unsub, err := b.Subscribe(func(_ context.Context, _ busen.Event[event.EchoCreated]) error {
		return sentinel
	})
	require.NoError(t, err)
	t.Cleanup(unsub)

	err = eventbus.Emit(context.Background(), b, event.EchoCreated{Echo: echoModel.Echo{ID: "x"}})
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestNotify_DeliversBestEffort(t *testing.T) {
	b := helpers.NewTestBus(t)

	var got string
	unsub, err := b.Subscribe(func(_ context.Context, e busen.Event[event.EchoCreated]) error {
		got = e.Value.Echo.ID
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(unsub)

	eventbus.Notify(context.Background(), b, event.EchoCreated{Echo: echoModel.Echo{ID: "notify-1"}})
	assert.Equal(t, "notify-1", got)
}

func TestNotify_SwallowsPublishError(t *testing.T) {
	b := busen.New()
	require.NoError(t, b.Close(context.Background()))

	assert.NotPanics(t, func() {
		eventbus.Notify(context.Background(), b, event.SystemSnapshot{Info: "after-close"})
	})
}
