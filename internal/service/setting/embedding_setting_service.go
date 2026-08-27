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

func (settingService *SettingService) GetEmbeddingSetting(
	ctx context.Context,
) (model.EmbeddingSetting, error) {
	return coreSetting.Get(ctx, settingService.durableKV, coreSetting.Embedding)
}

func (settingService *SettingService) UpdateEmbeddingSetting(
	ctx context.Context,
	dto model.EmbeddingSettingDto,
) error {
	userid := viewer.MustFromContext(ctx).UserID()
	user, err := settingService.commonService.CommonGetUserByUserId(ctx, userid)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		return errors.New(commonModel.NO_PERMISSION_DENIED)
	}

	setting := model.EmbeddingSetting{
		Enable:    dto.Enable,
		Model:     strings.TrimSpace(dto.Model),
		ApiKey:    strings.TrimSpace(dto.ApiKey),
		BaseURL:   strings.TrimSpace(dto.BaseURL),
		Dim:       dto.Dim,
		BatchSize: dto.BatchSize,
	}

	return coreSetting.Set(ctx, settingService.durableKV, coreSetting.Embedding, setting)
}
