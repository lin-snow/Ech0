// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package outbox

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/lin-snow/ech0/internal/event"
	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	outboxModel "github.com/lin-snow/ech0/internal/model/outbox"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/transaction"
	"github.com/lin-snow/ech0/pkg/busen"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeliveryIsCommittedAndRetriedWithStableID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "effects.db") + "?_busy_timeout=5000&_txlock=immediate"
	pool := func() *gorm.DB {
		db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
		require.NoError(t, err)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		t.Cleanup(func() { _ = sqlDB.Close() })
		return db
	}
	a, b := pool(), pool()
	require.NoError(t, a.AutoMigrate(&outboxModel.Delivery{}, &webhookModel.Webhook{}))
	hook := webhookModel.Webhook{ID: "hook", URL: "https://example.test/receiver", IsActive: true}
	require.NoError(t, a.Create(&hook).Error)
	var attempts atomic.Int32
	var ids []string
	send := func(_ context.Context, _ *webhookModel.Webhook, obs event.WebhookObservation, id string) error {
		require.Equal(t, "system.snapshot", obs.Topic)
		require.JSONEq(t, `{"Info":"test","Size":20}`, string(obs.Payload))
		ids = append(ids, id)
		if attempts.Add(1) == 1 {
			return errors.New("test transport timeout")
		}
		return nil
	}
	first := &Worker{db: func() *gorm.DB { return a }, send: send}
	second := &Worker{db: func() *gorm.DB { return b }, send: send}
	provider := eventbus.ProvideProvider()
	registrar := eventbus.NewEventRegistry(provider, []eventbus.Subscriber{first})
	require.NoError(t, registrar.Register())
	t.Cleanup(func() { _ = registrar.Stop() })
	tx := transaction.NewFunctionTransactor(first.db)
	ctx := t.Context()
	record := func(ctx context.Context, bus *busen.Bus) error {
		return eventbus.Record(ctx, bus, event.SystemSnapshot{Info: "test", Size: 20})
	}
	require.Error(t, tx.Run(ctx, func(txCtx context.Context) error {
		require.NoError(t, record(txCtx, provider()))
		return errors.New("rollback")
	}))
	ticket, err := second.Next(ctx)
	require.NoError(t, err)
	require.Nil(t, ticket)
	require.NoError(t, tx.Run(ctx, func(txCtx context.Context) error { return record(txCtx, provider()) }))
	ticket, err = second.Next(ctx)
	require.NoError(t, err)
	require.NotNil(t, ticket)
	require.NoError(t, second.Advance(ctx, *ticket))
	require.NoError(t, first.Advance(ctx, *ticket))
	require.EqualValues(t, 1, attempts.Load())
	retry, err := first.Next(ctx)
	require.NoError(t, err)
	require.Equal(t, ticket.ID, retry.ID)
	require.EqualValues(t, 1, retry.Attempt)
	require.NoError(t, b.Exec("UPDATE ech0_deliveries SET ready=0").Error)
	require.NoError(t, first.Advance(ctx, *retry))
	require.NoError(t, second.Advance(ctx, *retry))
	require.Equal(t, []string{ticket.ID, ticket.ID}, ids)
	require.NoError(t, b.First(&hook, "id=?", "hook").Error)
	require.Equal(t, "success", hook.LastStatus)
	ticket, err = second.Next(ctx)
	require.NoError(t, err)
	require.Nil(t, ticket)
}
