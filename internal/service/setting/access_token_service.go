// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) ListAccessTokens(
	ctx context.Context,
) ([]model.AccessTokenSetting, error) {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return nil, err
	}
	if !user.IsAdmin {
		return nil, errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	tokens, err := settingService.settingRepository.ListAccessTokens(ctx, user.ID)
	if err != nil {
		return []model.AccessTokenSetting{}, nil
	}

	var validTokens []model.AccessTokenSetting
	currentTime := time.Now().UTC().Unix()

	for _, token := range tokens {
		if token.Expiry == nil || *token.Expiry > currentTime {
			validTokens = append(validTokens, token)
		} else {
			_ = settingService.transactor.Run(ctx, func(txCtx context.Context) error {
				return settingService.settingRepository.DeleteAccessTokenByID(txCtx, token.ID)
			})
		}
	}

	return validTokens, nil
}

// CreateAccessToken is the settings-panel surface; minting itself lives in the
// access capability so the device flow and OAuth server share one policy.
func (settingService *SettingService) CreateAccessToken(
	ctx context.Context,
	newToken *model.AccessTokenSettingDto,
) (string, error) {
	return settingService.access.IssueForViewer(ctx, newToken)
}

func (settingService *SettingService) DeleteAccessToken(ctx context.Context, id string) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	if settingService.tokenRevoker != nil {
		token, getErr := settingService.settingRepository.GetAccessTokenByID(ctx, id)
		if getErr == nil && token.JTI != "" {
			ttl := remainingTTLForRevoke(token.Expiry)
			settingService.tokenRevoker.RevokeToken(token.JTI, ttl)
		}
	}

	return settingService.transactor.Run(ctx, func(txCtx context.Context) error {
		return settingService.settingRepository.DeleteAccessTokenByID(txCtx, id)
	})
}

func remainingTTLForRevoke(expiryUnix *int64) time.Duration {
	const neverFallback = 100 * 365 * 24 * time.Hour
	if expiryUnix == nil {
		return neverFallback
	}
	remaining := time.Until(time.Unix(*expiryUnix, 0).UTC())
	if remaining <= 0 {
		return time.Hour
	}
	return remaining
}
