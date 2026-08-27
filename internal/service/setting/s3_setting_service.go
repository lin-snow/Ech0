// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const s3TestTimeout = 15 * time.Second

func (settingService *SettingService) GetS3Setting(ctx context.Context, setting *model.S3Setting) error {
	userid := viewer.MustFromContext(ctx).UserID()
	v, err := coreSetting.Get(ctx, settingService.durableKV, coreSetting.S3)
	if err != nil {
		return err
	}
	*setting = v

	if userid == "" {
		maskS3Secrets(setting)
		return nil
	}
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		maskS3Secrets(setting)
	}
	return nil
}

func maskS3Secrets(setting *model.S3Setting) {
	setting.AccessKey = "******"
	setting.SecretKey = "******"
	setting.BucketName = "******"
	setting.Endpoint = "******"
}

func (settingService *SettingService) UpdateS3Setting(
	ctx context.Context,
	newSetting *model.S3SettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	oldRaw, _ := settingService.durableKV.Get(ctx, commonModel.S3SettingKey)
	var appliedSetting *model.S3Setting

	err = settingService.transactor.Run(ctx, func(ctx context.Context) error {
		s3Setting := normalizeS3SettingDto(newSetting)
		if err := coreSetting.Set(ctx, settingService.durableKV, coreSetting.S3, s3Setting); err != nil {
			return err
		}

		appliedSetting = &s3Setting
		return nil
	})
	if err != nil {
		return err
	}

	if settingService.storageManager != nil && appliedSetting != nil {
		if err := settingService.storageManager.ApplyS3Setting(*appliedSetting); err != nil {
			_ = settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
				if strings.TrimSpace(oldRaw) == "" {
					return settingService.durableKV.Delete(ctx, commonModel.S3SettingKey)
				}
				return settingService.durableKV.Set(ctx, commonModel.S3SettingKey, oldRaw)
			})
			_ = settingService.storageManager.ReloadFromConfigAndDB(context.Background())
			return err
		}
	}

	return nil
}

func (settingService *SettingService) TestS3Connection(
	ctx context.Context,
	newSetting *model.S3SettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	if settingService.storageManager == nil {
		return errors.New("存储管理器不可用")
	}

	ctx, cancel := context.WithTimeout(ctx, s3TestTimeout)
	defer cancel()
	return settingService.storageManager.TestS3Connection(ctx, normalizeS3SettingDto(newSetting))
}

func normalizeS3SettingDto(newSetting *model.S3SettingDto) model.S3Setting {
	useSSL := newSetting.UseSSL
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "https://") {
		useSSL = true
	} else if strings.HasPrefix(strings.ToLower(strings.TrimSpace(newSetting.Endpoint)), "http://") {
		useSSL = false
	}

	endpoint := strings.TrimSpace(newSetting.Endpoint)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	cdnURL := strings.TrimSpace(newSetting.CDNURL)
	if cdnURL != "" {
		cdnURL = strings.TrimRight(cdnURL, "/")
	}

	s3Setting := model.S3Setting{
		Enable:       newSetting.Enable,
		Provider:     newSetting.Provider,
		Endpoint:     urlUtil.TrimURL(endpoint),
		AccessKey:    newSetting.AccessKey,
		SecretKey:    newSetting.SecretKey,
		BucketName:   newSetting.BucketName,
		Region:       strings.TrimSpace(newSetting.Region),
		UseSSL:       useSSL,
		CDNURL:       cdnURL,
		PathPrefix:   urlUtil.TrimURL(newSetting.PathPrefix),
		PublicRead:   newSetting.PublicRead,
		UsePathStyle: newSetting.UsePathStyle,
	}

	switch s3Setting.Provider {
	case string(commonModel.R2):
		if s3Setting.Region == "" {
			s3Setting.Region = "auto"
		}
		s3Setting.UseSSL = true
	case string(commonModel.AWS):
		if s3Setting.Region == "" {
			s3Setting.Region = "us-east-1"
		}
	case string(commonModel.MINIO):
		if s3Setting.Region == "" {
			s3Setting.Region = "us-east-1"
		}
	case string(commonModel.OTHER):
		if s3Setting.Region == "" {
			s3Setting.Region = "auto"
		}
	default:
	}

	if s3Setting.Provider != string(commonModel.OTHER) {
		s3Setting.UsePathStyle = false
	}

	return s3Setting
}
