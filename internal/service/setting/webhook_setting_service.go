// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	"github.com/lin-snow/ech0/internal/util/egress"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) GetAllWebhooks(ctx context.Context) ([]webhookModel.Webhook, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	webhooks, err := settingService.webhookRepository.GetAllWebhooks(ctx)
	if err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (settingService *SettingService) DeleteWebhook(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
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
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	newWebhook.URL = urlUtil.TrimURL(newWebhook.URL)

	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
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
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	newWebhook.URL = urlUtil.TrimURL(newWebhook.URL)

	if newWebhook.Name == "" || newWebhook.URL == "" {
		return errors.New(commonModel.WEBHOOK_NAME_OR_URL_CANNOT_BE_EMPTY)
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
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
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
		return errors.New(commonModel.INVALID_WEBHOOK_URL)
	}
	return nil
}
