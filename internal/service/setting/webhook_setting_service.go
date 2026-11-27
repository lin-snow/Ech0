// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"time"

	"github.com/lin-snow/ech0/internal/apperr"
	"github.com/lin-snow/ech0/internal/authz"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/egress"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error) {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return nil, err
	}

	webhooks, err := settingService.webhookRepository.GetAllWebhooks(ctx)
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (settingService *SettingService) DeleteWebhook(ctx context.Context, id string) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	return settingService.transactor.Run(ctx, func(txCtx context.Context) error {
		return settingService.webhookRepository.DeleteWebhookByID(txCtx, id)
	})
}

func (settingService *SettingService) UpdateWebhook(
	ctx context.Context,
	id string,
	newWebhook *model.WebhookDto,
) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	newWebhook.URL = urlUtil.TrimURL(newWebhook.URL)

	if newWebhook.Name == "" || newWebhook.URL == "" {
		return apperr.ErrWebhookRequired
	}
	if err := validateWebhookURL(newWebhook.URL); err != nil {
		return err
	}

	webhook := &webhookModel.Webhook{
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		return settingService.webhookRepository.UpdateWebhookByID(ctx, id, webhook)
	})
}

func (settingService *SettingService) CreateWebhook(
	ctx context.Context,
	newWebhook *model.WebhookDto,
) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	newWebhook.URL = urlUtil.TrimURL(newWebhook.URL)

	if newWebhook.Name == "" || newWebhook.URL == "" {
		return apperr.ErrWebhookRequired
	}
	if err := validateWebhookURL(newWebhook.URL); err != nil {
		return err
	}

	webhook := &webhookModel.Webhook{
		Name:     newWebhook.Name,
		URL:      newWebhook.URL,
		Secret:   newWebhook.Secret,
		IsActive: newWebhook.IsActive,
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		return settingService.webhookRepository.CreateWebhook(ctx, webhook)
	})
}

func (settingService *SettingService) TestWebhook(ctx context.Context, id string) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	webhook, err := settingService.webhookRepository.GetWebhookByID(ctx, id)
	if err != nil {
		return err
	}
	if err := validateWebhookURL(webhook.URL); err != nil {
		return err
	}

	triggerAt := time.Now().UTC().Unix()
	sendErr := settingService.webhookSender.SendTest(webhook)
	status := "success"
	if sendErr != nil {
		status = "failed"
	}
	_ = settingService.webhookRepository.UpdateWebhookDeliveryStatus(ctx, webhook.ID, status, triggerAt)
	return sendErr
}

func validateWebhookURL(rawURL string) error {
	if err := egress.Validate(rawURL); err != nil {
		return apperr.ErrWebhookURL
	}
	return nil
}
