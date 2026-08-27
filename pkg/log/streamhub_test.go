// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package log

import "testing"

func msgsOf(entries []LogEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Msg)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func entry(msg string) LogEntry {
	return LogEntry{Level: "info", Msg: msg, Raw: msg}
}

func TestNewLogStreamHubDefaults(t *testing.T) {
	h := newLogStreamHub(0, 0, "drop_oldest")
	if h.recentCap != 2000 {
		t.Errorf("recentCap = %d, want 2000 (default)", h.recentCap)
	}
	if h.defaultSubBuffer != 2048 {
		t.Errorf("defaultSubBuffer = %d, want 2048 (default)", h.defaultSubBuffer)
	}
	if len(h.recent) != 2000 {
		t.Errorf("len(recent) = %d, want 2000", len(h.recent))
	}
}

func TestLogStreamHubRecentRingOrder(t *testing.T) {
	h := newLogStreamHub(8, 3, "drop_oldest")
	for _, m := range []string{"A", "B", "C", "D", "E"} {
		h.Publish(entry(m))
	}

	t.Run("recent returns most recent in insertion order", func(t *testing.T) {
		got := msgsOf(h.Recent(0))
		want := []string{"C", "D", "E"}
		if !equalStrings(got, want) {
			t.Errorf("Recent(0) = %v, want %v", got, want)
		}
	})

	t.Run("recent honors limit", func(t *testing.T) {
		got := msgsOf(h.Recent(2))
		want := []string{"D", "E"}
		if !equalStrings(got, want) {
			t.Errorf("Recent(2) = %v, want %v", got, want)
		}
	})

	t.Run("recent limit larger than buffer clamps", func(t *testing.T) {
		got := msgsOf(h.Recent(100))
		want := []string{"C", "D", "E"}
		if !equalStrings(got, want) {
			t.Errorf("Recent(100) = %v, want %v", got, want)
		}
	})
}

func TestLogStreamHubRecentEmpty(t *testing.T) {
	h := newLogStreamHub(8, 4, "drop_oldest")
	if got := h.Recent(10); got != nil {
		t.Errorf("Recent on empty hub = %v, want nil", got)
	}
}

func TestLogStreamHubSubscribePublish(t *testing.T) {
	h := newLogStreamHub(8, 10, "drop_oldest")

	t.Run("subscribe buffer size honored", func(t *testing.T) {
		_, ch, cancel := h.Subscribe(4)
		defer cancel()
		if cap(ch) != 4 {
			t.Errorf("cap(ch) = %d, want 4", cap(ch))
		}
	})

	t.Run("subscribe default buffer when non-positive", func(t *testing.T) {
		_, ch, cancel := h.Subscribe(0)
		defer cancel()
		if cap(ch) != 8 {
			t.Errorf("cap(ch) = %d, want 8 (default sub buffer)", cap(ch))
		}
	})

	t.Run("publish delivers to subscriber", func(t *testing.T) {
		_, ch, cancel := h.Subscribe(4)
		defer cancel()
		h.Publish(entry("hello"))
		select {
		case got := <-ch:
			if got.Msg != "hello" {
				t.Errorf("received Msg = %q, want %q", got.Msg, "hello")
			}
		default:
			t.Fatal("expected an entry on the subscriber channel")
		}
	})
}

func TestLogStreamHubCancelRemovesSubscriber(t *testing.T) {
	h := newLogStreamHub(8, 10, "drop_oldest")
	id, ch, cancel := h.Subscribe(4)

	cancel()

	h.mu.RLock()
	_, present := h.subs[id]
	h.mu.RUnlock()
	if present {
		t.Errorf("subscriber %d still present after cancel", id)
	}

	if _, ok := <-ch; ok {
		t.Error("channel should be closed after cancel")
	}

	cancel()

	h.Publish(entry("ignored"))
}

func TestLogStreamHubDropNewest(t *testing.T) {
	h := newLogStreamHub(8, 10, "drop_newest")
	_, ch, cancel := h.Subscribe(1)
	defer cancel()

	h.Publish(entry("first"))
	h.Publish(entry("second"))

	if got := h.dropped.Load(); got != 1 {
		t.Errorf("dropped = %d, want 1", got)
	}

	select {
	case got := <-ch:
		if got.Msg != "first" {
			t.Errorf("kept Msg = %q, want %q (newest dropped)", got.Msg, "first")
		}
	default:
		t.Fatal("expected the first entry to remain buffered")
	}
}

func TestLogStreamHubDropOldest(t *testing.T) {
	h := newLogStreamHub(8, 10, "drop_oldest")
	_, ch, cancel := h.Subscribe(1)
	defer cancel()

	h.Publish(entry("first"))
	h.Publish(entry("second"))

	select {
	case got := <-ch:
		if got.Msg != "second" {
			t.Errorf("kept Msg = %q, want %q (oldest dropped)", got.Msg, "second")
		}
	default:
		t.Fatal("expected the second entry to remain buffered")
	}
}

func TestLogStreamHubClose(t *testing.T) {
	h := newLogStreamHub(8, 10, "drop_oldest")
	_, ch, _ := h.Subscribe(4)

	h.Close()

	if _, ok := <-ch; ok {
		t.Error("subscriber channel should be closed after Close")
	}

	h.Close()

	h.Publish(entry("ignored"))
	if got := h.Recent(10); got != nil {
		t.Errorf("Recent after close with no prior entries = %v, want nil", got)
	}
}

func TestLogStreamHubSubscribeAfterClose(t *testing.T) {
	h := newLogStreamHub(8, 10, "drop_oldest")
	h.Close()

	id, ch, cancel := h.Subscribe(4)
	if id != 0 {
		t.Errorf("id = %d, want 0 on closed hub", id)
	}
	if _, ok := <-ch; ok {
		t.Error("channel from closed hub should be closed")
	}
	cancel()
}
