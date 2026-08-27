// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package subscriber

import (
	"context"
	"time"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	embeddingService "github.com/lin-snow/ech0/internal/service/embedding"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type EmbeddingProcessor struct {
	indexer embeddingService.Indexer
}

func NewEmbeddingProcessor(indexer embeddingService.Indexer) *EmbeddingProcessor {
	return &EmbeddingProcessor{indexer: indexer}
}

func (ep *EmbeddingProcessor) HandleEchoCreated(ctx context.Context, e event.EchoCreated) error {
	return ep.withRetry(func() error { return ep.indexer.IndexEcho(ctx, e.Echo) })
}

func (ep *EmbeddingProcessor) HandleEchoUpdated(ctx context.Context, e event.EchoUpdated) error {
	return ep.withRetry(func() error { return ep.indexer.IndexEcho(ctx, e.Echo) })
}

func (ep *EmbeddingProcessor) HandleEchoDeleted(ctx context.Context, e event.EchoDeleted) error {
	return ep.indexer.RemoveEcho(ctx, e.Echo.ID)
}

func (ep *EmbeddingProcessor) withRetry(fn func() error) error {
	var err error
	for attempt := range 3 {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	logUtil.GetLogger().Warn("embedding index failed after retries", logUtil.Err(err))
	return err
}

func (ep *EmbeddingProcessor) Registrations() []eventbus.Registration {
	return []eventbus.Registration{
		eventbus.On(ep.HandleEchoCreated, eventbus.AsyncParallel()...),
		eventbus.On(ep.HandleEchoUpdated, eventbus.AsyncParallel()...),
		eventbus.On(ep.HandleEchoDeleted, eventbus.AsyncParallel()...),
	}
}
