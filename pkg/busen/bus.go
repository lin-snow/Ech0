// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	intdispatch "github.com/lin-snow/ech0/pkg/busen/dispatch"
	"github.com/lin-snow/ech0/pkg/busen/router"
)

type Bus struct {
	cfg   config
	hooks Hooks
	gate  *intdispatch.Gate

	mu           sync.RWMutex
	subs         map[reflect.Type]map[uint64]*subscription
	subSnapshots map[reflect.Type][]*subscription
	middlewareMu sync.RWMutex
	middlewares  []Middleware
	middleware   func(Next) Next
	observerMu   sync.RWMutex
	observers    []observerEntry

	middlewareVersion atomic.Uint64
	observerCount     atomic.Uint64

	nextID atomic.Uint64
}

type subscription struct {
	id        uint64
	eventType reflect.Type
	bus       *Bus

	matcher   router.Matcher
	predicate func(envelope) bool
	handler   func(context.Context, envelope) error
	hooks     Hooks

	async       bool
	parallelism int
	overflow    OverflowPolicy
	mailboxes   []chan workItem
	rr          atomic.Uint64

	queueMu  sync.Mutex
	gate     *intdispatch.Gate
	stopOnce sync.Once
	done     chan struct{}

	dispatchMu      sync.Mutex
	dispatchVersion uint64
	dispatchHandler Next

	processedCount    atomic.Int64
	droppedCount      atomic.Int64
	rejectedCount     atomic.Int64
	shutdownDropCount atomic.Int64
}

type workItem struct {
	ctx context.Context
	env envelope
}

func New(opts ...Option) *Bus {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt.apply(&cfg); err != nil {
			panic(err)
		}
	}

	return &Bus{
		cfg:          cfg,
		hooks:        cfg.hooks,
		gate:         intdispatch.NewGate(),
		subs:         make(map[reflect.Type]map[uint64]*subscription),
		subSnapshots: make(map[reflect.Type][]*subscription),
		middlewares:  append([]Middleware(nil), cfg.middlewares...),
		middleware:   buildMiddlewareChain(cfg.middlewares),
	}
}

func (b *Bus) Close(ctx context.Context) error {
	_, err := b.Shutdown(ctx, ShutdownDrain)
	return err
}

func (b *Bus) allSubscriptions() []*subscription {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]*subscription, 0, len(b.subs))
	for _, group := range b.subs {
		for _, sub := range group {
			out = append(out, sub)
		}
	}
	return out
}

func (b *Bus) addSubscription(eventType reflect.Type, sub *subscription) error {
	if b.gate.Closed() {
		return ErrClosed
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.gate.Closed() {
		return ErrClosed
	}

	group := b.subs[eventType]
	if group == nil {
		group = make(map[uint64]*subscription)
		b.subs[eventType] = group
	}
	group[sub.id] = sub
	b.subSnapshots[eventType] = snapshotGroup(group)
	return nil
}

func (b *Bus) removeSubscription(eventType reflect.Type, id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	group := b.subs[eventType]
	if group == nil {
		return
	}

	delete(group, id)
	if len(group) == 0 {
		delete(b.subs, eventType)
		delete(b.subSnapshots, eventType)
		return
	}
	b.subSnapshots[eventType] = snapshotGroup(group)
}

func (b *Bus) snapshotSubscriptions(eventType reflect.Type) []*subscription {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.subSnapshots[eventType]
}

func newSubscription(
	bus *Bus,
	id uint64,
	eventType reflect.Type,
	matcher router.Matcher,
	predicate func(envelope) bool,
	handler func(context.Context, envelope) error,
	hooks Hooks,
	cfg subscribeConfig,
) *subscription {
	sub := &subscription{
		id:          id,
		eventType:   eventType,
		bus:         bus,
		matcher:     matcher,
		predicate:   predicate,
		handler:     handler,
		hooks:       hooks,
		async:       cfg.async,
		parallelism: cfg.parallelism,
		overflow:    cfg.overflow,
		gate:        intdispatch.NewGate(),
		done:        make(chan struct{}),
	}

	if sub.async {
		sub.mailboxes = make([]chan workItem, cfg.parallelism)
		for i := range sub.mailboxes {
			sub.mailboxes[i] = make(chan workItem, cfg.buffer)
		}
	} else {
		close(sub.done)
	}

	return sub
}

func (s *subscription) matches(env envelope) bool {
	if s.matcher != nil && !s.matcher.Match(env.topic) {
		return false
	}
	if s.predicate != nil && !s.predicate(env) {
		return false
	}
	return true
}

func snapshotGroup(group map[uint64]*subscription) []*subscription {
	if len(group) == 0 {
		return nil
	}

	out := make([]*subscription, 0, len(group))
	for _, sub := range group {
		out = append(out, sub)
	}
	return out
}

func closeWaitError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrCloseIncomplete, err)
}

