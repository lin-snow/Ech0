// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import "maps"

type Event[T any] struct {
	Topic   string
	Key     string
	Value   T
	Headers map[string]string
	Meta    map[string]string
}

type envelope struct {
	topic   string
	key     string
	value   any
	headers map[string]string
	meta    map[string]string
}

func typedEvent[T any](e envelope) Event[T] {
	return Event[T]{
		Topic:   e.topic,
		Key:     e.key,
		Value:   e.value.(T),
		Headers: cloneHeaders(e.headers),
		Meta:    cloneHeaders(e.meta),
	}
}

func cloneHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(headers))
	maps.Copy(cloned, headers)
	return cloned
}
