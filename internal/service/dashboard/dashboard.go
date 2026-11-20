// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lin-snow/ech0/internal/journal"
	"github.com/lin-snow/ech0/internal/serverless/platform"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lin-snow/ech0/internal/visitor"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type DashboardService struct {
	db             func() *gorm.DB
	visitorTracker *visitor.Tracker
}

func NewDashboardService(visitorTracker *visitor.Tracker, db func() *gorm.DB) *DashboardService {
	return &DashboardService{visitorTracker: visitorTracker, db: db}
}

func (s *DashboardService) GetSystemLogs(ctx context.Context, query SystemLogQuery) ([]logUtil.LogEntry, error) {
	tail := query.Tail
	if tail <= 0 {
		tail = 200
	}
	if platform.IsFunction(ctx) {
		return journal.Tail(ctx, s.db(), tail, query.Level, query.Keyword)
	}
	return logUtil.QueryLogFileTail(logUtil.CurrentLogFilePath(), tail, query.Level, query.Keyword)
}

func (s *DashboardService) GetVisitorStats(ctx context.Context) ([]visitor.DayStat, error) {
	return s.visitorTracker.Stats(ctx)
}

func (s *DashboardService) WSSubscribeSystemLogs(
	w http.ResponseWriter,
	r *http.Request,
	filter SystemLogStreamFilter,
) error {
	if platform.IsFunction(r.Context()) {
		after, err := journal.Latest(r.Context(), s.db())
		if err != nil {
			return err
		}
		upgrader := websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return err
		}
		defer func() { _ = conn.Close() }()
		ctx, stop := context.WithCancel(r.Context())
		defer stop()
		conn.SetReadLimit(4096)
		go func() {
			defer stop()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
		return s.followLogs(ctx, after, filter, func(_ int64, entry logUtil.LogEntry) error {
			if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
				return err
			}
			return conn.WriteJSON(streamMessage(entry))
		}, func() error { return conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)) })
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	_, stream, cancel := logUtil.SubscribeLogs(256)
	level := strings.ToLower(strings.TrimSpace(filter.Level))
	keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))

	go func() {
		defer cancel()
		defer func() { _ = conn.Close() }()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				if _, _, readErr := conn.ReadMessage(); readErr != nil {
					return
				}
			}
		}()

		for {
			select {
			case <-done:
				return
			case entry, ok := <-stream:
				if !ok {
					return
				}
				if !matchesSystemLogFilter(entry, level, keyword) {
					continue
				}
				resp := struct {
					Code int              `json:"code"`
					Msg  string           `json:"msg"`
					Data logUtil.LogEntry `json:"data"`
				}{
					Code: 1,
					Msg:  "system log update",
					Data: entry,
				}
				payload, _ := json.Marshal(resp)
				if writeErr := conn.WriteMessage(websocket.TextMessage, payload); writeErr != nil {
					return
				}
			}
		}
	}()
	return nil
}

func (s *DashboardService) SSESubscribeSystemLogs(
	w http.ResponseWriter,
	r *http.Request,
	filter SystemLogStreamFilter,
) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	headers := w.Header()
	headers.Set("Content-Type", "text/event-stream")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Connection", "keep-alive")
	headers.Set("X-Accel-Buffering", "no")

	if platform.IsFunction(r.Context()) {
		after, err := journal.Latest(r.Context(), s.db())
		if err != nil {
			return err
		}
		if id, err := strconv.ParseInt(r.Header.Get("Last-Event-ID"), 10, 64); err == nil && id >= 0 && id <= after {
			after = id
		}
		w.WriteHeader(http.StatusOK)
		flusher.Flush()
		return s.followLogs(r.Context(), after, filter, func(id int64, entry logUtil.LogEntry) error {
			body, err := json.Marshal(streamMessage(entry))
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "id: %d\ndata: %s\n\n", id, body); err != nil {
				return err
			}
			flusher.Flush()
			return nil
		}, func() error { _, err := fmt.Fprint(w, ": keep-alive\n\n"); flusher.Flush(); return err })
	}
	_, stream, cancel := logUtil.SubscribeLogs(512)
	defer cancel()

	level := strings.ToLower(strings.TrimSpace(filter.Level))
	keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))
	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return nil
		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		case entry, ok := <-stream:
			if !ok {
				return nil
			}
			if !matchesSystemLogFilter(entry, level, keyword) {
				continue
			}
			resp := struct {
				Code int              `json:"code"`
				Msg  string           `json:"msg"`
				Data logUtil.LogEntry `json:"data"`
			}{
				Code: 1,
				Msg:  "system log update",
				Data: entry,
			}
			payload, _ := json.Marshal(resp)
			_, _ = fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

func matchesSystemLogFilter(entry logUtil.LogEntry, level string, keyword string) bool {
	if level != "" && level != "all" && strings.ToLower(entry.Level) != level {
		return false
	}
	if keyword == "" {
		return true
	}
	target := strings.ToLower(entry.Msg + " " + entry.Raw)
	return strings.Contains(target, keyword)
}

func streamMessage(entry logUtil.LogEntry) any {
	return struct {
		Code int              `json:"code"`
		Msg  string           `json:"msg"`
		Data logUtil.LogEntry `json:"data"`
	}{1, "system log update", entry}
}
func (s *DashboardService) followLogs(ctx context.Context, after int64, filter SystemLogStreamFilter, send func(int64, logUtil.LogEntry) error, heartbeat func() error) error {
	level := strings.ToLower(strings.TrimSpace(filter.Level))
	keyword := strings.ToLower(strings.TrimSpace(filter.Keyword))
	err := journal.Follow(ctx, s.db(), after, func(id int64, entry logUtil.LogEntry) error {
		if matchesSystemLogFilter(entry, level, keyword) {
			return send(id, entry)
		}
		return nil
	}, heartbeat)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
