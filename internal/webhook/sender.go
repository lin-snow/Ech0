// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package webhook

import (
	"net/http"
	"time"

	"github.com/lin-snow/ech0/internal/event"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/util/egress"
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
