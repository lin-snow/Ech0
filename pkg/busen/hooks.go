// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"reflect"
)

type Hooks struct {
	OnPublishStart  func(PublishStart)
	OnPublishDone   func(PublishDone)
	OnHandlerError  func(HandlerError)
	OnHandlerPanic  func(HandlerPanic)
	OnEventDropped  func(DroppedEvent)
	OnEventRejected func(RejectedEvent)
	OnHookPanic     func(HookPanic)
}

type PublishStart struct {
	EventType reflect.Type
	Topic     string
	Key       string
	Headers   map[string]string
	Meta      map[string]string
}

type PublishDone struct {
	EventType            reflect.Type
	Topic                string
	Key                  string
	Headers              map[string]string
	Meta                 map[string]string
	MatchedSubscribers   int
	DeliveredSubscribers int
	Err                  error
}

type HandlerError struct {
	EventType reflect.Type
	Topic     string
	Key       string
	Meta      map[string]string
	Async     bool
	Err       error
}

type HandlerPanic struct {
	EventType reflect.Type
	Topic     string
	Key       string
	Meta      map[string]string
	Async     bool
	Value     any
}

type DroppedEvent struct {
	EventType    reflect.Type
	Topic        string
	Key          string
	Meta         map[string]string
	Async        bool
	Policy       OverflowPolicy
	SubscriberID uint64
	QueueLen     int
	QueueCap     int
	MailboxIndex int
	Reason       error
}

type RejectedEvent struct {
	EventType    reflect.Type
	Topic        string
	Key          string
	Meta         map[string]string
	Async        bool
	Policy       OverflowPolicy
	SubscriberID uint64
	QueueLen     int
	QueueCap     int
	MailboxIndex int
	Reason       error
}

type HookPanic struct {
	Hook  string
	Value any
}

func mergeHooks(dst *Hooks, src Hooks) {
	if dst == nil {
		return
	}

	dst.OnHookPanic = chainHookPanic(dst.OnHookPanic, src.OnHookPanic)
	dst.OnPublishStart = chainPublishStart(dst, dst.OnPublishStart, src.OnPublishStart)
	dst.OnPublishDone = chainPublishDone(dst, dst.OnPublishDone, src.OnPublishDone)
	dst.OnHandlerError = chainHandlerError(dst, dst.OnHandlerError, src.OnHandlerError)
	dst.OnHandlerPanic = chainHandlerPanic(dst, dst.OnHandlerPanic, src.OnHandlerPanic)
	dst.OnEventDropped = chainDroppedEvent(dst, dst.OnEventDropped, src.OnEventDropped)
	dst.OnEventRejected = chainRejectedEvent(dst, dst.OnEventRejected, src.OnEventRejected)
}

func chainPublishStart(dst *Hooks, a, b func(PublishStart)) func(PublishStart) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return func(info PublishStart) {
			safeCall("OnPublishStart", hookPanicReporter(dst), func() { a(info) })
			safeCall("OnPublishStart", hookPanicReporter(dst), func() { b(info) })
		}
	}
}

func chainPublishDone(dst *Hooks, a, b func(PublishDone)) func(PublishDone) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return func(info PublishDone) {
			safeCall("OnPublishDone", hookPanicReporter(dst), func() { a(info) })
			safeCall("OnPublishDone", hookPanicReporter(dst), func() { b(info) })
		}
	}
}

func chainHandlerError(dst *Hooks, a, b func(HandlerError)) func(HandlerError) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return func(info HandlerError) {
			safeCall("OnHandlerError", hookPanicReporter(dst), func() { a(info) })
			safeCall("OnHandlerError", hookPanicReporter(dst), func() { b(info) })
		}
	}
}

func chainHandlerPanic(dst *Hooks, a, b func(HandlerPanic)) func(HandlerPanic) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return func(info HandlerPanic) {
			safeCall("OnHandlerPanic", hookPanicReporter(dst), func() { a(info) })
			safeCall("OnHandlerPanic", hookPanicReporter(dst), func() { b(info) })
		}
	}
}

func chainDroppedEvent(dst *Hooks, a, b func(DroppedEvent)) func(DroppedEvent) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return func(info DroppedEvent) {
			safeCall("OnEventDropped", hookPanicReporter(dst), func() { a(info) })
			safeCall("OnEventDropped", hookPanicReporter(dst), func() { b(info) })
		}
	}
}

func chainRejectedEvent(dst *Hooks, a, b func(RejectedEvent)) func(RejectedEvent) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return func(info RejectedEvent) {
			safeCall("OnEventRejected", hookPanicReporter(dst), func() { a(info) })
			safeCall("OnEventRejected", hookPanicReporter(dst), func() { b(info) })
		}
	}
}

func chainHookPanic(a, b func(HookPanic)) func(HookPanic) {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	default:
		return func(info HookPanic) {
			safeCall("OnHookPanic", nil, func() { a(info) })
			safeCall("OnHookPanic", nil, func() { b(info) })
		}
	}
}

func hookPanicReporter(hooks *Hooks) func(HookPanic) {
	if hooks == nil {
		return nil
	}

	return func(info HookPanic) {
		if hooks.OnHookPanic == nil {
			return
		}
		safeCall("OnHookPanic", nil, func() { hooks.OnHookPanic(info) })
	}
}

func safeCall(name string, report func(HookPanic), fn func()) {
	defer func() {
		if recovered := recover(); recovered != nil && report != nil {
			report(HookPanic{
				Hook:  name,
				Value: recovered,
			})
		}
	}()
	fn()
}
