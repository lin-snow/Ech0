// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/authz"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
	fileService "github.com/lin-snow/ech0/internal/service/file"
	coreSetting "github.com/lin-snow/ech0/internal/setting"
	storage "github.com/lin-snow/ech0/internal/storage"
	urlUtil "github.com/lin-snow/ech0/internal/util/url"
	"github.com/lin-snow/ech0/pkg/viewer"
)

const s3TestTimeout = 15 * time.Second

func (settingService *SettingService) GetS3Setting(ctx context.Context, setting *model.S3Setting) error {
	userid := viewer.MustFromContext(ctx).UserID()
	v, err := coreSetting.S3.Get(ctx, settingService.durableKV)
	if err != nil {
		return err
	}
	*setting = v

	if userid == "" {
		maskS3Secrets(setting)
		return nil
	}
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
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
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	oldSetting, oldErr := coreSetting.S3.Get(ctx, settingService.durableKV)
	var appliedSetting *model.S3Setting

	if err := settingService.transactor.Run(ctx, func(ctx context.Context) error {
		s3Setting := normalizeS3SettingDto(newSetting)
		if err := coreSetting.S3.Set(ctx, settingService.durableKV, s3Setting); err != nil {
			return err
		}

		appliedSetting = &s3Setting
		return nil
	}); err != nil {
		return err
	}

	if appliedSetting != nil {
		if err := settingService.fileService.ApplyS3(
			context.Background(), s3StorageConfig(*appliedSetting), appliedSetting.Enable,
		); err != nil {
			_ = settingService.transactor.Run(context.Background(), func(ctx context.Context) error {
				if oldErr != nil {
					return coreSetting.S3.Unset(ctx, settingService.durableKV)
				}
				return coreSetting.S3.Set(ctx, settingService.durableKV, oldSetting)
			})
			_ = settingService.fileService.ReloadStorages(context.Background())
			return err
		}
	}

	return nil
}

func (settingService *SettingService) TestS3Connection(
	ctx context.Context,
	newSetting *model.S3SettingDto,
) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, s3TestTimeout)
	defer cancel()
	return settingService.fileService.TestStorage(
		ctx,
		storage.Config{ID: fileService.S3StorageID, Driver: storage.DriverS3, Config: s3StorageConfig(normalizeS3SettingDto(newSetting))},
	)
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

// s3StorageConfig maps the panel's settings onto a store configuration. It is
// the single translation point between the settings shape and the byte layer.
func s3StorageConfig(setting model.S3Setting) map[string]string {
	cfg := map[string]string{
		"endpoint":       setting.Endpoint,
		"access_key":     setting.AccessKey,
		"secret_key":     setting.SecretKey,
		"bucket":         setting.BucketName,
		"region":         setting.Region,
		"use_ssl":        strconv.FormatBool(setting.UseSSL),
		"use_path_style": strconv.FormatBool(setting.UsePathStyle),
		"path_prefix":    setting.PathPrefix,
	}
	if setting.Provider != "" {
		cfg["provider"] = strings.ToLower(setting.Provider)
	}
	if setting.CDNURL != "" {
		cfg["cdn_url"] = setting.CDNURL
	}
	return cfg
}
