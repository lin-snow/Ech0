// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package agent

import "context"

type textPipeline struct {
	splitter reasoningSplitter
	guard    toolCallLeakGuard
}

func (t *textPipeline) feed(ctx context.Context, ch chan<- Event, delta string) bool {
	answer, reasoning := t.splitter.feed(delta)
	if reasoning != "" && !send(ctx, ch, Event{Kind: EventReasoningDelta, Text: reasoning}) {
		return false
	}
	return t.emitAnswer(ctx, ch, answer)
}

func (t *textPipeline) flush(ctx context.Context, ch chan<- Event) bool {
	answer, reasoning := t.splitter.flush()
	if reasoning != "" && !send(ctx, ch, Event{Kind: EventReasoningDelta, Text: reasoning}) {
		return false
	}
	if !t.emitAnswer(ctx, ch, answer) {
		return false
	}
	if rest := t.guard.flush(); rest != "" {
		return send(ctx, ch, Event{Kind: EventTextDelta, Text: rest})
	}
	return true
}

func (t *textPipeline) emitAnswer(ctx context.Context, ch chan<- Event, answer string) bool {
	if answer == "" {
		return true
	}
	safe, tripped := t.guard.feed(answer)
	if tripped {
		send(ctx, ch, Event{Kind: EventError, Err: errTextToolCallLeak})
		return false
	}
	if safe == "" {
		return true
	}
	return send(ctx, ch, Event{Kind: EventTextDelta, Text: safe})
}
