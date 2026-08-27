// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

func (b *Bus) Publish[T any](ctx context.Context, value T, opts ...PublishOption) error {
	if b == nil {
		return fmt.Errorf("%w: nil bus", ErrInvalidOption)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if !b.gate.Enter() {
		return ErrClosed
	}
	defer b.gate.Leave()

	cfg := publishConfig{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt.applyPublish(&cfg); err != nil {
			return err
		}
	}

	eventType := reflect.TypeFor[T]()
	meta := b.buildMetadata(ctx, eventType, value, cfg)
	if b.hooks.OnPublishStart != nil {
		info := PublishStart{
			EventType: eventType,
			Topic:     cfg.topic,
			Key:       cfg.key,
			Headers:   cloneHeaders(cfg.headers),
			Meta:      cloneHeaders(meta),
		}
		safeCall("OnPublishStart", hookPanicReporter(&b.hooks), func() { b.hooks.OnPublishStart(info) })
	}

	subs := b.snapshotSubscriptions(eventType)
	if len(subs) == 0 {
		if b.hooks.OnPublishDone != nil {
			info := PublishDone{
				EventType:            eventType,
				Topic:                cfg.topic,
				Key:                  cfg.key,
				Headers:              cloneHeaders(cfg.headers),
				Meta:                 cloneHeaders(meta),
				MatchedSubscribers:   0,
				DeliveredSubscribers: 0,
			}
			safeCall("OnPublishDone", hookPanicReporter(&b.hooks), func() { b.hooks.OnPublishDone(info) })
		}
		return nil
	}

	env := envelope{
		topic:   cfg.topic,
		key:     cfg.key,
		value:   value,
		headers: cloneHeaders(cfg.headers),
		meta:    cloneHeaders(meta),
	}

	var errs []error
	matched := 0
	delivered := 0
	for _, sub := range subs {
		if !sub.matches(env) {
			continue
		}
		matched++
		accepted, deliverErr := sub.deliver(ctx, env)
		if accepted {
			delivered++
			b.notifyObservers(ctx, Observation{
				EventType:    eventType,
				Topic:        env.topic,
				Key:          env.key,
				Headers:      cloneHeaders(env.headers),
				Meta:         cloneHeaders(env.meta),
				Value:        env.value,
				Async:        sub.async,
				SubscriberID: sub.id,
			})
		}
		if deliverErr != nil {
			errs = append(errs, deliverErr)
		}
	}

	err := errors.Join(errs...)
	if b.hooks.OnPublishDone != nil {
		info := PublishDone{
			EventType:            eventType,
			Topic:                cfg.topic,
			Key:                  cfg.key,
			Headers:              cloneHeaders(cfg.headers),
			Meta:                 cloneHeaders(meta),
			MatchedSubscribers:   matched,
			DeliveredSubscribers: delivered,
			Err:                  err,
		}
		safeCall("OnPublishDone", hookPanicReporter(&b.hooks), func() { b.hooks.OnPublishDone(info) })
	}

	return err
}

func (b *Bus) buildMetadata(ctx context.Context, eventType reflect.Type, value any, cfg publishConfig) map[string]string {
	meta := cloneHeaders(cfg.meta)
	if b.cfg.metadataBuilder == nil {
		return meta
	}

	input := PublishMetadataInput{
		Context:   ctx,
		EventType: eventType,
		Topic:     cfg.topic,
		Key:       cfg.key,
		Headers:   cloneHeaders(cfg.headers),
		Value:     value,
	}
	for k, v := range b.cfg.metadataBuilder(input) {
		if meta == nil {
			meta = make(map[string]string)
		}
		if _, exists := meta[k]; !exists {
			meta[k] = v
		}
	}
	return meta
}
