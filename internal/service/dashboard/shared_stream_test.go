// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	journalModel "github.com/lin-snow/ech0/internal/model/journal"
	"github.com/lin-snow/ech0/internal/serverless/platform"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWebSocketFollowsSharedLog(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "log.db")))
	require.NoError(t, err)
	p, err := db.DB()
	require.NoError(t, err)
	defer func() { _ = p.Close() }()
	require.NoError(t, db.AutoMigrate(&journalModel.Entry{}))
	service := NewDashboardService(nil, func() *gorm.DB { return db })
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = service.WSSubscribeSystemLogs(w, r.WithContext(platform.WithRequest(r.Context(), r)), SystemLogStreamFilter{Level: "error"})
	}))
	defer endpoint.Close()
	url := "ws" + strings.TrimPrefix(endpoint.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()
	require.NoError(t, db.Create(&journalModel.Entry{Source: "other-instance", Seq: 1, Raw: `{"level":"error","msg":"event"}`}).Error)
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))
	var message struct {
		Code int
		Data logUtil.LogEntry
	}
	require.NoError(t, conn.ReadJSON(&message))
	require.Equal(t, 1, message.Code)
	require.Equal(t, "event", message.Data.Msg)
}
