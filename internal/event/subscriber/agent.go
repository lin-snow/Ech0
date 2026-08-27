// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package subscriber

import (
	"context"

	"github.com/lin-snow/ech0/internal/agent"
	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/kvstore"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
)

type AgentProcessor struct {
	durableKV kvstore.Store
}

func NewAgentProcessor(
	durableKV kvstore.Store,
) *AgentProcessor {
	return &AgentProcessor{
		durableKV: durableKV,
	}
}

func (ap *AgentProcessor) handle(ctx context.Context) error {
	if _, err := coreSetting.Get(ctx, ap.durableKV, coreSetting.Agent); err != nil {
		return err
	}
	return ap.clearCache()
}

func (ap *AgentProcessor) HandleEchoCreated(ctx context.Context, e event.EchoCreated) error {
	_ = e
	return ap.handle(ctx)
}

func (ap *AgentProcessor) HandleEchoUpdated(ctx context.Context, e event.EchoUpdated) error {
	_ = e
	return ap.handle(ctx)
}

func (ap *AgentProcessor) HandleUserDeleted(ctx context.Context, e event.UserDeleted) error {
	_ = e
	return ap.handle(ctx)
}

func (ap *AgentProcessor) clearCache() error {
	return ap.durableKV.Delete(context.Background(), string(agent.GEN_RECENT))
}

func (ap *AgentProcessor) Registrations() []eventbus.Registration {
	return []eventbus.Registration{
		eventbus.On(ap.HandleEchoCreated, eventbus.AsyncParallel()...),
		eventbus.On(ap.HandleEchoUpdated, eventbus.AsyncParallel()...),
		eventbus.On(ap.HandleUserDeleted, eventbus.AsyncParallel()...),
	}
}
