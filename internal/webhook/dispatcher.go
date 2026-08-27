// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"context"
	"log/slog"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/event"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	asyncUtil "github.com/lin-snow/ech0/internal/util/async"
	logUtil "github.com/lin-snow/ech0/pkg/log"
)

type WebhookStore interface {
	ListActiveWebhooks(ctx context.Context) ([]webhookModel.Webhook, error)
	UpdateWebhookDeliveryStatus(
		ctx context.Context,
		id string,
		status string,
		lastTrigger int64,
	) error
}

type Dispatcher struct {
	sender *Sender
	repo   WebhookStore
	pool   *asyncUtil.WorkerPool
}

func NewDispatcher(repo WebhookStore) *Dispatcher {
	return &Dispatcher{
		repo:   repo,
		sender: NewSender(),
		pool: asyncUtil.NewWorkerPool(
			config.Config().Event.WebhookPoolWorkers,
			config.Config().Event.WebhookPoolQueue,
		),
	}
}

func (wd *Dispatcher) HandleObservation(ctx context.Context, obs event.WebhookObservation) error {
	webhooks, err := wd.repo.ListActiveWebhooks(ctx)
	if err != nil {
		return err
	}
	for _, wh := range webhooks {
		wd.pool.Submit(func() error {
			wd.Dispatch(ctx, &wh, obs)
			return nil
		})
	}

	return nil
}

func (wd *Dispatcher) Dispatch(ctx context.Context, wh *webhookModel.Webhook, obs event.WebhookObservation) {
	triggerAt := time.Now().UTC().Unix()
	if err := wd.sender.Deliver(wh, obs); err != nil {
		wd.updateWebhookStatus(ctx, wh.ID, "failed", triggerAt)
		logUtil.GetLogger().Error("Webhook Handle Failed", slog.String("name", wh.Name), slog.String("url", wh.URL), logUtil.Err(err))
		return
	}
	wd.updateWebhookStatus(ctx, wh.ID, "success", triggerAt)
}

func (wd *Dispatcher) Wait() {
	wd.pool.Wait()
}

func (wd *Dispatcher) Stop() {
	wd.pool.Stop()
}

func (wd *Dispatcher) updateWebhookStatus(
	ctx context.Context,
	webhookID string,
	status string,
	triggerAt int64,
) {
	if webhookID == "" {
		return
	}
	if err := wd.repo.UpdateWebhookDeliveryStatus(ctx, webhookID, status, triggerAt); err != nil {
		logUtil.GetLogger().Warn(
			"update webhook delivery status failed",
			slog.String("webhook_id", webhookID),
			slog.String("status", status),
			logUtil.Err(err),
		)
	}
}