func (s *subscription) deliver(ctx context.Context, env envelope) (bool, error) {
	if !s.gate.Enter() {
		return false, nil
	}
	defer s.gate.Leave()

	if !s.async {
		return true, s.callHandler(ctx, env)
	}

	return true, s.enqueue(ctx, env)
}

func (s *subscription) enqueue(ctx context.Context, env envelope) error {
	item := workItem{ctx: ctx, env: env}
	mailbox, mailboxIndex := s.mailboxForKey(env.key)

	switch s.overflow {
	case OverflowBlock:
		select {
		case mailbox <- item:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	case OverflowFailFast:
		select {
		case mailbox <- item:
			return nil
		default:
			s.onRejected(env, ErrBufferFull, mailbox, mailboxIndex)
			return ErrBufferFull
		}
	case OverflowDropNewest:
		select {
		case mailbox <- item:
			return nil
		default:
			s.onDropped(env, ErrDropped, mailbox, mailboxIndex)
			return ErrDropped
		}
	case OverflowDropOldest:
		s.queueMu.Lock()
		defer s.queueMu.Unlock()

		select {
		case mailbox <- item:
			return nil
		default:
		}

		select {
		case dropped := <-mailbox:
			s.onDropped(dropped.env, ErrDropped, mailbox, mailboxIndex)
		default:
			return ErrBufferFull
		}

		select {
		case mailbox <- item:
			return ErrDropped
		default:
			return ErrBufferFull
		}
	default:
		return fmt.Errorf("%w: unknown overflow policy", ErrInvalidOption)
	}
}

func (s *subscription) callHandler(ctx context.Context, env envelope) (err error) {
	defer s.processedCount.Add(1)

	dispatch := Dispatch{
		EventType: s.eventType,
		Topic:     env.topic,
		Key:       env.key,
		Headers:   cloneHeaders(env.headers),
		Meta:      cloneHeaders(env.meta),
		Value:     env.value,
		Async:     s.async,
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			err = &HandlerPanicError{
				Value: recovered,
			}
			s.onPanic(envelope{
				topic:   dispatch.Topic,
				key:     dispatch.Key,
				value:   dispatch.Value,
				headers: dispatch.Headers,
				meta:    dispatch.Meta,
			}, recovered)
		}
	}()

	err = s.invoke(ctx, dispatch)
	if err != nil {
		s.onError(envelope{
			topic:   dispatch.Topic,
			key:     dispatch.Key,
			value:   dispatch.Value,
			headers: dispatch.Headers,
			meta:    dispatch.Meta,
		}, err)
	}
	return err
}

func (s *subscription) startWorkers() {
	var wg sync.WaitGroup
	for _, mailbox := range s.mailboxes {
		wg.Go(func() {
			for item := range mailbox {
				_ = s.callHandler(item.ctx, item.env)
			}
		})
	}

	go func() {
		wg.Wait()
		close(s.done)
	}()
}

func (s *subscription) stopAccepting() {
	s.gate.Close()
}

func (s *subscription) scheduleStop() {
	if !s.async {
		return
	}

	s.stopOnce.Do(func() {
		go func() {
			_ = s.gate.Wait(context.Background())
			for _, mailbox := range s.mailboxes {
				close(mailbox)
			}
		}()
	})
}

func (s *subscription) waitStopped(ctx context.Context) error {
	if !s.async {
		return s.gate.Wait(ctx)
	}

	s.scheduleStop()

	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *subscription) onError(env envelope, err error) {
	if s.hooks.OnHandlerError == nil {
		return
	}

	info := HandlerError{
		EventType: s.eventType,
		Topic:     env.topic,
		Key:       env.key,
		Meta:      cloneHeaders(env.meta),
		Async:     s.async,
		Err:       err,
	}
	safeCall("OnHandlerError", hookPanicReporter(&s.hooks), func() { s.hooks.OnHandlerError(info) })
}

func (s *subscription) onPanic(env envelope, value any) {
	if s.hooks.OnHandlerPanic == nil {
		return
	}

	info := HandlerPanic{
		EventType: s.eventType,
		Topic:     env.topic,
		Key:       env.key,
		Meta:      cloneHeaders(env.meta),
		Async:     s.async,
		Value:     value,
	}
	safeCall("OnHandlerPanic", hookPanicReporter(&s.hooks), func() { s.hooks.OnHandlerPanic(info) })
}

