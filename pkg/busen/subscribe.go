// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/lin-snow/ech0/pkg/busen/router"
)

func (b *Bus) Subscribe[T any](handler Handler[T], opts ...SubscribeOption) (func(), error) {
	return b.subscribeWithMatcher(nil, nil, handler, opts...)
}

func (b *Bus) SubscribeTopic[T any](pattern string, handler Handler[T], opts ...SubscribeOption) (func(), error) {
	matcher, err := router.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPattern, pattern)
	}

	return b.subscribeWithMatcher(matcher, nil, handler, opts...)
}

func (b *Bus) SubscribeTopics[T any](patterns []string, handler Handler[T], opts ...SubscribeOption) (func(), error) {
	matcher, err := compileMatchers(patterns)
	if err != nil {
		return nil, err
	}

	return b.subscribeWithMatcher(matcher, nil, handler, opts...)
}

func (b *Bus) SubscribeMatch[T any](match func(Event[T]) bool, handler Handler[T], opts ...SubscribeOption) (func(), error) {
	if match == nil {
		return nil, fmt.Errorf("%w: match predicate is nil", ErrInvalidOption)
	}

	predicate := func(env envelope) bool {
		return match(typedEvent[T](env))
	}

	return b.subscribeWithMatcher(nil, predicate, handler, opts...)
}

func (b *Bus) subscribeWithMatcher[T any](
	matcher router.Matcher,
	basePredicate func(envelope) bool,
	handler Handler[T],
	opts ...SubscribeOption,
) (func(), error) {
	if b == nil {
		return nil, fmt.Errorf("%w: nil bus", ErrInvalidOption)
	}
	if handler == nil {
		return nil, ErrHandlerNil
	}

	cfg := defaultSubscribeConfig(b.cfg)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt.applySubscribe(&cfg); err != nil {
			return nil, err
		}
	}

	if cfg.parallelism <= 0 {
		cfg.parallelism = 1
	}
	if cfg.buffer <= 0 {
		cfg.buffer = b.cfg.defaultBuffer
	}
	if !cfg.overflow.valid() {
		return nil, fmt.Errorf("%w: unknown overflow policy", ErrInvalidOption)
	}

	eventType := reflect.TypeFor[T]()
	predicate := basePredicate
	if cfg.filter != nil {
		if predicate == nil {
			predicate = cfg.filter
		} else {
			prev := predicate
			predicate = func(env envelope) bool {
				return prev(env) && cfg.filter(env)
			}
		}
	}

	runtimeHandler := func(ctx context.Context, env envelope) error {
		return handler(ctx, typedEvent[T](env))
	}

	id := b.nextID.Add(1)
	sub := newSubscription(b, id, eventType, matcher, predicate, runtimeHandler, b.hooks, cfg)
	if err := b.addSubscription(eventType, sub); err != nil {
		return nil, err
	}
	if sub.async {
		sub.startWorkers()
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			sub.stopAccepting()
			b.removeSubscription(eventType, id)
			sub.scheduleStop()
		})
	}, nil
}

type matchAny []router.Matcher

func (m matchAny) Match(topic string) bool {
	for _, matcher := range m {
		if matcher.Match(topic) {
			return true
		}
	}
	return false
}

func compileMatchers(patterns []string) (router.Matcher, error) {
	if len(patterns) == 0 {
		return nil, fmt.Errorf("%w: patterns must not be empty", ErrInvalidOption)
	}

	matchers := make(matchAny, 0, len(patterns))
	for _, pattern := range patterns {
		matcher, err := router.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidPattern, pattern)
		}
		matchers = append(matchers, matcher)
	}

	return matchers, nil
}
