// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"strings"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) GetPasskeySetting(
	ctx context.Context,
	setting *model.PasskeySetting,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	v, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.Passkey)
	if err != nil {
		return err
	}
	*setting = v
	return nil
}

func (settingService *SettingService) UpdatePasskeySetting(
	ctx context.Context,
	newSetting *model.PasskeySettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	passkeySetting := model.PasskeySetting{
		WebAuthnRPID:           strings.TrimSpace(newSetting.WebAuthnRPID),
		WebAuthnAllowedOrigins: sanitizeURLList(newSetting.WebAuthnAllowedOrigins),
	}
	return coreSetting.Set(ctx, settingService.durableKV, coreSetting.Passkey, passkeySetting)
}

func (settingService *SettingService) GetPasskeyStatus(status *model.PasskeyStatus) error {
	v, err := coreSetting.Get(context.Background(), settingService.durableKV, coreSetting.Passkey)
	if err != nil {
		return err
	}
	status.PasskeyReady = strings.TrimSpace(v.WebAuthnRPID) != "" &&
		len(v.WebAuthnAllowedOrigins) > 0
	return nil
}
