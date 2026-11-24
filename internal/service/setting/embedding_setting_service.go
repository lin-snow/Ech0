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

func (settingService *SettingService) GetEmbeddingSetting(
	ctx context.Context,
) (model.EmbeddingSetting, error) {
	return coreSetting.Embedding.Get(ctx, settingService.durableKV)
}

func (settingService *SettingService) UpdateEmbeddingSetting(
	ctx context.Context,
	dto model.EmbeddingSettingDto,
) error {
	if err := authz.Decide(viewer.MustFromContext(ctx), authz.SettingWrite, authz.Instance()); err != nil {
		return err
	}

	setting := model.EmbeddingSetting{
		Enable:    dto.Enable,
		Model:     strings.TrimSpace(dto.Model),
		ApiKey:    strings.TrimSpace(dto.ApiKey),
		BaseURL:   strings.TrimSpace(dto.BaseURL),
		Dim:       dto.Dim,
		BatchSize: dto.BatchSize,
	}

	return coreSetting.Embedding.Set(ctx, settingService.durableKV, setting)
}
