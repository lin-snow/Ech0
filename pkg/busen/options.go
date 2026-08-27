// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"fmt"
)

type OverflowPolicy int

const (
	OverflowBlock OverflowPolicy = iota
	OverflowFailFast
	OverflowDropNewest
	OverflowDropOldest
)

type Handler[T any] func(ctx context.Context, event Event[T]) error

type config struct {
	defaultBuffer   int
	defaultOverflow OverflowPolicy
	hooks           Hooks
	middlewares     []Middleware
	metadataBuilder MetadataBuilder
}

type subscribeConfig struct {
	async       bool
	buffer      int
	parallelism int
	overflow    OverflowPolicy
	filter      func(envelope) bool
}

type publishConfig struct {
	topic   string
	key     string
	headers map[string]string
	meta    map[string]string
}

type Option interface {
	apply(*config) error
}

type PublishOption interface {
	applyPublish(*publishConfig) error
}

type SubscribeOption interface {
	applySubscribe(*subscribeConfig) error
}

type optionFunc func(*config) error

func (f optionFunc) apply(cfg *config) error {
	return f(cfg)
}

type publishOptionFunc func(*publishConfig) error

func (f publishOptionFunc) applyPublish(cfg *publishConfig) error {
	return f(cfg)
}

type subscribeOptionFunc func(*subscribeConfig) error

func (f subscribeOptionFunc) applySubscribe(cfg *subscribeConfig) error {
	return f(cfg)
}

func WithDefaultBuffer(size int) Option {
	return optionFunc(func(cfg *config) error {
		if size <= 0 {
			return fmt.Errorf("%w: default buffer must be > 0", ErrInvalidOption)
		}
		cfg.defaultBuffer = size
		return nil
	})
}

func WithDefaultOverflow(policy OverflowPolicy) Option {
	return optionFunc(func(cfg *config) error {
		if !policy.valid() {
			return fmt.Errorf("%w: unknown overflow policy", ErrInvalidOption)
		}
		cfg.defaultOverflow = policy
		return nil
	})
}

func WithHooks(hooks Hooks) Option {
	return optionFunc(func(cfg *config) error {
		mergeHooks(&cfg.hooks, hooks)
		return nil
	})
}

func WithMiddleware(middlewares ...Middleware) Option {
	return optionFunc(func(cfg *config) error {
		for _, middleware := range middlewares {
			if middleware == nil {
				return fmt.Errorf("%w: middleware is nil", ErrInvalidOption)
			}
			cfg.middlewares = append(cfg.middlewares, middleware)
		}
		return nil
	})
}

func WithMetadataBuilder(builder MetadataBuilder) Option {
	return optionFunc(func(cfg *config) error {
		if builder == nil {
			return fmt.Errorf("%w: metadata builder is nil", ErrInvalidOption)
		}
		cfg.metadataBuilder = builder
		return nil
	})
}

func WithTopic(topic string) PublishOption {
	return publishOptionFunc(func(cfg *publishConfig) error {
		cfg.topic = topic
		return nil
	})
}

func WithKey(key string) PublishOption {
	return publishOptionFunc(func(cfg *publishConfig) error {
		cfg.key = key
		return nil
	})
}

func WithHeaders(headers map[string]string) PublishOption {
	return publishOptionFunc(func(cfg *publishConfig) error {
		cfg.headers = cloneHeaders(headers)
		return nil
	})
}

func WithMetadata(meta map[string]string) PublishOption {
	return publishOptionFunc(func(cfg *publishConfig) error {
		cfg.meta = cloneHeaders(meta)
		return nil
	})
}

func Async() SubscribeOption {
	return subscribeOptionFunc(func(cfg *subscribeConfig) error {
		cfg.async = true
		return nil
	})
}

func Sequential() SubscribeOption {
	return subscribeOptionFunc(func(cfg *subscribeConfig) error {
		cfg.async = true
		cfg.parallelism = 1
		return nil
	})
}

func WithParallelism(n int) SubscribeOption {
	return subscribeOptionFunc(func(cfg *subscribeConfig) error {
		if n <= 0 {
			return fmt.Errorf("%w: parallelism must be > 0", ErrInvalidOption)
		}
		cfg.async = true
		cfg.parallelism = n
		return nil
	})
}

func WithBuffer(size int) SubscribeOption {
	return subscribeOptionFunc(func(cfg *subscribeConfig) error {
		if size <= 0 {
			return fmt.Errorf("%w: buffer must be > 0", ErrInvalidOption)
		}
		cfg.async = true
		cfg.buffer = size
		return nil
	})
}

func WithOverflow(policy OverflowPolicy) SubscribeOption {
	return subscribeOptionFunc(func(cfg *subscribeConfig) error {
		if !policy.valid() {
			return fmt.Errorf("%w: unknown overflow policy", ErrInvalidOption)
		}
		cfg.async = true
		cfg.overflow = policy
		return nil
	})
}

func WithFilter[T any](fn func(Event[T]) bool) SubscribeOption {
	return subscribeOptionFunc(func(cfg *subscribeConfig) error {
		if fn == nil {
			return fmt.Errorf("%w: filter is nil", ErrInvalidOption)
		}

		next := func(env envelope) bool {
			return fn(typedEvent[T](env))
		}

		if cfg.filter == nil {
			cfg.filter = next
			return nil
		}

		prev := cfg.filter
		cfg.filter = func(env envelope) bool {
			return prev(env) && next(env)
		}
		return nil
	})
}

func defaultConfig() config {
	return config{
		defaultBuffer:   64,
		defaultOverflow: OverflowBlock,
	}
}

func defaultSubscribeConfig(cfg config) subscribeConfig {
	return subscribeConfig{
		buffer:      cfg.defaultBuffer,
		parallelism: 1,
		overflow:    cfg.defaultOverflow,
	}
}

func (p OverflowPolicy) valid() bool {
	return p >= OverflowBlock && p <= OverflowDropOldest
}
