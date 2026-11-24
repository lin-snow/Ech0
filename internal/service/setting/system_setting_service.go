// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"strings"

	"github.com/lin-snow/ech0/internal/authz"
	i18nUtil "github.com/lin-snow/ech0/internal/i18n"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/viewer"
)

func (settingService *SettingService) GetSetting(setting *model.SystemSetting) error {
	v, err := coreSetting.System.Get(context.Background(), settingService.durableKV)
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
		current, err := coreSetting.System.Get(ctx, settingService.durableKV)
		if err != nil {
			return err
		}
		if i18nUtil.ResolveLocale(current.DefaultLocale) != string(commonModel.DefaultLocale) {
			return nil
		}
		current.DefaultLocale = resolved
		return coreSetting.System.Set(ctx, settingService.durableKV, current)
	})
}

// BootstrapSiteTitle applies the site title captured during first-run setup.
// It is intentionally one-shot-ish: an already customised title wins, so the
// value can only replace the pristine default.
func (settingService *SettingService) BootstrapSiteTitle(
	ctx context.Context,
	title string,
) error {
	resolved := strings.TrimSpace(title)
	if resolved == "" {
		return nil
	}

	pristine := strings.TrimSpace(coreSetting.System.DefaultValue().SiteTitle)

	return settingService.transactor.Run(ctx, func(ctx context.Context) error {
		current, err := coreSetting.System.Get(ctx, settingService.durableKV)
		if err != nil {
			return err
		}
		if trimmed := strings.TrimSpace(current.SiteTitle); trimmed != "" && trimmed != pristine {
			return nil
		}
		current.SiteTitle = resolved
		return coreSetting.System.Set(ctx, settingService.durableKV, current)
	})
}

func (settingService *SettingService) UpdateSetting(
	ctx context.Context,
	newSetting *model.SystemSettingDto,
) error {
	newLogoFileID := ""
	hadLogo := false
	serverLogoChanged := false
	if newSetting != nil {
		newLogoFileID = strings.TrimSpace(newSetting.ServerLogoFileID)
		var current model.SystemSetting
		if err := settingService.GetSetting(&current); err == nil {
			hadLogo = strings.TrimSpace(current.ServerLogo) != ""
			serverLogoChanged = strings.TrimSpace(current.ServerLogo) != strings.TrimSpace(newSetting.ServerLogo) ||
				newLogoFileID != ""
		}
	}
	// The instance logo is a link of the instance owner, not a table of its own.
	const instanceOwnerID = "instance"
	if err := settingService.transactor.Run(ctx, func(ctx context.Context) error {
		if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
			return err
		}

		serverLogo := newSetting.ServerLogo
		if newLogoFileID != "" {
			// Get authorizes: an instance logo is public, but a private upload
			// is not the caller's to promote.
			logoDTO, err := settingService.fileService.Get(ctx, newLogoFileID)
			if err != nil {
				return err
			}
			serverLogo = logoDTO.URL
			if err := settingService.fileService.Replace(
				ctx, []string{newLogoFileID},
				fileModel.OwnerKindInstance, instanceOwnerID, fileModel.RoleLogo,
			); err != nil {
				return err
			}
		} else if serverLogoChanged && hadLogo {
			// An external logo URL replaced a stored one: the slot empties.
			if err := settingService.fileService.ClearOwner(
				ctx, fileModel.OwnerKindInstance, instanceOwnerID,
			); err != nil {
				return err
			}
		}

		var setting model.SystemSetting
		setting.SiteTitle = newSetting.SiteTitle
		setting.ServerLogo = serverLogo
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
		setting.PrivacyPolicy = newSetting.PrivacyPolicy
		setting.ShowPrivacyFooter = newSetting.ShowPrivacyFooter
		setting.EnableAgentDiscovery = newSetting.EnableAgentDiscovery

		if err := coreSetting.System.Set(ctx, settingService.durableKV, setting); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}
