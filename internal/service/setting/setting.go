// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"github.com/lin-snow/ech0/internal/kvstore"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
	webhookclient "github.com/lin-snow/ech0/internal/webhook"
	"github.com/lin-snow/ech0/pkg/busen"
)

type SettingService struct {
	transactor        transaction.Transactor
	commonService     CommonService
	fileService       FileService
	storageManager    *storage.Manager
	durableKV         kvstore.Store
	settingRepository SettingRepository
	webhookRepository WebhookRepository
	webhookSender     *webhookclient.Sender
	tokenRevoker      TokenRevoker
	bus               *busen.Bus
}

func NewSettingService(
	tx transaction.Transactor,
	commonService CommonService,
	fileService FileService,
	storageManager *storage.Manager,
	durableKV kvstore.Store,
	settingRepository SettingRepository,
	webhookRepository WebhookRepository,
	webhookSender *webhookclient.Sender,
	tokenRevoker TokenRevoker,
	busProvider func() *busen.Bus,
) *SettingService {
	return &SettingService{
		transactor:        tx,
		commonService:     commonService,
		fileService:       fileService,
		storageManager:    storageManager,
		durableKV:         durableKV,
		webhookRepository: webhookRepository,
		webhookSender:     webhookSender,
		settingRepository: settingRepository,
		tokenRevoker:      tokenRevoker,
		bus:               busProvider(),
	}
}
