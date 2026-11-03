// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"context"
	"errors"

	authzModel "github.com/lin-snow/ech0/internal/model/authz"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

// ErrClientNotFound is returned when a dynamically registered client id is
// unknown; callers map it to OAuth's invalid_client.
var ErrClientNotFound = errors.New("authz: client not found")

type ClientRepository struct {
	db func() *gorm.DB
}

func NewClientRepository(dbProvider func() *gorm.DB) *ClientRepository {
	return &ClientRepository{db: dbProvider}
}

func (r *ClientRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *ClientRepository) FindClient(ctx context.Context, clientID string) (authzModel.Client, error) {
	var client authzModel.Client
	err := r.getDB(ctx).WithContext(ctx).Where("client_id = ?", clientID).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return authzModel.Client{}, ErrClientNotFound
	}
	return client, err
}

func (r *ClientRepository) CreateClient(ctx context.Context, client *authzModel.Client) error {
	return r.getDB(ctx).WithContext(ctx).Create(client).Error
}

func (r *ClientRepository) CountClients(ctx context.Context) (int64, error) {
	var count int64
	err := r.getDB(ctx).WithContext(ctx).Model(&authzModel.Client{}).Count(&count).Error
	return count, err
}
