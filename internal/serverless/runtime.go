// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package serverless

import (
	"context"
	"errors"

	eventbus "github.com/lin-snow/ech0/internal/event/bus"
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/server"
	"github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
)

// Runtime shares the same graph as the local HTTP server. The platform owns
// the listener and Vercel Queues owns the scheduling clock.
type Runtime struct{ Server *server.Server }

func New(ctx context.Context, httpServer *server.Server, registrar *eventbus.EventRegistrar, tx transaction.Transactor, store kvstore.Store, objects *storage.Manager) (*Runtime, error) {
	if !objects.GetSelector().ObjectEnabled() {
		return nil, errors.New("object storage is unavailable")
	}
	// Acquire a shared-database write lock before the upstream check/seed
	// sequence, so concurrent cold starts cannot overwrite a setting.
	err := tx.Run(ctx, func(txCtx context.Context) error {
		conn, _ := transaction.TxFromContext(txCtx)
		if err := conn.WithContext(txCtx).Exec("UPDATE ech0_schema SET version=version WHERE id=1").Error; err != nil {
			return err
		}
		return setting.Seed(txCtx, store)
	})
	if err != nil {
		return nil, err
	}
	if err := registrar.Register(); err != nil {
		return nil, err
	}
	return &Runtime{Server: httpServer}, nil
}
