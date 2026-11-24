// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"strings"

	"github.com/lin-snow/ech0/internal/authz"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) GetPasskeySetting(
	ctx context.Context,
	setting *model.PasskeySetting,
) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	v, err := coreSetting.Passkey.Get(ctx, settingService.durableKV)
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
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	passkeySetting := model.PasskeySetting{
		WebAuthnRPID:           strings.TrimSpace(newSetting.WebAuthnRPID),
		WebAuthnAllowedOrigins: sanitizeURLList(newSetting.WebAuthnAllowedOrigins),
	}
	return coreSetting.Passkey.Set(ctx, settingService.durableKV, passkeySetting)
}

func (settingService *SettingService) GetPasskeyStatus(status *model.PasskeyStatus) error {
	v, err := coreSetting.Passkey.Get(context.Background(), settingService.durableKV)
	if err != nil {
		return err
	}
	status.PasskeyReady = strings.TrimSpace(v.WebAuthnRPID) != "" &&
		len(v.WebAuthnAllowedOrigins) > 0
	return nil
}
