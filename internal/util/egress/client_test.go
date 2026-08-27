// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package egress

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestNewClient_Guard_blocks_loopback_server(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(Guard(), Timeout(3*time.Second))

	resp, err := client.Get(srv.URL)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected guarded client to block request to loopback server")
	}
	if !strings.Contains(err.Error(), "不被允许") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestNewClient_Guard_blocks_redirect_to_loopback(t *testing.T) {
	t.Parallel()

	loopbackSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer loopbackSrv.Close()

	redirectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, loopbackSrv.URL, http.StatusFound)
	}))
	defer redirectSrv.Close()

	client := NewClient(Guard(), Timeout(3*time.Second))

	resp, err := client.Get(redirectSrv.URL)
	if resp != nil {
		_ = resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected guarded client to block redirect to loopback")
	}
}

func TestRetry_stops_on_first_success(t *testing.T) {
	t.Parallel()

	calls := 0
	err := Retry(3, time.Millisecond, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetry_exhausts_attempts(t *testing.T) {
	t.Parallel()

	calls := 0
	sentinel := errors.New("boom")
	err := Retry(3, time.Millisecond, func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
}

func TestRetry_backoff_schedule(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const initial = 100 * time.Millisecond
		start := time.Now()

		sentinel := errors.New("boom")
		var offsets []time.Duration
		err := Retry(4, initial, func() error {
			offsets = append(offsets, time.Since(start))
			return sentinel
		})
		if !errors.Is(err, sentinel) {
			t.Fatalf("expected sentinel error, got %v", err)
		}

		want := []time.Duration{0, initial, 3 * initial, 7 * initial}
		if !slices.Equal(offsets, want) {
			t.Fatalf("attempt offsets = %v, want %v", offsets, want)
		}
		if elapsed := time.Since(start); elapsed != want[len(want)-1] {
			t.Fatalf("Retry slept after the final attempt: elapsed = %v, want %v", elapsed, want[len(want)-1])
		}
	})
}