func (s *subscription) onDropped(env envelope, reason error, mailbox chan workItem, mailboxIndex int) {
	if s.hooks.OnEventDropped == nil {
		s.droppedCount.Add(1)
		return
	}
	s.droppedCount.Add(1)

	info := DroppedEvent{
		EventType:    s.eventType,
		Topic:        env.topic,
		Key:          env.key,
		Meta:         cloneHeaders(env.meta),
		Async:        true,
		Policy:       s.overflow,
		SubscriberID: s.id,
		QueueLen:     len(mailbox),
		QueueCap:     cap(mailbox),
		MailboxIndex: mailboxIndex,
		Reason:       reason,
	}
	safeCall("OnEventDropped", hookPanicReporter(&s.hooks), func() { s.hooks.OnEventDropped(info) })
}

func (s *subscription) onRejected(env envelope, reason error, mailbox chan workItem, mailboxIndex int) {
	if s.hooks.OnEventRejected == nil {
		s.rejectedCount.Add(1)
		return
	}
	s.rejectedCount.Add(1)

	info := RejectedEvent{
		EventType:    s.eventType,
		Topic:        env.topic,
		Key:          env.key,
		Meta:         cloneHeaders(env.meta),
		Async:        true,
		Policy:       s.overflow,
		SubscriberID: s.id,
		QueueLen:     len(mailbox),
		QueueCap:     cap(mailbox),
		MailboxIndex: mailboxIndex,
		Reason:       reason,
	}
	safeCall("OnEventRejected", hookPanicReporter(&s.hooks), func() { s.hooks.OnEventRejected(info) })
}

func (s *subscription) mailboxForKey(key string) (chan workItem, int) {
	if len(s.mailboxes) == 1 {
		return s.mailboxes[0], 0
	}
	if key != "" {
		idx := int(hashString(key) % uint64(len(s.mailboxes)))
		return s.mailboxes[idx], idx
	}
	next := s.rr.Add(1) - 1
	idx := int(next % uint64(len(s.mailboxes)))
	return s.mailboxes[idx], idx
}

type subscriptionStats struct {
	processed    int64
	dropped      int64
	rejected     int64
	shutdownDrop int64
}

func (s *subscription) snapshotStats() subscriptionStats {
	return subscriptionStats{
		processed:    s.processedCount.Load(),
		dropped:      s.droppedCount.Load(),
		rejected:     s.rejectedCount.Load(),
		shutdownDrop: s.shutdownDropCount.Load(),
	}
}

func (s *subscription) abortPending() int64 {
	if !s.async {
		return 0
	}

	var dropped int64
	for _, mailbox := range s.mailboxes {
		for {
			select {
			case <-mailbox:
				dropped++
			default:
				goto drained
			}
		}
	drained:
	}

	if dropped > 0 {
		s.shutdownDropCount.Add(dropped)
	}
	return dropped
}

func (s *subscription) invoke(ctx context.Context, dispatch Dispatch) error {
	return s.dispatch(ctx, dispatch)(ctx, dispatch)
}

func (s *subscription) dispatch(ctx context.Context, dispatch Dispatch) Next {
	bus := s.bus
	if bus == nil {
		return func(ctx context.Context, dispatch Dispatch) error {
			return s.handler(ctx, envelope{
				topic:   dispatch.Topic,
				key:     dispatch.Key,
				value:   dispatch.Value,
				headers: dispatch.Headers,
				meta:    dispatch.Meta,
			})
		}
	}

	version := bus.middlewareVersion.Load()

	s.dispatchMu.Lock()
	defer s.dispatchMu.Unlock()

	if s.dispatchHandler != nil && s.dispatchVersion == version {
		return s.dispatchHandler
	}

	final := func(ctx context.Context, dispatch Dispatch) error {
		return s.handler(ctx, envelope{
			topic:   dispatch.Topic,
			key:     dispatch.Key,
			value:   dispatch.Value,
			headers: dispatch.Headers,
			meta:    dispatch.Meta,
		})
	}

	chain := bus.currentMiddlewareChain()
	if chain == nil {
		s.dispatchHandler = final
		s.dispatchVersion = version
		return final
	}

	s.dispatchHandler = chain(final)
	s.dispatchVersion = version
	return s.dispatchHandler
}

func (b *Bus) currentMiddlewareChain() func(Next) Next {
	b.middlewareMu.RLock()
	defer b.middlewareMu.RUnlock()
	return b.middleware
}

type HandlerPanicError struct {
	Value any
}

func (e *HandlerPanicError) Error() string {
	return fmt.Sprintf("%s: %v", ErrHandlerPanic, e.Value)
}

func (e *HandlerPanicError) Unwrap() error {
	return ErrHandlerPanic
}

func hashString(value string) uint64 {
	const offset64 = 14695981039346656037
	const prime64 = 1099511628211

	hash := uint64(offset64)
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= prime64
	}
	return hash
}
