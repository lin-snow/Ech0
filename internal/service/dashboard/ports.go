// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"net/http"

	"github.com/lin-snow/ech0/internal/visitor"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type SystemLogQuery struct {
	Tail    int
	Level   string
	Keyword string
}

type SystemLogStreamFilter struct {
	Level   string
	Keyword string
}

type Service interface {
	GetSystemLogs(query SystemLogQuery) ([]logUtil.LogEntry, error)
	GetVisitorStats() []visitor.DayStat
	WSSubscribeSystemLogs(w http.ResponseWriter, r *http.Request, filter SystemLogStreamFilter) error
	SSESubscribeSystemLogs(w http.ResponseWriter, r *http.Request, filter SystemLogStreamFilter) error
}
