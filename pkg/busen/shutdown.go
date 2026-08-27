// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package busen

import (
	"context"
	"fmt"
)

type ShutdownMode int

const (
	ShutdownDrain ShutdownMode = iota
	ShutdownBestEffort
	ShutdownAbort
)

type ShutdownResult struct {
	Mode                ShutdownMode
	Processed           int64
	Dropped             int64
	Rejected            int64
	TimedOutSubscribers []uint64
	Completed           bool
}

func (b *Bus) Shutdown(ctx context.Context, mode ShutdownMode) (ShutdownResult, error) {
	result := ShutdownResult{Mode: mode}

	if b == nil {
		return result, fmt.Errorf("%w: nil bus", ErrInvalidOption)
	}
	if !mode.valid() {
		return result, fmt.Errorf("%w: unknown shutdown mode", ErrInvalidOption)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	before := snapshotSubscriptionStats(b.allSubscriptions())
	b.gate.Close()
	subs := b.allSubscriptions()

	if err := b.gate.Wait(ctx); err != nil {
		result.Completed = false
		if mode == ShutdownDrain {
			return result, closeWaitError(err)
		}
		result.TimedOutSubscribers = allSubscriptionIDs(subs)
		applyStatsDelta(&result, before, snapshotSubscriptionStats(subs))
		return result, nil
	}

	for _, sub := range subs {
		sub.stopAccepting()
	}
	if mode == ShutdownAbort {
		for _, sub := range subs {
			sub.abortPending()
		}
	}
	for _, sub := range subs {
		sub.scheduleStop()
	}

	for _, sub := range subs {
		if err := sub.waitStopped(ctx); err != nil {
			result.TimedOutSubscribers = append(result.TimedOutSubscribers, sub.id)
			if mode == ShutdownDrain {
				applyStatsDelta(&result, before, snapshotSubscriptionStats(subs))
				result.Completed = false
				return result, closeWaitError(err)
			}
		}
	}

	applyStatsDelta(&result, before, snapshotSubscriptionStats(subs))
	result.Completed = len(result.TimedOutSubscribers) == 0
	return result, nil
}

func snapshotSubscriptionStats(subs []*subscription) map[uint64]subscriptionStats {
	stats := make(map[uint64]subscriptionStats, len(subs))
	for _, sub := range subs {
		stats[sub.id] = sub.snapshotStats()
	}
	return stats
}

func applyStatsDelta(result *ShutdownResult, before, after map[uint64]subscriptionStats) {
	if result == nil {
		return
	}
	for id, end := range after {
		start, ok := before[id]
		if !ok {
			start = subscriptionStats{}
		}
		result.Processed += end.processed - start.processed
		result.Dropped += (end.dropped - start.dropped) + (end.shutdownDrop - start.shutdownDrop)
		result.Rejected += end.rejected - start.rejected
	}
}

func allSubscriptionIDs(subs []*subscription) []uint64 {
	if len(subs) == 0 {
		return nil
	}
	ids := make([]uint64, 0, len(subs))
	for _, sub := range subs {
		ids = append(ids, sub.id)
	}
	return ids
}

func (m ShutdownMode) valid() bool {
	return m >= ShutdownDrain && m <= ShutdownAbort
}
