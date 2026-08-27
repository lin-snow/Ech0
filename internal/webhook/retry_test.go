// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fastBackoff = time.Nanosecond

type recordedRequest struct {
	headers http.Header
	body    []byte
}

type requestRecorder struct {
	mu       sync.Mutex
	requests []recordedRequest
}

func (r *requestRecorder) record(req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, recordedRequest{
		headers: req.Header.Clone(),
		body:    body,
	})
}

func (r *requestRecorder) snapshot() []recordedRequest {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]recordedRequest, len(r.requests))
	copy(out, r.requests)
	return out
}

func (r *requestRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.requests)
}

func TestSendWithRetry_Success2xx(t *testing.T) {
	cases := []struct {
		name   string
		status int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"202 Accepted", http.StatusAccepted},
		{"204 NoContent", http.StatusNoContent},
		{"299 edge", 299},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &requestRecorder{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				rec.record(req)
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			wh := &webhookModel.Webhook{URL: srv.URL, Secret: "s"}
			err := sendWithRetry(srv.Client(), wh, newObs(t), 3, fastBackoff)

			require.NoError(t, err, "2xx 必须视为成功")
			assert.Equal(t, 1, rec.count(), "成功时不应重试")
		})
	}
}

func TestSendWithRetry_ExhaustsOn5xx(t *testing.T) {
	cases := []struct {
		name       string
		status     int
		maxRetries int
	}{
		{"500 retries 3", http.StatusInternalServerError, 3},
		{"503 retries 2", http.StatusServiceUnavailable, 2},
		{"502 retries 5", http.StatusBadGateway, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := &requestRecorder{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				rec.record(req)
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			wh := &webhookModel.Webhook{URL: srv.URL, Secret: "s"}
			err := sendWithRetry(srv.Client(), wh, newObs(t), tc.maxRetries, fastBackoff)

			require.Error(t, err, "持续 5xx 必须最终报错")
			assert.Contains(t, err.Error(), "unexpected status code", "错误应描述非 2xx 状态码")
			assert.Equal(t, tc.maxRetries, rec.count(), "应恰好重试 maxRetries 次")
		})
	}
}

