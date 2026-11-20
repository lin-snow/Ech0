// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lin-snow/ech0/internal/event"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/pkg/egress"
)

const (
	defaultWebhookTimeout = 5 * time.Second

	deliverMaxRetries = 3
	deliverBackoff    = 500 * time.Millisecond
	testMaxRetries    = 2
	testBackoff       = 300 * time.Millisecond
)

type Sender struct {
	client *http.Client
}

func NewSender() *Sender {
	return &Sender{
		client: egress.NewClient(egress.Guard(), egress.Timeout(defaultWebhookTimeout)),
	}
}

func (s *Sender) Deliver(wh *webhookModel.Webhook, obs event.WebhookObservation) error {
	return sendWithRetry(s.client, wh, obs, deliverMaxRetries, deliverBackoff)
}

func (s *Sender) SendTest(wh *webhookModel.Webhook) error {
	obs, err := event.NewWebhookObservation("webhook.test", map[string]any{
		"message": "webhook connectivity test from ech0",
		"webhook": wh.Name,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}, map[string]string{"source": "setting.test"})
	if err != nil {
		return err
	}
	return sendWithRetry(s.client, wh, obs, testMaxRetries, testBackoff)
}

// DeliverOnce lets the durable worker own retries and retain the same event ID.
func (s *Sender) DeliverOnce(ctx context.Context, hook *webhookModel.Webhook, observation event.WebhookObservation, id string) error {
	req, err := buildRequest(hook, observation)
	if err != nil {
		return err
	}
	req.Header.Set("X-Ech0-Event-ID", id)
	response, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("webhook returned HTTP %d", response.StatusCode)
	}
	return nil
}
