// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"strings"

	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) GetSetting(setting *model.SystemSetting) error {
	v, err := coreSetting.Get(context.Background(), settingService.durableKV, coreSetting.System)
	if err != nil {
		return err
	}
	*setting = v
	return nil
}

func (settingService *SettingService) BootstrapDefaultLocale(
	ctx context.Context,
	locale string,
) error {
	resolved := strings.TrimSpace(locale)
	if resolved == "" {
		return nil
	}
	resolved = i18nUtil.ResolveLocale(resolved)
	if resolved == "" || resolved == string(commonModel.DefaultLocale) {
		return nil
	}

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		current, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.System)
		if err != nil {
			return err
		}
		if i18nUtil.ResolveLocale(current.DefaultLocale) != string(commonModel.DefaultLocale) {
			return nil
		}
		current.DefaultLocale = resolved
		return coreSetting.Set(ctx, settingService.durableKV, coreSetting.System, current)
	})
}

func (settingService *SettingService) UpdateSetting(
	ctx context.Context,
	newSetting *model.SystemSettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	serverLogoChanged := false
	if newSetting != nil {
		var current model.SystemSetting
		if err := settingService.GetSetting(&current); err == nil {
			serverLogoChanged = strings.TrimSpace(current.ServerLogo) != strings.TrimSpace(newSetting.ServerLogo)
		}
	}
	if err := settingService.transactor.Run(ctx, func(ctx context.Context) error {
		user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
		if err != nil {
			return err
		}
		if !user.IsAdmin {
			return errors.New(commonModel.NO_PERMISSION_DENIED)
		}

		var setting model.SystemSetting
		setting.SiteTitle = newSetting.SiteTitle
		setting.ServerLogo = newSetting.ServerLogo
		setting.ServerName = newSetting.ServerName
		setting.ServerURL = urlUtil.TrimURL(newSetting.ServerURL)
		setting.AllowRegister = newSetting.AllowRegister
		setting.DefaultLocale = i18nUtil.ResolveLocale(newSetting.DefaultLocale)
		setting.ICPNumber = newSetting.ICPNumber
		setting.FooterContent = newSetting.FooterContent
		setting.FooterLink = urlUtil.TrimURL(newSetting.FooterLink)
		setting.MetingAPI = urlUtil.TrimURL(newSetting.MetingAPI)
		setting.CustomCSS = newSetting.CustomCSS
		setting.CustomJS = newSetting.CustomJS

		if err := coreSetting.Set(ctx, settingService.durableKV, coreSetting.System, setting); err != nil {
			return err
		}

		if err := settingService.durableKV.Set(ctx, commonModel.ServerURLKey, setting.ServerURL); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	if serverLogoChanged && strings.TrimSpace(newSetting.ServerLogoFileID) != "" {
		if err := settingService.fileService.ConfirmTempFiles(ctx, []string{newSetting.ServerLogoFileID}); err != nil {
			logUtil.GetLogger().Warn("confirm temp server logo file failed", logUtil.Err(err))
		}
	}
	return nil
}
