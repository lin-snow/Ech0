// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/storage"
	"golang.org/x/sync/singleflight"
)

type CopilotService struct {
	echoService    EchoService
	embedding      EmbeddingService
	userReader     UserReader
	durableKV      kvstore.Store
	storage        *storage.Manager
	recentGenGroup singleflight.Group
}

var (
	_ SummaryService = (*CopilotService)(nil)
	_ ChatService    = (*CopilotService)(nil)
)

type chatUser struct {
	ID       string
	Username string
}

func NewCopilotService(
	echoService EchoService,
	embedding EmbeddingService,
	userReader UserReader,
	durableKV kvstore.Store,
	storageManager *storage.Manager,
) *CopilotService {
	return &CopilotService{
		echoService: echoService,
		embedding:   embedding,
		userReader:  userReader,
		durableKV:   durableKV,
		storage:     storageManager,
	}
}
