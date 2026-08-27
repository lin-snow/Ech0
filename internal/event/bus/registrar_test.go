// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	"github.com/lin-snow/ech0/internal/test/helpers"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSubscriber struct {
	regs []eventbus.Registration
}

func (f fakeSubscriber) Registrations() []eventbus.Registration { return f.regs }

type drainingSubscriber struct {
	fakeSubscriber
	stopCalls int
	waitCalls int
}

func (d *drainingSubscriber) Stop() { d.stopCalls++ }
func (d *drainingSubscriber) Wait() { d.waitCalls++ }

func recordingReg(id int, unsubLog *[]int) eventbus.Registration {
	return func(_ *busen.Bus) (func(), error) {
		return func() { *unsubLog = append(*unsubLog, id) }, nil
	}
}

func newProvider(t *testing.T) func() *busen.Bus {
	t.Helper()
	b := helpers.NewTestBus(t)
	return func() *busen.Bus { return b }
}

func TestEventRegistrar_RegisterIsIdempotent(t *testing.T) {
	var registerCalls int
	mk := func() eventbus.Registration {
		return func(_ *busen.Bus) (func(), error) {
			registerCalls++
			return func() {}, nil
		}
	}
	sub := fakeSubscriber{regs: []eventbus.Registration{mk(), mk()}}
	reg := eventbus.NewEventRegistry(newProvider(t), []eventbus.Subscriber{sub})

	require.NoError(t, reg.Register())
	require.NoError(t, reg.Register(), "second Register must be a no-op")
	assert.Equal(t, 2, registerCalls, "registrations must run exactly once across repeated Register calls")
}

func TestEventRegistrar_RollbackOnRegistrationFailure(t *testing.T) {
	boom := errors.New("registration boom")
	var unsubLog []int

	failing := func(_ *busen.Bus) (func(), error) { return nil, boom }
	sub := fakeSubscriber{regs: []eventbus.Registration{
		recordingReg(1, &unsubLog),
		recordingReg(2, &unsubLog),
		failing,
	}}
	reg := eventbus.NewEventRegistry(newProvider(t), []eventbus.Subscriber{sub})

	err := reg.Register()
	require.ErrorIs(t, err, boom)
	assert.Equal(t, []int{2, 1}, unsubLog)
}

func TestEventRegistrar_RegisterAgainAfterFailure(t *testing.T) {
	var attempts int
	fail := true
	sub := fakeSubscriber{regs: []eventbus.Registration{
		func(_ *busen.Bus) (func(), error) {
			attempts++
			if fail {
				return nil, errors.New("transient")
			}
			return func() {}, nil
		},
	}}
	reg := eventbus.NewEventRegistry(newProvider(t), []eventbus.Subscriber{sub})

	require.Error(t, reg.Register())
	fail = false
	require.NoError(t, reg.Register())
	assert.Equal(t, 2, attempts, "Register must retry after a prior failure")
}

func TestEventRegistrar_StopUnsubscribesInReverseOrder(t *testing.T) {
	var unsubLog []int
	sub := fakeSubscriber{regs: []eventbus.Registration{
		recordingReg(1, &unsubLog),
		recordingReg(2, &unsubLog),
		recordingReg(3, &unsubLog),
	}}
	reg := eventbus.NewEventRegistry(newProvider(t), []eventbus.Subscriber{sub})

	require.NoError(t, reg.Register())
	require.NoError(t, reg.Stop())
	assert.Equal(t, []int{3, 2, 1}, unsubLog)
}

func TestEventRegistrar_StopDrainsDrainingSubscribers(t *testing.T) {
	var unsubLog []int
	d := &drainingSubscriber{
		regs: []eventbus.Registration{recordingReg(1, &unsubLog)},
	}
	reg := eventbus.NewEventRegistry(newProvider(t), []eventbus.Subscriber{d})

	require.NoError(t, reg.Register())
	require.NoError(t, reg.Stop())

	assert.Equal(t, []int{1}, unsubLog, "subscriptions are torn down before draining")
	assert.Equal(t, 1, d.stopCalls)
	assert.Equal(t, 1, d.waitCalls)
}

func TestEventRegistrar_StopBeforeRegisterIsNoop(t *testing.T) {
	var unsubLog []int
	d := &drainingSubscriber{
		regs: []eventbus.Registration{recordingReg(1, &unsubLog)},
	}
	reg := eventbus.NewEventRegistry(newProvider(t), []eventbus.Subscriber{d})

	require.NoError(t, reg.Stop(), "Stop before Register must be a no-op")
	assert.Empty(t, unsubLog)
	assert.Zero(t, d.stopCalls)
	assert.Zero(t, d.waitCalls)
}

func TestEventRegistrar_SkipsNilSubscribers(t *testing.T) {
	var registered bool
	sub := fakeSubscriber{regs: []eventbus.Registration{
		func(_ *busen.Bus) (func(), error) {
			registered = true
			return func() {}, nil
		},
	}}
	reg := eventbus.NewEventRegistry(newProvider(t), []eventbus.Subscriber{nil, sub})

	require.NoError(t, reg.Register())
	assert.True(t, registered, "non-nil subscribers must still register when a nil entry precedes them")
}

func TestEventRegistrar_EndToEndWithRealBus(t *testing.T) {
	b := helpers.NewTestBus(t)
	provider := func() *busen.Bus { return b }

	var delivered int
	sub := fakeSubscriber{regs: []eventbus.Registration{
		eventbus.On(func(_ context.Context, _ event.EchoCreated) error {
			delivered++
			return nil
		}),
	}}
	reg := eventbus.NewEventRegistry(provider, []eventbus.Subscriber{sub})
	require.NoError(t, reg.Register())

	eventbus.Notify(context.Background(), b, event.EchoCreated{Echo: echoModel.Echo{ID: "e2e-1"}})
	assert.Equal(t, 1, delivered)

	require.NoError(t, reg.Stop())
	eventbus.Notify(context.Background(), b, event.EchoCreated{Echo: echoModel.Echo{ID: "e2e-2"}})
	assert.Equal(t, 1, delivered)
}