func TestSendWithRetry_RecoversAfterTransient(t *testing.T) {
	const failBefore = 2
	var hits atomic.Int32
	rec := &requestRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.record(req)
		if hits.Add(1) <= failBefore {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := &webhookModel.Webhook{URL: srv.URL, Secret: "s"}
	err := sendWithRetry(srv.Client(), wh, newObs(t), 5, fastBackoff)

	require.NoError(t, err, "瞬时失败后恢复应最终成功")
	assert.Equal(t, failBefore+1, rec.count(), "应在第一次成功后停止重试")
}

func TestSendWithRetry_SignatureEndToEnd(t *testing.T) {
	const secret = "end-to-end-secret"
	rec := &requestRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.record(req)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	obs := newObs(t)
	wh := &webhookModel.Webhook{URL: srv.URL, Secret: secret}
	err := sendWithRetry(srv.Client(), wh, obs, 3, fastBackoff)
	require.NoError(t, err)

	reqs := rec.snapshot()
	require.Len(t, reqs, 1)
	got := reqs[0]

	sig := got.headers.Get("X-Ech0-Signature")
	require.NotEmpty(t, sig, "带密钥时服务端必须收到签名头")
	require.True(t, strings.HasPrefix(sig, "sha256="), "签名头必须带 sha256= 前缀")

	want := "sha256=" + buildWebhookSignature(secret, got.body)
	assert.Equal(t, want, sig, "服务端收到的签名必须等于对收到 body 的 HMAC")

	assert.Equal(t, obs.Topic, got.headers.Get("X-Ech0-Event"))
	assert.Equal(t, "application/json", got.headers.Get("Content-Type"))
	assert.Equal(t, "Ech0-Webhook-Client", got.headers.Get("User-Agent"))
}

func TestSendWithRetry_NoSecretNoSignature(t *testing.T) {
	rec := &requestRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.record(req)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := &webhookModel.Webhook{URL: srv.URL, Secret: ""}
	require.NoError(t, sendWithRetry(srv.Client(), wh, newObs(t), 2, fastBackoff))

	reqs := rec.snapshot()
	require.Len(t, reqs, 1)
	assert.Empty(t, reqs[0].headers.Get("X-Ech0-Signature"), "无密钥不应携带签名头")
}

func TestSendWithRetry_BodyDeliveredEachAttempt(t *testing.T) {
	const secret = "k"
	var hits atomic.Int32
	rec := &requestRecorder{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec.record(req)
		if hits.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	obs := newObs(t)
	wh := &webhookModel.Webhook{URL: srv.URL, Secret: secret}
	require.NoError(t, sendWithRetry(srv.Client(), wh, obs, 3, fastBackoff))

	reqs := rec.snapshot()
	require.Len(t, reqs, 2, "应发生一次重试，共两次请求")

	for i, r := range reqs {
		require.NotEmpty(t, r.body, "第 %d 次请求 body 不应为空", i)
		assert.Contains(t, string(r.body), `"topic":"`+obs.Topic+`"`, "第 %d 次 body 应含 topic", i)
		want := "sha256=" + buildWebhookSignature(secret, r.body)
		assert.Equal(t, want, r.headers.Get("X-Ech0-Signature"), "第 %d 次签名应与该次 body 一致", i)
	}
	assert.Equal(t, reqs[0].body, reqs[1].body, "重试发送的 body 必须与首次一致")
}

type countingErrTransport struct {
	calls *int32
}

func (c countingErrTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	atomic.AddInt32(c.calls, 1)
	return nil, errors.New("simulated connection failure")
}

func TestSendWithRetry_TransportErrorRetried(t *testing.T) {
	var calls int32
	client := &http.Client{Transport: countingErrTransport{calls: &calls}}
	wh := &webhookModel.Webhook{URL: "https://example.com/hook", Secret: "s"}

	const maxRetries = 4
	err := sendWithRetry(client, wh, newObs(t), maxRetries, fastBackoff)

	require.Error(t, err, "传输错误应最终冒泡")
	assert.Contains(t, err.Error(), "simulated connection failure")
	assert.Equal(t, int32(maxRetries), atomic.LoadInt32(&calls), "传输错误也应重试 maxRetries 次")
}

func TestSendWithRetry_BuildRequestError(t *testing.T) {
	var calls int32
	client := &http.Client{Transport: countingErrTransport{calls: &calls}}
	wh := &webhookModel.Webhook{URL: "://bad-url", Secret: "s"}

	err := sendWithRetry(client, wh, newObs(t), 3, fastBackoff)

	require.Error(t, err, "非法 URL 必须导致 buildRequest 报错")
	assert.Zero(t, atomic.LoadInt32(&calls), "buildRequest 失败时不应触达 transport")
}

func TestSendWithRetry_BackoffScheduleIsExponential(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const initialBackoff = time.Second
		const maxRetries = 3

		start := time.Now()
		var mu sync.Mutex
		var offsets []time.Duration
		srv := httptest.NewTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			mu.Lock()
			offsets = append(offsets, time.Since(start))
			mu.Unlock()
			w.WriteHeader(http.StatusInternalServerError)
		}))

		client := srv.Client()
		wh := &webhookModel.Webhook{URL: srv.URL, Secret: "s"}
		err := sendWithRetry(client, wh, newObs(t), maxRetries, initialBackoff)
		require.Error(t, err, "持续 5xx 必须最终报错")

		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t,
			[]time.Duration{0, initialBackoff, 3 * initialBackoff},
			offsets,
			"两次重试之间的等待必须逐次翻倍，且最后一次尝试后不再等待",
		)
		assert.Equal(t, 3*initialBackoff, time.Since(start), "最后一次尝试之后不应再 sleep")
	})
}
