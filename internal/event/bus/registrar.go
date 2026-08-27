// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package bus

import (
	"slices"
	"sync/atomic"

	"github.com/lin-snow/ech0/pkg/busen"
)

type Draining interface {
	Stop()
	Wait()
}

type EventRegistrar struct {
	bus         *busen.Bus
	subscribers []Subscriber
	unsub       []func()
	registered  atomic.Bool
}

func NewEventRegistry(
	busProvider func() *busen.Bus,
	subscribers []Subscriber,
) *EventRegistrar {
	return &EventRegistrar{
		bus:         busProvider(),
		subscribers: subscribers,
	}
}

func (er *EventRegistrar) Register() error {
	if er.registered.Load() {
		return nil
	}

	for _, sub := range er.subscribers {
		if sub == nil {
			continue
		}
		for _, reg := range sub.Registrations() {
			unsub, err := reg(er.bus)
			if err != nil {
				er.stopSubscriptions()
				return err
			}
			er.unsub = append(er.unsub, unsub)
		}
	}

	er.registered.Store(true)
	return nil
}

func (er *EventRegistrar) Stop() error {
	if !er.registered.Load() {
		return nil
	}
	er.stopSubscriptions()
	for _, sub := range er.subscribers {
		if d, ok := sub.(Draining); ok {
			d.Stop()
			d.Wait()
		}
	}
	return nil
}

func (er *EventRegistrar) stopSubscriptions() {
	for _, v := range slices.Backward(er.unsub) {
		v()
	}
	er.unsub = nil
}
