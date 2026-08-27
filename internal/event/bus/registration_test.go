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

func TestOn_SubscribesAndUnsubscribes(t *testing.T) {
	b := helpers.NewTestBus(t)

	var count int
	reg := eventbus.On(func(_ context.Context, e event.EchoCreated) error {
		count++
		assert.Equal(t, "on-1", e.Echo.ID)
		return nil
	})

	unsub, err := reg(b)
	require.NoError(t, err)

	require.NoError(t, eventbus.Emit(context.Background(), b, event.EchoCreated{Echo: echoModel.Echo{ID: "on-1"}}))
	assert.Equal(t, 1, count)

	unsub()
	require.NoError(t, eventbus.Emit(context.Background(), b, event.EchoCreated{Echo: echoModel.Echo{ID: "on-1"}}))
	assert.Equal(t, 1, count, "handler should not run after unsubscribe")
}

func TestOnWithMeta_ForwardsMetadata(t *testing.T) {
	b := helpers.NewTestBus(t)

	var gotMeta map[string]string
	reg := eventbus.OnWithMeta(func(_ context.Context, e event.EchoCreated, meta map[string]string) error {
		gotMeta = meta
		return nil
	})
	unsub, err := reg(b)
	require.NoError(t, err)
	t.Cleanup(unsub)

	require.NoError(t, b.Publish(context.Background(),
		event.EchoCreated{Echo: echoModel.Echo{ID: "meta-1"}},
		busen.WithMetadata(map[string]string{"source": "ech0", "trace_id": "t-9"})))

	require.NotNil(t, gotMeta)
	assert.Equal(t, "ech0", gotMeta["source"])
	assert.Equal(t, "t-9", gotMeta["trace_id"])
}

func TestOn_DropsMetadata(t *testing.T) {
	b := helpers.NewTestBus(t)

	var (
		gotValue event.EchoCreated
		fired    bool
	)
	reg := eventbus.On(func(_ context.Context, e event.EchoCreated) error {
		gotValue = e
		fired = true
		return nil
	})
	unsub, err := reg(b)
	require.NoError(t, err)
	t.Cleanup(unsub)

	require.NoError(t, b.Publish(context.Background(),
		event.EchoCreated{Echo: echoModel.Echo{ID: "drop-1"}},
		busen.WithMetadata(map[string]string{"source": "ech0"})))

	assert.True(t, fired)
	assert.Equal(t, "drop-1", gotValue.Echo.ID)
}

func TestOn_PropagatesNilBusError(t *testing.T) {
	reg := eventbus.On(func(_ context.Context, _ event.EchoCreated) error { return nil })
	_, err := reg(nil)
	require.Error(t, err, "subscribing on a nil bus must error")
}
