// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"fmt"
	"reflect"

	"github.com/lin-snow/ech0/pkg/busen/router"
)

type Observation struct {
	EventType reflect.Type
	Topic     string
	Key       string
	Headers   map[string]string
	Meta      map[string]string
	Value     any
	Async     bool

	SubscriberID uint64
}

type Observer func(context.Context, Observation)

type observerFilter struct {
	eventType reflect.Type
	matcher   router.Matcher
	meta      map[string]string
	predicate func(Observation) bool
}

type observerEntry struct {
	fn     Observer
	filter observerFilter
}

type ObserverOption interface {
	applyObserver(*observerFilter) error
}

type observerOptionFunc func(*observerFilter) error

func (f observerOptionFunc) applyObserver(filter *observerFilter) error {
	return f(filter)
}

func ObserveType[T any]() ObserverOption {
	return observerOptionFunc(func(filter *observerFilter) error {
		filter.eventType = reflect.TypeFor[T]()
		return nil
	})
}

func ObserveTopic(pattern string) ObserverOption {
	return observerOptionFunc(func(filter *observerFilter) error {
		matcher, err := router.Compile(pattern)
		if err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidPattern, pattern)
		}
		filter.matcher = matcher
		return nil
	})
}

func ObserveMetadata(meta map[string]string) ObserverOption {
	return observerOptionFunc(func(filter *observerFilter) error {
		filter.meta = cloneHeaders(meta)
		return nil
	})
}

func ObserveMatch(fn func(Observation) bool) ObserverOption {
	return observerOptionFunc(func(filter *observerFilter) error {
		if fn == nil {
			return fmt.Errorf("%w: observer match predicate is nil", ErrInvalidOption)
		}
		filter.predicate = fn
		return nil
	})
}

func (b *Bus) UseObserver(observer Observer, opts ...ObserverOption) error {
	if b == nil {
		return fmt.Errorf("%w: nil bus", ErrInvalidOption)
	}
	if observer == nil {
		return fmt.Errorf("%w: observer is nil", ErrInvalidOption)
	}
	if b.gate.Closed() {
		return ErrClosed
	}

	filter := observerFilter{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt.applyObserver(&filter); err != nil {
			return err
		}
	}

	b.observerMu.Lock()
	b.observers = append(b.observers, observerEntry{
		fn:     observer,
		filter: filter,
	})
	b.observerMu.Unlock()
	b.observerCount.Store(uint64(len(b.observers)))
	return nil
}

func (b *Bus) notifyObservers(ctx context.Context, obs Observation) {
	if b.observerCount.Load() == 0 {
		return
	}

	b.observerMu.RLock()
	observers := append([]observerEntry(nil), b.observers...)
	b.observerMu.RUnlock()

	for _, entry := range observers {
		if !entry.matches(obs) {
			continue
		}
		safeCall("Observer", hookPanicReporter(&b.hooks), func() {
			entry.fn(ctx, obs)
		})
	}
}

func (e observerEntry) matches(obs Observation) bool {
	if e.filter.eventType != nil && e.filter.eventType != obs.EventType {
		return false
	}
	if e.filter.matcher != nil && !e.filter.matcher.Match(obs.Topic) {
		return false
	}
	if len(e.filter.meta) > 0 {
		if len(obs.Meta) == 0 {
			return false
		}
		for k, v := range e.filter.meta {
			if obs.Meta[k] != v {
				return false
			}
		}
	}
	if e.filter.predicate != nil && !e.filter.predicate(obs) {
		return false
	}
	return true
}
