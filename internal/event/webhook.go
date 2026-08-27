// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package event

import (
	"encoding/json"
	"reflect"
	"time"
)

type WebhookObservation struct {
	Topic      string            `json:"topic"`
	EventName  string            `json:"event_name"`
	Payload    json.RawMessage   `json:"payload"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	OccurredAt int64             `json:"occurred_at"`
}

func NewWebhookObservation(topic string, payload any, metadata map[string]string) (WebhookObservation, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return WebhookObservation{}, err
	}
	return WebhookObservation{
		Topic:      topic,
		EventName:  eventNameOf(payload),
		Payload:    raw,
		Metadata:   metadata,
		OccurredAt: time.Now().UTC().Unix(),
	}, nil
}

func eventNameOf(payload any) string {
	if payload == nil {
		return ""
	}
	t := reflect.TypeOf(payload)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}
