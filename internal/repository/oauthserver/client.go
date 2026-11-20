// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"context"
	"errors"

	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrClientNotFound is returned when a dynamically registered client id is
// unknown; callers map it to OAuth's invalid_client.
var ErrClientNotFound = errors.New("oauthserver: client not found")

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

func (r *ClientRepository) FindClient(ctx context.Context, clientID string) (oauthserverModel.Client, error) {
	var client oauthserverModel.Client
	err := r.getDB(ctx).WithContext(ctx).Where("client_id = ?", clientID).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return oauthserverModel.Client{}, ErrClientNotFound
	}
	return client, err
}

func (r *ClientRepository) CreateClient(ctx context.Context, client *oauthserverModel.Client) error {
	return r.getDB(ctx).WithContext(ctx).Create(client).Error
}

func (r *ClientRepository) CountClients(ctx context.Context) (int64, error) {
	var count int64
	err := r.getDB(ctx).WithContext(ctx).Model(&oauthserverModel.Client{}).Count(&count).Error
	return count, err
}

func (r *ClientRepository) ConsumeCode(ctx context.Context, jti string, now, expiresAt int64) (bool, error) {
	var inserted bool
	err := r.getDB(ctx).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("expires_at < ?", now).Delete(&oauthserverModel.ConsumedCode{}).Error; err != nil {
			return err
		}
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&oauthserverModel.ConsumedCode{JTI: jti, ExpiresAt: expiresAt})
		inserted = res.RowsAffected == 1
		return res.Error
	})
	return inserted && err == nil, err
}
